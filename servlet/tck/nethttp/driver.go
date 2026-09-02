// Package nethttp 提供 Arkarta TCK 的标准库 HTTP 参考驱动。
package nethttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"

	"goark.dev/arkarta/servlet"
	servletnethttp "goark.dev/arkarta/servlet/nethttp"
	"goark.dev/arkarta/servlet/tck"
)

// HandlerFactory 将 Servlet Handler 暴露为标准库 HTTP Handler。
type HandlerFactory func(servlet.Handler) http.Handler

// Driver 使用标准库内存请求执行 TCK 交换。
type Driver struct {
	factory HandlerFactory
}

// NewDriver 创建标准库 HTTP TCK 驱动；factory 为空时使用 Arkarta 参考适配器。
func NewDriver(factory HandlerFactory) *Driver {
	if factory == nil {
		factory = servletnethttp.Handler
	}
	return &Driver{factory: factory}
}

// Exchange 执行一次标准库 HTTP 请求交换。
func (d *Driver) Exchange(ctx context.Context, handler servlet.Handler, request tck.Request) (tck.Response, error) {
	httpRequest := httptest.NewRequestWithContext(ctx, request.Method, request.Target, bytes.NewReader(request.Body))
	if request.Header != nil {
		request.Header.Visit(func(name, value string) bool {
			httpRequest.Header.Add(name, value)
			return true
		})
	}
	recorder := httptest.NewRecorder()
	d.factory(handler).ServeHTTP(recorder, httpRequest)
	return tck.Response{
		Status: recorder.Code,
		Header: cloneHeader(recorder.Header()),
		Body:   append([]byte(nil), recorder.Body.Bytes()...),
	}, nil
}

func cloneHeader(source http.Header) servlet.Header {
	target := servlet.NewHeader()
	for name, values := range source {
		for _, value := range values {
			target.Add(name, value)
		}
	}
	return target
}
