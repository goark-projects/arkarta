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

// WithRequestOptions 追加传输层中立的 Servlet 请求构造选项。
func WithRequestOptions(options ...servlet.RequestOption) Option {
	copied := append([]servlet.RequestOption(nil), options...)
	return func(adapter *adapter) {
		for _, option := range copied {
			if option != nil {
				adapter.requestOptions = append(adapter.requestOptions, option)
			}
		}
	}
}
