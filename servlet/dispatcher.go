package servlet

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// ErrNilRouter 表示 RequestDispatcher 缺少路由器。
var ErrNilRouter = errors.New("arkarta/servlet: router is nil")

// ErrDispatcherTargetNotFound 表示分发目标不存在。
var ErrDispatcherTargetNotFound = errors.New("arkarta/servlet: dispatcher target not found")

// RequestDispatcher 执行服务端请求分发。
type RequestDispatcher interface {
	Forward(ctx context.Context, req *Request, res Response) error
	Include(ctx context.Context, req *Request, res Response) error
	Error(ctx context.Context, req *Request, res Response, statusCode int, cause error) error
}

// Dispatcher 是基于 Router 的 RequestDispatcher 实现。
type Dispatcher struct {
	router      *Router
	path        string
	queryString string
	hasQuery    bool
}

// NewRequestDispatcher 创建指定目标路径的请求分发器。
func NewRequestDispatcher(router *Router, path string) (*Dispatcher, error) {
	if router == nil {
		return nil, ErrNilRouter
	}
	targetPath, queryString, hasQuery, err := splitDispatcherPath(path)
	if err != nil {
		return nil, ErrInvalidMappingPattern
	}
	return &Dispatcher{router: router, path: targetPath, queryString: queryString, hasQuery: hasQuery}, nil
}

// Forward 在响应提交前转发请求。
func (d *Dispatcher) Forward(ctx context.Context, req *Request, res Response) error {
	if res != nil && res.Committed() {
		return ErrResponseCommitted
	}
	if res != nil {
		if err := res.Reset(); err != nil {
			return err
		}
	}
	setForwardAttributes(req)
	return d.dispatch(ctx, req, res, DispatchForward)
}

// Include 包含目标处理器输出，但隔离目标状态码和 Header。
func (d *Dispatcher) Include(ctx context.Context, req *Request, res Response) error {
	setIncludeAttributes(req)
	return d.dispatch(ctx, req, newIncludeResponse(res), DispatchInclude)
}

// Error 执行错误分发。
func (d *Dispatcher) Error(ctx context.Context, req *Request, res Response, statusCode int, cause error) error {
	if statusCode < 100 || statusCode > 999 {
		statusCode = http.StatusInternalServerError
	}
	if res != nil && !res.Committed() {
		res.SetStatus(statusCode)
	}
	setErrorAttributes(req, statusCode, cause)
	return d.dispatch(ctx, req, res, DispatchError)
}

func (d *Dispatcher) dispatch(ctx context.Context, req *Request, res Response, dispatchType DispatchType) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	match, ok := d.router.MatchRoute(d.path)
	if !ok {
		return ErrDispatcherTargetNotFound
	}

	snapshot := req.dispatchSnapshot()
	queryString := snapshot.queryString
	if d.hasQuery {
		queryString = d.queryString
	}
	req.applyDispatch(d.path, queryString, dispatchType)
	req.applyMapping(match.Mapping())
	defer req.restoreDispatch(snapshot)
	return match.Handler().Serve(ctx, req, res)
}

func splitDispatcherPath(path string) (string, string, bool, error) {
	if path == "" || path[0] != '/' {
		return "", "", false, ErrInvalidMappingPattern
	}
	targetPath, queryString, hasQuery := strings.Cut(path, "?")
	if targetPath == "" || strings.Contains(targetPath, "*") {
		return "", "", false, ErrInvalidMappingPattern
	}
	if queryString != "" {
		if _, err := url.ParseQuery(queryString); err != nil {
			return "", "", false, err
		}
	}
	return targetPath, queryString, hasQuery, nil
}

func setForwardAttributes(req *Request) {
	if _, exists := req.Attribute(AttributeForwardRequestURI); exists {
		return
	}
	setPathAttributeGroup(req, "forward")
}

func setIncludeAttributes(req *Request) {
	setPathAttributeGroup(req, "include")
}

func setPathAttributeGroup(req *Request, group string) {
	var requestURIKey, contextPathKey, servletPathKey, pathInfoKey, queryStringKey, mappingKey string
	switch group {
	case "forward":
		requestURIKey = AttributeForwardRequestURI
		contextPathKey = AttributeForwardContextPath
		servletPathKey = AttributeForwardServletPath
		pathInfoKey = AttributeForwardPathInfo
		queryStringKey = AttributeForwardQueryString
		mappingKey = AttributeForwardMapping
	case "include":
		requestURIKey = AttributeIncludeRequestURI
		contextPathKey = AttributeIncludeContextPath
		servletPathKey = AttributeIncludeServletPath
		pathInfoKey = AttributeIncludePathInfo
		queryStringKey = AttributeIncludeQueryString
		mappingKey = AttributeIncludeMapping
	default:
		return
	}
	req.SetAttribute(requestURIKey, req.Path())
	req.SetAttribute(contextPathKey, req.ContextPath())
	req.SetAttribute(servletPathKey, req.ServletPath())
	req.SetAttribute(pathInfoKey, req.PathInfo())
	req.SetAttribute(queryStringKey, req.QueryString())
	req.SetAttribute(mappingKey, req.Mapping())
}

func setErrorAttributes(req *Request, statusCode int, cause error) {
	req.SetAttribute(AttributeErrorStatusCode, statusCode)
	req.SetAttribute(AttributeErrorException, cause)
	req.SetAttribute(AttributeErrorExceptionType, errorTypeName(cause))
	req.SetAttribute(AttributeErrorMessage, errorMessage(statusCode, cause))
	req.SetAttribute(AttributeErrorRequestURI, req.Path())
	req.SetAttribute(AttributeErrorQueryString, req.QueryString())
	if servletName, ok := req.Attribute(AttributeServletName); ok {
		req.SetAttribute(AttributeErrorServletName, servletName)
	}
}

func errorTypeName(cause error) string {
	if cause == nil {
		return ""
	}
	return reflect.TypeOf(cause).String()
}

func errorMessage(statusCode int, cause error) string {
	var statusErr StatusError
	if errors.As(cause, &statusErr) && statusErr.PublicMessage() != "" {
		return statusErr.PublicMessage()
	}
	if text := http.StatusText(statusCode); text != "" {
		return text
	}
	if cause != nil {
		return cause.Error()
	}
	return ""
}
