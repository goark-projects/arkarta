package resource

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"goark.dev/arkarta/servlet"
)

// DefaultServlet 负责按 Servlet default mapping 语义写出静态资源。
type DefaultServlet struct {
	provider     Provider
	welcomeFiles []string
}

// DefaultServletOption 定制 default servlet 行为。
type DefaultServletOption func(*DefaultServlet) error

// NewDefaultServlet 创建静态资源 default servlet。
func NewDefaultServlet(provider Provider, options ...DefaultServletOption) (*DefaultServlet, error) {
	if provider == nil {
		return nil, ErrNilProvider
	}
	servlet := &DefaultServlet{
		provider:     provider,
		welcomeFiles: cloneStrings(defaultWelcomeFiles),
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(servlet); err != nil {
			return nil, err
		}
	}
	return servlet, nil
}

// Init 接收容器初始化回调；静态资源 Servlet 无需额外运行时状态。
func (s *DefaultServlet) Init(ctx context.Context, _ servlet.ServletConfig) error {
	return ctx.Err()
}

// Serve 按请求路径写出静态资源。
func (s *DefaultServlet) Serve(ctx context.Context, req *servlet.Request, res servlet.Response) error {
	if req == nil {
		return servlet.NewHTTPError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil)
	}
	if res == nil {
		return servlet.ErrNilResponse
	}
	if req.Method() != http.MethodGet && req.Method() != http.MethodHead {
		res.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		return servlet.NewHTTPError(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), nil)
	}
	item, err := s.openResource(ctx, resourcePath(req))
	if err != nil {
		return mapResourceError(err)
	}
	defer item.Body().Close()

	writeResourceHeaders(res, item)
	if isNotModified(req, item) {
		res.SetStatus(http.StatusNotModified)
		return nil
	}
	if handled, err := serveRange(req, res, item); handled || err != nil {
		return err
	}
	res.SetStatus(http.StatusOK)
	if req.Method() == http.MethodHead {
		return nil
	}
	_, err = io.Copy(res.BodyWriter(), item.Body())
	return err
}

func resourcePath(req *servlet.Request) string {
	if pathInfo := req.PathInfo(); pathInfo != "" {
		return pathInfo
	}
	return req.Path()
}

func (s *DefaultServlet) openResource(ctx context.Context, path string) (Resource, error) {
	item, err := s.provider.Open(ctx, path)
	if errors.Is(err, ErrDirectory) {
		return s.openWelcome(ctx, path)
	}
	return item, err
}

// Destroy 接收容器销毁回调；静态资源 Servlet 无需释放额外资源。
func (s *DefaultServlet) Destroy(ctx context.Context) error {
	return ctx.Err()
}

func writeResourceHeaders(res servlet.Response, item Resource) {
	res.Header().Set("Content-Type", item.ContentType())
	res.Header().Set("Accept-Ranges", "bytes")
	if item.Size() >= 0 {
		_ = servlet.SetContentLength(res, item.Size())
	}
	if !item.ModTime().IsZero() {
		res.Header().Set("Last-Modified", item.ModTime().UTC().Format(http.TimeFormat))
	}
	if item.ETag() != "" {
		res.Header().Set("ETag", item.ETag())
	}
}

func serveRange(req *servlet.Request, res servlet.Response, item Resource) (bool, error) {
	if !ifRangeAllows(req, item) {
		return false, nil
	}
	targets, ok, invalid := parseRanges(req.Header().Get("Range"), item.Size())
	if invalid {
		res.Header().Set("Content-Range", unsatisfiedRange(item.Size()))
		res.SetStatus(rangeNotSatisfiable())
		return true, nil
	}
	if !ok {
		return false, nil
	}
	if len(targets) > 1 {
		return serveMultipleRanges(req, res, item, targets)
	}
	target := targets[0]
	res.Header().Set("Content-Range", target.contentRange(item.Size()))
	_ = servlet.SetContentLength(res, target.length())
	res.SetStatus(http.StatusPartialContent)
	if req.Method() == http.MethodHead {
		return true, nil
	}
	if _, err := io.CopyN(io.Discard, item.Body(), target.start); err != nil {
		return true, err
	}
	_, err := io.CopyN(res.BodyWriter(), item.Body(), target.length())
	return true, err
}

func serveMultipleRanges(req *servlet.Request, res servlet.Response, item Resource, targets []byteRange) (bool, error) {
	const boundary = "arkarta-resource-boundary"
	res.Header().Set("Content-Type", "multipart/byteranges; boundary="+boundary)
	res.Header().Delete("Content-Length")
	res.Header().Delete("Content-Range")
	res.SetStatus(http.StatusPartialContent)
	if req.Method() == http.MethodHead {
		return true, nil
	}
	data, err := io.ReadAll(item.Body())
	if err != nil {
		return true, err
	}
	for _, target := range targets {
		start := int(target.start)
		end := int(target.end) + 1
		if start < 0 || end > len(data) || start > end {
			return true, io.ErrUnexpectedEOF
		}
		if _, err := fmt.Fprintf(res.BodyWriter(),
			"--%s\r\nContent-Type: %s\r\nContent-Range: %s\r\n\r\n",
			boundary,
			item.ContentType(),
			target.contentRange(item.Size()),
		); err != nil {
			return true, err
		}
		if _, err := io.CopyN(res.BodyWriter(), bytes.NewReader(data[start:end]), target.length()); err != nil {
			return true, err
		}
		if _, err := io.WriteString(res.BodyWriter(), "\r\n"); err != nil {
			return true, err
		}
	}
	_, err = fmt.Fprintf(res.BodyWriter(), "--%s--\r\n", boundary)
	return true, err
}

func ifRangeAllows(req *servlet.Request, item Resource) bool {
	value := strings.TrimSpace(req.Header().Get("If-Range"))
	if value == "" {
		return true
	}
	if strings.HasPrefix(value, "W/") {
		return false
	}
	if strings.HasPrefix(value, `"`) {
		return item.ETag() != "" && !strings.HasPrefix(item.ETag(), "W/") && value == item.ETag()
	}
	if item.ModTime().IsZero() {
		return false
	}
	date, err := http.ParseTime(value)
	if err != nil {
		return false
	}
	return !item.ModTime().UTC().Truncate(time.Second).After(date)
}

func mapResourceError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidPath), errors.Is(err, ErrNotFound), errors.Is(err, ErrDirectory):
		return servlet.NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound), err)
	default:
		return err
	}
}

func isNotModified(req *servlet.Request, item Resource) bool {
	if matchesETag(req.Header().Values("If-None-Match"), item.ETag()) {
		return true
	}
	if item.ModTime().IsZero() {
		return false
	}
	value := req.Header().Get("If-Modified-Since")
	if value == "" {
		return false
	}
	since, err := http.ParseTime(value)
	if err != nil {
		return false
	}
	return !item.ModTime().UTC().Truncate(time.Second).After(since)
}

func matchesETag(values []string, etag string) bool {
	if etag == "" {
		return false
	}
	for _, value := range values {
		for _, candidate := range strings.Split(value, ",") {
			candidate = strings.TrimSpace(candidate)
			if candidate == "*" || candidate == etag {
				return true
			}
		}
	}
	return false
}
