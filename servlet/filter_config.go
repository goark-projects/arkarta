package servlet

import "context"

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
