package servlet

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
)

// ErrNilHTTPRequest 表示构造请求时传入了空的标准库请求。
var ErrNilHTTPRequest = errors.New("arkarta/servlet: http request is nil")

// DispatchType 表示请求进入处理链的原因。
type DispatchType uint8

const (
	// DispatchRequest 表示客户端原始请求。
	DispatchRequest DispatchType = iota
	// DispatchForward 表示服务端转发。
	DispatchForward
	// DispatchInclude 表示服务端包含。
	DispatchInclude
	// DispatchError 表示错误分发。
	DispatchError
	// DispatchAsync 表示异步分发。
	DispatchAsync
)

// RequestOption 定制 Request 构造行为。
type RequestOption func(*Request)

// WithDispatchType 设置请求分发类型。
func WithDispatchType(dispatchType DispatchType) RequestOption {
	return func(req *Request) {
		req.dispatchType = dispatchType
	}
}

// Request 表示容器传给应用的请求视图。
type Request struct {
	httpRequest  *http.Request
	dispatchType DispatchType

	mu        sync.RWMutex
	attribute map[string]any
}

// NewRequest 从标准库请求创建 Arkarta Servlet 请求。
func NewRequest(httpRequest *http.Request, options ...RequestOption) (*Request, error) {
	if httpRequest == nil {
		return nil, ErrNilHTTPRequest
	}
	req := &Request{
		httpRequest:  httpRequest,
		dispatchType: DispatchRequest,
		attribute:    make(map[string]any),
	}
	for _, option := range options {
		if option != nil {
			option(req)
		}
	}
	return req, nil
}

// Context 返回请求上下文。
func (r *Request) Context() context.Context {
	return r.httpRequest.Context()
}

// Method 返回 HTTP 方法。
func (r *Request) Method() string {
	return r.httpRequest.Method
}

// Protocol 返回 HTTP 协议版本。
func (r *Request) Protocol() string {
	return r.httpRequest.Proto
}

// Scheme 返回请求协议。
func (r *Request) Scheme() string {
	if r.httpRequest.URL != nil && r.httpRequest.URL.Scheme != "" {
		return r.httpRequest.URL.Scheme
	}
	if r.httpRequest.TLS != nil {
		return "https"
	}
	return "http"
}

// Host 返回请求主机。
func (r *Request) Host() string {
	return r.httpRequest.Host
}

// Path 返回 URL 路径。
func (r *Request) Path() string {
	if r.httpRequest.URL == nil {
		return ""
	}
	return r.httpRequest.URL.Path
}

// Query 返回 URL 查询参数副本。
func (r *Request) Query() url.Values {
	if r.httpRequest.URL == nil {
		return url.Values{}
	}
	return r.httpRequest.URL.Query()
}

// Header 返回请求头。
func (r *Request) Header() http.Header {
	return r.httpRequest.Header
}

// Cookie 返回指定名称的 Cookie。
func (r *Request) Cookie(name string) (*http.Cookie, error) {
	return r.httpRequest.Cookie(name)
}

// Body 返回请求体读取器。
func (r *Request) Body() io.ReadCloser {
	return r.httpRequest.Body
}

// ContentLength 返回请求体长度。
func (r *Request) ContentLength() int64 {
	return r.httpRequest.ContentLength
}

// RemoteAddr 返回远端网络地址。
func (r *Request) RemoteAddr() string {
	return r.httpRequest.RemoteAddr
}

// IsSecure 表示请求是否通过安全传输进入容器。
func (r *Request) IsSecure() bool {
	return r.Scheme() == "https"
}

// Attribute 返回请求属性。
func (r *Request) Attribute(key string) (any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.attribute[key]
	return value, ok
}

// SetAttribute 设置请求属性；传入 nil 会删除该属性。
func (r *Request) SetAttribute(key string, value any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if value == nil {
		delete(r.attribute, key)
		return
	}
	r.attribute[key] = value
}

// DispatchType 返回当前分发类型。
func (r *Request) DispatchType() DispatchType {
	return r.dispatchType
}

// HTTPRequest 返回底层标准库请求。
func (r *Request) HTTPRequest() *http.Request {
	return r.httpRequest
}
