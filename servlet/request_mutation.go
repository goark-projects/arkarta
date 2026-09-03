package servlet

import (
	"io"
	"strings"
)

// SetMethod 替换过滤器链后续处理器看到的请求方法。
func (r *Request) SetMethod(method string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.method = method
	if r.httpRequest != nil {
		r.httpRequest.Method = method
	}
}

// SetScheme 替换过滤器链后续处理器看到的请求协议。
func (r *Request) SetScheme(scheme string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scheme = scheme
	if r.httpRequest != nil && r.httpRequest.URL != nil {
		r.httpRequest.URL.Scheme = scheme
	}
}

// SetHost 替换过滤器链后续处理器看到的请求主机。
func (r *Request) SetHost(host string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.host = host
	if r.httpRequest != nil {
		r.httpRequest.Host = host
		if r.httpRequest.URL != nil {
			r.httpRequest.URL.Host = host
		}
	}
}

// SetRemoteAddr 替换过滤器链后续处理器看到的远端地址。
func (r *Request) SetRemoteAddr(address string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.remoteAddr = address
	if r.httpRequest != nil {
		r.httpRequest.RemoteAddr = address
	}
}

// SetBody 替换请求体及其长度；空请求体会被规范化为空读取器。
func (r *Request) SetBody(body io.ReadCloser, contentLength int64) {
	if r == nil {
		return
	}
	if body == nil {
		body = io.NopCloser(strings.NewReader(""))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.body = body
	r.contentLength = contentLength
	if r.httpRequest != nil {
		r.httpRequest.Body = body
		r.httpRequest.ContentLength = contentLength
	}
}
