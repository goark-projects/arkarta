package nethttp

import "goark.dev/arkarta/servlet"

// Option 定制 net/http 适配器。
type Option func(*adapter)

// WithErrorPages 设置错误页注册表。
func WithErrorPages(registry *servlet.ErrorPageRegistry) Option {
	return func(adapter *adapter) {
		adapter.errorPages = registry
	}
}

// WithRequestContextPath 设置适配器创建请求时使用的 Web 应用上下文路径。
func WithRequestContextPath(contextPath string) Option {
	return func(adapter *adapter) {
		adapter.requestOptions = append(adapter.requestOptions, servlet.WithRequestContextPath(contextPath))
	}
}
