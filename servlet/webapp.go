package servlet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// ErrInvalidContextPath 表示 Web 应用上下文路径非法。
var ErrInvalidContextPath = errors.New("arkarta/servlet: invalid context path")

// WebAppOption 定制 WebApp。
type WebAppOption func(*WebApp) error

// WithContextPath 设置 Web 应用上下文路径。
func WithContextPath(contextPath string) WebAppOption {
	return func(app *WebApp) error {
		if contextPath == "" {
			contextPath = "/"
		}
		if !strings.HasPrefix(contextPath, "/") {
			return ErrInvalidContextPath
		}
		if len(contextPath) > 1 && strings.HasSuffix(contextPath, "/") {
			contextPath = strings.TrimRight(contextPath, "/")
		}
		app.contextPath = contextPath
		return nil
	}
}

// WithInitParam 设置应用初始化参数。
func WithInitParam(name, value string) WebAppOption {
	return func(app *WebApp) error {
		if strings.TrimSpace(name) == "" {
			return ErrInvalidWebAppConfig
		}
		app.initParam[name] = value
		return nil
	}
}

// WithContextListener 添加应用上下文监听器。
func WithContextListener(listener ContextListener) WebAppOption {
	return func(app *WebApp) error {
		if listener != nil {
			app.contextListeners = append(app.contextListeners, listener)
		}
		return nil
	}
}

// WithRequestListener 添加请求生命周期监听器。
func WithRequestListener(listener RequestListener) WebAppOption {
	return func(app *WebApp) error {
		if listener != nil {
			app.requestListeners = append(app.requestListeners, listener)
		}
		return nil
	}
}

// WebApp 表示一个部署单元的应用上下文。
type WebApp struct {
	name                      string
	contextPath               string
	virtualServerName         string
	requestCharacterEncoding  string
	responseCharacterEncoding string
	sessionTimeout            time.Duration
	tempDir                   string
	initParam                 map[string]string
	mimeTypes                 map[string]string
	resourceFS                fs.FS
	logger                    *slog.Logger
	state                     WebAppState

	mu                        sync.RWMutex
	attribute                 map[string]any
	contextListeners          []ContextListener
	requestListeners          []RequestListener
	contextAttributeListeners []ContextAttributeListener
	requestAttributeListeners []RequestAttributeListener
	dispatcherProvider        DispatcherProvider
}

// NewWebApp 创建 Web 应用上下文。
func NewWebApp(name string, options ...WebAppOption) (*WebApp, error) {
	app := &WebApp{
		name:                      name,
		contextPath:               "/",
		virtualServerName:         DefaultVirtualServerName,
		requestCharacterEncoding:  DefaultCharacterEncoding,
		responseCharacterEncoding: DefaultCharacterEncoding,
		sessionTimeout:            DefaultSessionTimeout,
		initParam:                 make(map[string]string),
		mimeTypes:                 defaultMimeMappings(),
		logger:                    slog.Default(),
		attribute:                 make(map[string]any),
		state:                     WebAppStateNew,
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(app); err != nil {
			return nil, err
		}
	}
	return app, nil
}

// Name 返回应用名称。
func (a *WebApp) Name() string {
	return a.name
}

// ContextPath 返回应用上下文路径。
func (a *WebApp) ContextPath() string {
	return a.contextPath
}

// InitParam 返回初始化参数。
func (a *WebApp) InitParam(name string) (string, bool) {
	value, ok := a.initParam[name]
	return value, ok
}

// InitParams 返回初始化参数副本。
func (a *WebApp) InitParams() map[string]string {
	return cloneStringMap(a.initParam)
}

// Attribute 返回应用属性。
func (a *WebApp) Attribute(key string) (any, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	value, ok := a.attribute[key]
	return value, ok
}

// SetAttribute 设置应用属性；传入 nil 会删除该属性。
func (a *WebApp) SetAttribute(key string, value any) {
	a.SetAttributeContext(context.Background(), key, value)
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

const (
	// DefaultVirtualServerName 是标准上下文默认虚拟主机名。
	DefaultVirtualServerName = "default"
	// DefaultCharacterEncoding 是请求与响应的默认字符编码。
	DefaultCharacterEncoding = "utf-8"
)

// DefaultSessionTimeout 是标准会话默认空闲超时。
const DefaultSessionTimeout = 30 * time.Minute

// ErrInvalidWebAppConfig 表示 WebApp 配置非法。
var ErrInvalidWebAppConfig = errors.New("arkarta/servlet: invalid web app config")

// WithVirtualServerName 设置虚拟主机名。
func WithVirtualServerName(name string) WebAppOption {
	return func(app *WebApp) error {
		name = strings.TrimSpace(name)
		if name == "" {
			return ErrInvalidWebAppConfig
		}
		app.virtualServerName = name
		return nil
	}
}

// WithRequestCharacterEncoding 设置默认请求字符编码。
func WithRequestCharacterEncoding(charset string) WebAppOption {
	return func(app *WebApp) error {
		charset = strings.TrimSpace(charset)
		if charset == "" {
			return ErrInvalidWebAppConfig
		}
		app.requestCharacterEncoding = charset
		return nil
	}
}

// WithResponseCharacterEncoding 设置默认响应字符编码。
func WithResponseCharacterEncoding(charset string) WebAppOption {
	return func(app *WebApp) error {
		charset = strings.TrimSpace(charset)
		if charset == "" {
			return ErrInvalidWebAppConfig
		}
		app.responseCharacterEncoding = charset
		return nil
	}
}

// WithSessionTimeout 设置默认会话空闲超时。
func WithSessionTimeout(timeout time.Duration) WebAppOption {
	return func(app *WebApp) error {
		if timeout < 0 {
			return fmt.Errorf("%w: negative session timeout", ErrInvalidWebAppConfig)
		}
		app.sessionTimeout = timeout
		return nil
	}
}

// VirtualServerName 返回虚拟主机名。
func (a *WebApp) VirtualServerName() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.virtualServerName
}

// RequestCharacterEncoding 返回默认请求字符编码。
func (a *WebApp) RequestCharacterEncoding() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.requestCharacterEncoding
}

// ResponseCharacterEncoding 返回默认响应字符编码。
func (a *WebApp) ResponseCharacterEncoding() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.responseCharacterEncoding
}

// SessionTimeout 返回默认会话空闲超时。
func (a *WebApp) SessionTimeout() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sessionTimeout
}

// DispatcherProvider 按路径或名称提供请求分发器。
type DispatcherProvider interface {
	RequestDispatcher(path string) (RequestDispatcher, error)
	NamedDispatcher(name string) (RequestDispatcher, error)
}

// WithDispatcherProvider 设置请求分发器提供者。
func WithDispatcherProvider(provider DispatcherProvider) WebAppOption {
	return func(app *WebApp) error {
		app.dispatcherProvider = provider
		return nil
	}
}

// SetDispatcherProvider 设置请求分发器提供者。
func (a *WebApp) SetDispatcherProvider(provider DispatcherProvider) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.dispatcherProvider = provider
}

// RequestDispatcher 返回指定路径的请求分发器。
func (a *WebApp) RequestDispatcher(path string) (RequestDispatcher, error) {
	a.mu.RLock()
	provider := a.dispatcherProvider
	a.mu.RUnlock()
	if provider == nil {
		return nil, ErrNilRouter
	}
	return provider.RequestDispatcher(path)
}

// NamedDispatcher 返回指定 Servlet 名称的请求分发器。
func (a *WebApp) NamedDispatcher(name string) (RequestDispatcher, error) {
	a.mu.RLock()
	provider := a.dispatcherProvider
	a.mu.RUnlock()
	if provider == nil {
		return nil, ErrDispatcherTargetNotFound
	}
	return provider.NamedDispatcher(name)
}

type includeResponse struct {
	target Response
	header Header
}

func newIncludeResponse(target Response) Response {
	if target == nil {
		return nil
	}
	return &includeResponse{target: target, header: NewHeader()}
}

func (r *includeResponse) Header() Header                    { return r.header }
func (r *includeResponse) SetStatus(int)                     {}
func (r *includeResponse) Status() int                       { return r.target.Status() }
func (r *includeResponse) Write(data []byte) (int, error)    { return r.target.Write(data) }
func (r *includeResponse) WriteString(v string) (int, error) { return r.target.WriteString(v) }
func (r *includeResponse) Flush() error                      { return r.target.Flush() }
func (r *includeResponse) Committed() bool                   { return r.target.Committed() }
func (r *includeResponse) Reset() error                      { return ErrResponseCommitted }
func (r *includeResponse) BodyWriter() io.Writer             { return r }
