package resource

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"goark.dev/arkarta/servlet"
)

// DefaultServlet 负责按 Servlet default mapping 语义写出静态资源。
type DefaultServlet struct {
	provider Provider
}

// NewDefaultServlet 创建静态资源 default servlet。
func NewDefaultServlet(provider Provider) (*DefaultServlet, error) {
	if provider == nil {
		return nil, ErrNilProvider
	}
	return &DefaultServlet{provider: provider}, nil
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
	item, err := s.provider.Open(ctx, resourcePath(req))
	if err != nil {
		return mapResourceError(err)
	}
	defer item.Body().Close()

	writeResourceHeaders(res, item)
	if isNotModified(req, item) {
		res.SetStatus(http.StatusNotModified)
		return nil
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

func writeResourceHeaders(res servlet.Response, item Resource) {
	res.Header().Set("Content-Type", item.ContentType())
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
