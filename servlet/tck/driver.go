package tck

import (
	"context"
	"testing"

	"goark.dev/arkarta/servlet"
)

// Request 描述 TCK 发给容器的传输无关 HTTP 请求。
type Request struct {
	Method string
	Target string
	Header servlet.Header
	Body   []byte
}

// NewRequest 创建 TCK 请求。
func NewRequest(method, target string) Request {
	return Request{Method: method, Target: target, Header: servlet.NewHeader()}
}

// Response 描述容器返回给 TCK 的传输无关 HTTP 响应。
type Response struct {
	Status int
	Header servlet.Header
	Body   []byte
}

// Driver 执行一次 Servlet Handler 请求交换。
type Driver interface {
	Exchange(ctx context.Context, handler servlet.Handler, request Request) (Response, error)
}

// DriverFunc 将函数适配为 TCK 驱动。
type DriverFunc func(ctx context.Context, handler servlet.Handler, request Request) (Response, error)

// Exchange 执行一次 Servlet Handler 请求交换。
func (f DriverFunc) Exchange(ctx context.Context, handler servlet.Handler, request Request) (Response, error) {
	return f(ctx, handler, request)
}

func exchange(t *testing.T, driver Driver, handler servlet.Handler, request Request) Response {
	t.Helper()
	if request.Header == nil {
		request.Header = servlet.NewHeader()
	}
	response, err := driver.Exchange(t.Context(), handler, request)
	if err != nil {
		t.Fatalf("TCK exchange failed: %v", err)
	}
	if response.Header == nil {
		response.Header = servlet.NewHeader()
	}
	return response
}
