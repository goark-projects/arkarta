package servlet

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
)

// ErrChainAlreadyAdvanced 表示同一个过滤器链节点已经向后推进过。
var ErrChainAlreadyAdvanced = errors.New("arkarta/servlet: filter chain already advanced")

// Filter 是请求进入目标处理器前后的横切处理器。
type Filter interface {
	Filter(ctx context.Context, req *Request, res Response, chain Chain) error
}

// FilterFunc 将普通函数适配为 Filter。
type FilterFunc func(ctx context.Context, req *Request, res Response, chain Chain) error

// Filter 执行函数式过滤器。
func (f FilterFunc) Filter(ctx context.Context, req *Request, res Response, chain Chain) error {
	return f(ctx, req, res, chain)
}

// Chain 表示当前过滤器之后的剩余链路。
type Chain interface {
	Next(ctx context.Context, req *Request, res Response) error
}

// ChainFilters 将过滤器和目标处理器组合为一个 Handler。
func ChainFilters(target Handler, filters ...Filter) Handler {
	chainFilters := make([]Filter, 0, len(filters))
	for _, filter := range filters {
		if filter != nil {
			chainFilters = append(chainFilters, filter)
		}
	}
	return HandlerFunc(func(ctx context.Context, req *Request, res Response) error {
		chain := &filterChain{
			target:  target,
			filters: chainFilters,
		}
		return chain.Next(ctx, req, res)
	})
}

type filterChain struct {
	target  Handler
	filters []Filter
	index   int

	mu       sync.Mutex
	advanced bool
}

func (c *filterChain) Next(ctx context.Context, req *Request, res Response) error {
	c.mu.Lock()
	if c.advanced {
		c.mu.Unlock()
		return ErrChainAlreadyAdvanced
	}
	c.advanced = true
	c.mu.Unlock()

	if c.index >= len(c.filters) {
		return c.target.Serve(ctx, req, res)
	}
	next := &filterChain{
		target:  c.target,
		filters: c.filters,
		index:   c.index + 1,
	}
	return c.filters[c.index].Filter(ctx, req, res, next)
}

// ManagedFilter 是带生命周期的容器托管过滤器。
type ManagedFilter interface {
	Filter
	Init(ctx context.Context, cfg FilterConfig) error
	Destroy(ctx context.Context) error
}

// FilterConfig 表示容器传递给 Filter 初始化阶段的只读配置。
type FilterConfig struct {
	name      string
	initParam map[string]string
	webApp    *WebApp
}

// NewFilterConfig 创建 Filter 初始化配置。
func NewFilterConfig(name string, webApp *WebApp, initParam map[string]string) FilterConfig {
	return FilterConfig{
		name:      name,
		webApp:    webApp,
		initParam: cloneStringMap(initParam),
	}
}

// Name 返回 Filter 名称。
func (c FilterConfig) Name() string {
	return c.name
}

// WebApp 返回所属 Web 应用上下文。
func (c FilterConfig) WebApp() *WebApp {
	return c.webApp
}

// InitParam 返回指定初始化参数。
func (c FilterConfig) InitParam(name string) (string, bool) {
	value, ok := c.initParam[name]
	return value, ok
}

// InitParams 返回初始化参数副本。
func (c FilterConfig) InitParams() map[string]string {
	return cloneStringMap(c.initParam)
}

// ChainFilterBindings 将带分发约束的 Filter 与目标处理器组合为 Handler。
func ChainFilterBindings(target Handler, bindings ...FilterBinding) Handler {
	return HandlerFunc(func(ctx context.Context, req *Request, res Response) error {
		filters := make([]Filter, 0, len(bindings))
		for _, binding := range bindings {
			if binding.Filter() != nil && binding.MatchesRequest(req) {
				filters = append(filters, binding.Filter())
			}
		}
		return ChainFilters(target, filters...).Serve(ctx, req, res)
	})
}

// ErrNilErrorPageRegistry 表示错误页注册表为空。
var ErrNilErrorPageRegistry = errors.New("arkarta/servlet: error page registry is nil")

// ErrErrorPageLoop 表示错误页处理过程中再次进入错误页分发。
var ErrErrorPageLoop = errors.New("arkarta/servlet: error page dispatch loop")

const attributeErrorPageActive = "arkarta.servlet.error.active"

// ErrorPageRegistry 保存状态码和错误类型到错误页处理器的映射。
type ErrorPageRegistry struct {
	mu             sync.RWMutex
	status         map[int]Handler
	errorTypes     []errorTypeMapping
	defaultHandler Handler
}

type errorTypeMapping struct {
	match   func(error) bool
	handler Handler
}

// NewErrorPageRegistry 创建错误页注册表。
func NewErrorPageRegistry() *ErrorPageRegistry {
	return &ErrorPageRegistry{
		status: make(map[int]Handler),
	}
}

// RegisterStatus 注册 HTTP 状态码错误页。
func (r *ErrorPageRegistry) RegisterStatus(statusCode int, handler Handler) error {
	if r == nil {
		return ErrNilErrorPageRegistry
	}
	if handler == nil {
		return ErrNilHandler
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status[normalizeStatus(statusCode)] = handler
	return nil
}

// RegisterDefault 注册默认错误页。
func (r *ErrorPageRegistry) RegisterDefault(handler Handler) error {
	if r == nil {
		return ErrNilErrorPageRegistry
	}
	if handler == nil {
		return ErrNilHandler
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultHandler = handler
	return nil
}

// Handle 尝试用已注册错误页处理错误。
func (r *ErrorPageRegistry) Handle(
	ctx context.Context,
	req *Request,
	res Response,
	statusCode int,
	cause error,
) (bool, error) {
	if r == nil {
		return false, nil
	}
	if active, _ := req.Attribute(attributeErrorPageActive); active == true {
		return false, ErrErrorPageLoop
	}
	if res != nil && res.Committed() {
		return false, ErrResponseCommitted
	}
	handler := r.match(statusCode, cause)
	if handler == nil {
		return false, nil
	}

	statusCode = normalizeStatus(statusCode)
	if res != nil {
		res.SetStatus(statusCode)
	}
	setErrorAttributes(req, statusCode, cause)
	req.SetAttribute(attributeErrorPageActive, true)
	defer req.SetAttribute(attributeErrorPageActive, nil)

	snapshot := req.dispatchSnapshot()
	req.applyDispatch(req.Path(), snapshot.queryString, DispatchError)
	defer req.restoreDispatch(snapshot)
	return true, handler.Serve(ctx, req, res)
}

func (r *ErrorPageRegistry) registerErrorType(match func(error) bool, handler Handler) error {
	if r == nil {
		return ErrNilErrorPageRegistry
	}
	if handler == nil {
		return ErrNilHandler
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errorTypes = append(r.errorTypes, errorTypeMapping{
		match:   match,
		handler: handler,
	})
	return nil
}

func (r *ErrorPageRegistry) match(statusCode int, cause error) Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if cause != nil {
		for i := len(r.errorTypes) - 1; i >= 0; i-- {
			mapping := r.errorTypes[i]
			if mapping.match(cause) {
				return mapping.handler
			}
		}
	}
	if handler := r.status[normalizeStatus(statusCode)]; handler != nil {
		return handler
	}
	return r.defaultHandler
}

func normalizeStatus(statusCode int) int {
	if statusCode < 100 || statusCode > 999 {
		return http.StatusInternalServerError
	}
	return statusCode
}

// RegisterErrorType 注册基于 errors.As 匹配的错误类型错误页。
func RegisterErrorType[T error](registry *ErrorPageRegistry, handler Handler) error {
	return registry.registerErrorType(func(err error) bool {
		var target T
		return errors.As(err, &target)
	}, handler)
}

type MappingType uint8

const (
	// MappingUnknown 表示请求尚未命中任何 Servlet 映射。
	MappingUnknown MappingType = iota
	// MappingDefault 表示默认映射 "/"。
	MappingDefault
	// MappingExact 表示精确路径映射。
	MappingExact
	// MappingPrefix 表示路径前缀映射。
	MappingPrefix
	// MappingExtension 表示扩展名映射。
	MappingExtension
)

// RequestMapping 描述一次请求命中的 Servlet 映射信息。
type RequestMapping struct {
	pattern     string
	mappingType MappingType
	servletPath string
	pathInfo    string
}

func newRequestMapping(
	pattern string,
	mappingType MappingType,
	servletPath, pathInfo string,
) RequestMapping {
	return RequestMapping{
		pattern: pattern, mappingType: mappingType,
		servletPath: servletPath, pathInfo: pathInfo,
	}
}

// Pattern 返回声明的 Servlet 映射模式。
func (m RequestMapping) Pattern() string { return m.pattern }

// Type 返回映射命中类型。
func (m RequestMapping) Type() MappingType { return m.mappingType }

// ServletPath 返回映射对应的 Servlet 路径。
func (m RequestMapping) ServletPath() string { return m.servletPath }

// PathInfo 返回 ServletPath 之后的剩余路径。
func (m RequestMapping) PathInfo() string { return m.pathInfo }

// WithRequestContextPath 设置请求所属 Web 应用的上下文路径。
func WithRequestContextPath(contextPath string) RequestOption {
	return func(req *Request) {
		req.contextPath = normalizeRequestContextPath(contextPath)
		req.path = stripRequestContextPath(req.path, req.contextPath)
		req.servletPath = ""
		req.pathInfo = ""
		req.mapping = RequestMapping{}
	}
}

func normalizeRequestContextPath(contextPath string) string {
	if contextPath == "" || contextPath == "/" {
		return ""
	}
	if !strings.HasPrefix(contextPath, "/") {
		contextPath = "/" + contextPath
	}
	return strings.TrimRight(contextPath, "/")
}

func stripRequestContextPath(path, contextPath string) string {
	if path == "" || contextPath == "" {
		return path
	}
	if path == contextPath {
		return "/"
	}
	if strings.HasPrefix(path, contextPath+"/") {
		return strings.TrimPrefix(path, contextPath)
	}
	return path
}
