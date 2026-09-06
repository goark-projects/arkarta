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

// DispatchTypes 是 DispatchType 的紧凑位集合。
type DispatchTypes uint8

const (
	// DispatchOnRequest 匹配客户端原始请求。
	DispatchOnRequest DispatchTypes = 1 << iota
	// DispatchOnForward 匹配服务端 forward。
	DispatchOnForward
	// DispatchOnInclude 匹配服务端 include。
	DispatchOnInclude
	// DispatchOnError 匹配错误分发。
	DispatchOnError
	// DispatchOnAsync 匹配异步分发。
	DispatchOnAsync
)

const allDispatchTypes = DispatchOnRequest |
	DispatchOnForward |
	DispatchOnInclude |
	DispatchOnError |
	DispatchOnAsync

// NewDispatchTypes 从枚举值构造位集合；未传入时默认匹配 REQUEST。
func NewDispatchTypes(types ...DispatchType) (DispatchTypes, error) {
	if len(types) == 0 {
		return DispatchOnRequest, nil
	}
	var result DispatchTypes
	for _, item := range types {
		mask, ok := dispatchMask(item)
		if !ok {
			return 0, ErrInvalidDispatchTypes
		}
		result |= mask
	}
	return result, nil
}

// Contains 判断位集合是否包含指定 DispatchType。
func (d DispatchTypes) Contains(dispatchType DispatchType) bool {
	mask, ok := dispatchMask(dispatchType)
	return ok && d&mask != 0
}

// List 按 Arkarta Servlet 运行时顺序返回 DispatchType 切片。
func (d DispatchTypes) List() []DispatchType {
	d = NormalizeDispatchTypes(d)
	result := make([]DispatchType, 0, 5)
	for _, item := range []DispatchType{
		DispatchRequest,
		DispatchForward,
		DispatchInclude,
		DispatchError,
		DispatchAsync,
	} {
		if d.Contains(item) {
			result = append(result, item)
		}
	}
	return result
}

// NormalizeDispatchTypes 将空位集合归一化为 REQUEST。
func NormalizeDispatchTypes(dispatchers DispatchTypes) DispatchTypes {
	if dispatchers == 0 {
		return DispatchOnRequest
	}
	return dispatchers
}

// ValidateDispatchTypes 校验位集合是否只包含合法 DispatchType。
func ValidateDispatchTypes(dispatchers DispatchTypes) error {
	if NormalizeDispatchTypes(dispatchers)&^allDispatchTypes != 0 {
		return ErrInvalidDispatchTypes
	}
	return nil
}

func dispatchMask(dispatchType DispatchType) (DispatchTypes, bool) {
	switch dispatchType {
	case DispatchRequest:
		return DispatchOnRequest, true
	case DispatchForward:
		return DispatchOnForward, true
	case DispatchInclude:
		return DispatchOnInclude, true
	case DispatchError:
		return DispatchOnError, true
	case DispatchAsync:
		return DispatchOnAsync, true
	default:
		return 0, false
	}
}

const (
	// AttributeServletName 保存当前请求命中的 Servlet 名称。
	AttributeServletName = "arkarta.servlet.servlet_name"
	// AttributeForwardRequestURI 保存 forward 前的原始请求路径。
	AttributeForwardRequestURI = "arkarta.servlet.forward.request_uri"
	// AttributeForwardContextPath 保存 forward 前的上下文路径。
	AttributeForwardContextPath = "arkarta.servlet.forward.context_path"
	// AttributeForwardServletPath 保存 forward 前的 Servlet 路径。
	AttributeForwardServletPath = "arkarta.servlet.forward.servlet_path"
	// AttributeForwardPathInfo 保存 forward 前的 PathInfo。
	AttributeForwardPathInfo = "arkarta.servlet.forward.path_info"
	// AttributeForwardQueryString 保存 forward 前的查询串。
	AttributeForwardQueryString = "arkarta.servlet.forward.query_string"
	// AttributeForwardMapping 保存 forward 前的映射信息。
	AttributeForwardMapping = "arkarta.servlet.forward.mapping"
	// AttributeIncludeRequestURI 保存 include 前的原始请求路径。
	AttributeIncludeRequestURI = "arkarta.servlet.include.request_uri"
	// AttributeIncludeContextPath 保存 include 前的上下文路径。
	AttributeIncludeContextPath = "arkarta.servlet.include.context_path"
	// AttributeIncludeServletPath 保存 include 前的 Servlet 路径。
	AttributeIncludeServletPath = "arkarta.servlet.include.servlet_path"
	// AttributeIncludePathInfo 保存 include 前的 PathInfo。
	AttributeIncludePathInfo = "arkarta.servlet.include.path_info"
	// AttributeIncludeQueryString 保存 include 前的查询串。
	AttributeIncludeQueryString = "arkarta.servlet.include.query_string"
	// AttributeIncludeMapping 保存 include 前的映射信息。
	AttributeIncludeMapping = "arkarta.servlet.include.mapping"
	// AttributeErrorStatusCode 保存错误分发的 HTTP 状态码。
	AttributeErrorStatusCode = "arkarta.servlet.error.status_code"
	// AttributeErrorException 保存错误分发的错误对象。
	AttributeErrorException = "arkarta.servlet.error.exception"
	// AttributeErrorExceptionType 保存错误分发的错误类型名称。
	AttributeErrorExceptionType = "arkarta.servlet.error.exception_type"
	// AttributeErrorMessage 保存错误分发的公开错误消息。
	AttributeErrorMessage = "arkarta.servlet.error.message"
	// AttributeErrorRequestURI 保存错误发生时的请求路径。
	AttributeErrorRequestURI = "arkarta.servlet.error.request_uri"
	// AttributeErrorQueryString 保存错误发生时的查询串。
	AttributeErrorQueryString = "arkarta.servlet.error.query_string"
	// AttributeErrorServletName 保存错误发生时的 Servlet 名称。
	AttributeErrorServletName = "arkarta.servlet.error.servlet_name"
)

// DispatcherRegistry 基于 Router 提供路径和名称分发器。
