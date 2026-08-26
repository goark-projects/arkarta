package nethttp

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"goark.dev/arkarta/servlet"
)

// Handler 将 Arkarta Servlet 处理器适配为标准库 http.Handler。
func Handler(handler servlet.Handler) http.Handler {
	return HandlerWithOptions(handler)
}

// HandlerWithOptions 将 Arkarta Servlet 处理器按配置适配为标准库 http.Handler。
func HandlerWithOptions(handler servlet.Handler, options ...Option) http.Handler {
	adapter := &adapter{handler: handler}
	for _, option := range options {
		if option != nil {
			option(adapter)
		}
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		adapter.ServeHTTP(writer, request)
	})
}

// ServeHTTP 执行一次标准库到 Servlet 的请求适配。
func ServeHTTP(writer http.ResponseWriter, request *http.Request, handler servlet.Handler) {
	Handler(handler).ServeHTTP(writer, request)
}

type adapter struct {
	handler        servlet.Handler
	errorPages     *servlet.ErrorPageRegistry
	requestOptions []servlet.RequestOption
}

func (a *adapter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	response := NewResponse(writer)
	if a.handler == nil {
		writeError(response, servlet.NewHTTPError(http.StatusInternalServerError, "handler is nil", servlet.ErrNilHandler))
		return
	}
	req, err := servlet.NewRequest(request, a.requestOptions...)
	if err != nil {
		writeError(response, err)
		return
	}
	defer response.finish()
	defer a.recoverPanic(request, req, response)

	if err := a.handler.Serve(request.Context(), req, response); err != nil {
		a.writeError(request, req, response, err)
		return
	}
}

func (a *adapter) recoverPanic(httpRequest *http.Request, req *servlet.Request, response *Response) {
	value := recover()
	if value == nil {
		return
	}
	err := fmt.Errorf("panic recovered: %v\n%s", value, debug.Stack())
	a.writeError(httpRequest, req, response, servlet.NewHTTPError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), err))
}

func (a *adapter) writeError(httpRequest *http.Request, req *servlet.Request, response *Response, err error) {
	statusCode, _ := errorStatus(err)
	if handled, dispatchErr := a.errorPages.Handle(httpRequest.Context(), req, response, statusCode, err); handled && dispatchErr == nil {
		return
	}
	writeError(response, err)
}

func writeError(response *Response, err error) {
	if response.Committed() {
		return
	}

	statusCode, message := errorStatus(err)
	if message == "" {
		message = http.StatusText(statusCode)
	}
	if message == "" {
		message = "HTTP error"
	}

	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.SetStatus(statusCode)
	_, _ = response.WriteString(message + "\n")
}

func errorStatus(err error) (int, string) {
	statusCode := http.StatusInternalServerError
	message := http.StatusText(statusCode)
	var statusErr servlet.StatusError
	if errors.As(err, &statusErr) {
		statusCode = statusErr.StatusCode()
		message = statusErr.PublicMessage()
	}
	return statusCode, message
}
