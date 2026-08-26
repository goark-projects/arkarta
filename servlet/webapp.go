package servlet

import (
	"errors"
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
	initParam                 map[string]string
	mimeTypes                 map[string]string
	resourceFS                fs.FS
	logger                    *slog.Logger
	state                     WebAppState

	mu               sync.RWMutex
	attribute        map[string]any
	contextListeners []ContextListener
	requestListeners []RequestListener
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
	a.mu.Lock()
	defer a.mu.Unlock()
	if value == nil {
		delete(a.attribute, key)
		return
	}
	a.attribute[key] = value
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
