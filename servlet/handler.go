package servlet

import (
	"context"
	"errors"
)

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
