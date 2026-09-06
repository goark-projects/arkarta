package servlet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

// ChainFunc 将函数适配为 Chain。
type ChainFunc func(ctx context.Context, req *Request, res Response) error

// Next 执行后续处理链函数。
func (f ChainFunc) Next(ctx context.Context, req *Request, res Response) error {
	if f == nil {
		return nil
	}
	return f(ctx, req, res)
}

// ErrNilHandler 表示注册或组合处理器时传入了空处理器。
var ErrNilHandler = errors.New("arkarta/servlet: handler is nil")

// Handler 是应用处理请求的最小契约。
type Handler interface {
	Serve(ctx context.Context, req *Request, res Response) error
}

// HandlerFunc 将普通函数适配为 Handler。
type HandlerFunc func(ctx context.Context, req *Request, res Response) error

// Serve 调用底层函数处理请求。
func (f HandlerFunc) Serve(ctx context.Context, req *Request, res Response) error {
	return f(ctx, req, res)
}

// Servlet 是带生命周期的容器托管处理器。
type Servlet interface {
	Handler
	Init(ctx context.Context, cfg ServletConfig) error
	Destroy(ctx context.Context) error
}

// ServletConfig 表示容器传递给 Servlet 初始化阶段的只读配置。
type ServletConfig struct {
	name      string
	initParam map[string]string
	webApp    *WebApp
}

// NewServletConfig 创建 Servlet 初始化配置。
func NewServletConfig(name string, webApp *WebApp, initParam map[string]string) ServletConfig {
	return ServletConfig{
		name:      name,
		webApp:    webApp,
		initParam: cloneStringMap(initParam),
	}
}

// Name 返回 Servlet 名称。
func (c ServletConfig) Name() string {
	return c.name
}

// WebApp 返回所属 Web 应用上下文。
func (c ServletConfig) WebApp() *WebApp {
	return c.webApp
}

// InitParam 返回指定初始化参数。
func (c ServletConfig) InitParam(name string) (string, bool) {
	value, ok := c.initParam[name]
	return value, ok
}

// InitParams 返回初始化参数副本。
func (c ServletConfig) InitParams() map[string]string {
	return cloneStringMap(c.initParam)
}

// ErrNilFilter 表示过滤器为空。
var ErrNilFilter = errors.New("arkarta/servlet: filter is nil")

// ErrInvalidFilterConfig 表示 Filter 配置非法。
var ErrInvalidFilterConfig = errors.New("arkarta/servlet: invalid filter config")

// FilterBindingOption 定制 FilterBinding。
type FilterBindingOption func(*FilterBinding) error

// FilterBinding 表示一个 Filter 在运行时链路中的映射约束。
type FilterBinding struct {
	name          string
	filter        Filter
	urlPattern    string
	dispatchTypes DispatchTypes
	initParam     map[string]string
}

// NewFilterBinding 创建 Filter 运行时映射。
func NewFilterBinding(name string, filter Filter, options ...FilterBindingOption) (FilterBinding, error) {
	if isNilFilter(filter) {
		return FilterBinding{}, ErrNilFilter
	}
	binding := FilterBinding{
		name:          name,
		filter:        filter,
		dispatchTypes: DispatchOnRequest,
		initParam:     make(map[string]string),
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&binding); err != nil {
			return FilterBinding{}, err
		}
	}
	if err := ValidateDispatchTypes(binding.dispatchTypes); err != nil {
		return FilterBinding{}, err
	}
	binding.dispatchTypes = NormalizeDispatchTypes(binding.dispatchTypes)
	return binding, nil
}

// BindFilter 以指定分发类型快速创建 FilterBinding。
func BindFilter(filter Filter, dispatchTypes ...DispatchType) (FilterBinding, error) {
	dispatchers, err := NewDispatchTypes(dispatchTypes...)
	if err != nil {
		return FilterBinding{}, err
	}
	return NewFilterBinding("", filter, WithFilterDispatchTypes(dispatchers))
}

// WithFilterDispatchTypes 设置 Filter 匹配的分发类型集合。
func WithFilterDispatchTypes(dispatchTypes DispatchTypes) FilterBindingOption {
	return func(binding *FilterBinding) error {
		if err := ValidateDispatchTypes(dispatchTypes); err != nil {
			return err
		}
		binding.dispatchTypes = NormalizeDispatchTypes(dispatchTypes)
		return nil
	}
}

// WithFilterInitParam 设置 Filter 初始化参数。
func WithFilterInitParam(name, value string) FilterBindingOption {
	return func(binding *FilterBinding) error {
		if name == "" {
			return ErrInvalidFilterConfig
		}
		binding.initParam[name] = value
		return nil
	}
}

// WithFilterInitParams 批量设置 Filter 初始化参数。
func WithFilterInitParams(params map[string]string) FilterBindingOption {
	return func(binding *FilterBinding) error {
		for name, value := range params {
			if name == "" {
				return ErrInvalidFilterConfig
			}
			binding.initParam[name] = value
		}
		return nil
	}
}

// WithFilterURLPattern 设置 Filter 的 URL 模式约束。
func WithFilterURLPattern(pattern string) FilterBindingOption {
	return func(binding *FilterBinding) error {
		if pattern == "" {
			binding.urlPattern = ""
			return nil
		}
		if _, _, err := parseMappingPattern(pattern); err != nil {
			return err
		}
		binding.urlPattern = pattern
		return nil
	}
}

// Name 返回 Filter 名称。
func (b FilterBinding) Name() string {
	return b.name
}

// Filter 返回 Filter 实例。
func (b FilterBinding) Filter() Filter {
	return b.filter
}

// URLPattern 返回 Filter 的 URL 模式约束；空字符串表示不限制请求路径。
func (b FilterBinding) URLPattern() string {
	return b.urlPattern
}

// DispatchTypes 返回匹配的分发类型集合。
func (b FilterBinding) DispatchTypes() DispatchTypes {
	return b.dispatchTypes
}

// InitParams 返回初始化参数副本。
func (b FilterBinding) InitParams() map[string]string {
	return cloneStringMap(b.initParam)
}

// Matches 判断当前分发类型是否应该执行该 Filter。
func (b FilterBinding) Matches(dispatchType DispatchType) bool {
	return b.dispatchTypes.Contains(dispatchType)
}

// MatchesRequest 判断当前请求是否应该执行该 Filter。
func (b FilterBinding) MatchesRequest(req *Request) bool {
	dispatchType := DispatchRequest
	path := "/"
	if req != nil {
		dispatchType = req.DispatchType()
		path = req.Path()
	}
	if !b.Matches(dispatchType) {
		return false
	}
	return matchFilterURLPattern(path, b.urlPattern)
}

func matchFilterURLPattern(path, pattern string) bool {
	if pattern == "" {
		return true
	}
	kind, value, err := parseMappingPattern(pattern)
	if err != nil {
		return false
	}
	if path == "" {
		path = "/"
	}
	switch kind {
	case mappingDefault:
		return true
	case mappingExact:
		return path == value
	case mappingPrefix:
		return matchPrefix(path, value)
	case mappingExtension:
		return extensionOf(path) == value
	default:
		return false
	}
}

func isNilFilter(filter Filter) bool {
	if filter == nil {
		return true
	}
	reflected := reflect.ValueOf(filter)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

// ErrInvalidDispatchTypes 表示 DispatchType 位集合非法。
var ErrInvalidDispatchTypes = errors.New("arkarta/servlet: invalid dispatch types")

// StatusError 表示可映射到 HTTP 状态码的处理错误。
type StatusError interface {
	error
	StatusCode() int
	PublicMessage() string
}

// HTTPError 是标准 HTTP 状态错误实现。
type HTTPError struct {
	statusCode    int
	publicMessage string
	cause         error
}

// NewHTTPError 创建带 HTTP 状态码的错误。
func NewHTTPError(statusCode int, publicMessage string, cause error) *HTTPError {
	if statusCode < 100 || statusCode > 999 {
		statusCode = http.StatusInternalServerError
	}
	if publicMessage == "" {
		publicMessage = http.StatusText(statusCode)
	}
	if publicMessage == "" {
		publicMessage = "HTTP error"
	}
	return &HTTPError{
		statusCode:    statusCode,
		publicMessage: publicMessage,
		cause:         cause,
	}
}

// Error 返回内部错误文本。
func (e *HTTPError) Error() string {
	if e.cause == nil {
		return fmt.Sprintf("%d %s", e.statusCode, e.publicMessage)
	}
	return fmt.Sprintf("%d %s: %v", e.statusCode, e.publicMessage, e.cause)
}

// Unwrap 返回底层错误。
func (e *HTTPError) Unwrap() error {
	return e.cause
}

// StatusCode 返回 HTTP 状态码。
func (e *HTTPError) StatusCode() int {
	return e.statusCode
}

// PublicMessage 返回可以写给客户端的安全错误信息。
func (e *HTTPError) PublicMessage() string {
	return e.publicMessage
}

type ServletContext = WebApp

// EffectiveMajorVersion 返回当前上下文采用的 Servlet 主版本。
func (a *WebApp) EffectiveMajorVersion() int { return ServletSpecMajorVersion }

// EffectiveMinorVersion 返回当前上下文采用的 Servlet 次版本。
func (a *WebApp) EffectiveMinorVersion() int { return ServletSpecMinorVersion }

// ArkartaMajorVersion 返回 Arkarta Servlet 标准主版本。
func (a *WebApp) ArkartaMajorVersion() int { return ArkartaServletMajorVersion }

// ArkartaMinorVersion 返回 Arkarta Servlet 标准次版本。
func (a *WebApp) ArkartaMinorVersion() int { return ArkartaServletMinorVersion }
