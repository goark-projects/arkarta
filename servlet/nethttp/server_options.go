package nethttp

import (
	"net/http"
	"time"
)

// ServerOption 定制 net/http 参考服务器。
type ServerOption func(*http.Server)

// WithAddress 设置监听地址。
func WithAddress(address string) ServerOption {
	return func(server *http.Server) {
		if address != "" {
			server.Addr = address
		}
	}
}

// WithReadTimeout 设置完整请求读取超时。
func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(server *http.Server) {
		server.ReadTimeout = timeout
	}
}

// WithReadHeaderTimeout 设置请求头读取超时。
func WithReadHeaderTimeout(timeout time.Duration) ServerOption {
	return func(server *http.Server) {
		server.ReadHeaderTimeout = timeout
	}
}

// WithWriteTimeout 设置响应写出超时。
func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(server *http.Server) {
		server.WriteTimeout = timeout
	}
}

// WithIdleTimeout 设置 keep-alive 空闲超时。
func WithIdleTimeout(timeout time.Duration) ServerOption {
	return func(server *http.Server) {
		server.IdleTimeout = timeout
	}
}

// WithMaxHeaderBytes 设置请求头最大字节数。
func WithMaxHeaderBytes(size int) ServerOption {
	return func(server *http.Server) {
		if size > 0 {
			server.MaxHeaderBytes = size
		}
	}
}
