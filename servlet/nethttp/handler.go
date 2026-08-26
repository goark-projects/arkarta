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
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ServeHTTP(writer, request, handler)
	})
}

// ServeHTTP 执行一次标准库到 Servlet 的请求适配。
func ServeHTTP(writer http.ResponseWriter, request *http.Request, handler servlet.Handler) {
	response := NewResponse(writer)
	if handler == nil {
		writeError(response, servlet.NewHTTPError(http.StatusInternalServerError, "handler is nil", servlet.ErrNilHandler))
		return
	}
	req, err := servlet.NewRequest(request)
	if err != nil {
		writeError(response, err)
		return
	}
	defer recoverPanic(response)

	if err := handler.Serve(request.Context(), req, response); err != nil {
		writeError(response, err)
	}
}

func recoverPanic(response *Response) {
	value := recover()
	if value == nil {
		return
	}
	err := fmt.Errorf("panic recovered: %v\n%s", value, debug.Stack())
	writeError(response, servlet.NewHTTPError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), err))
}

func writeError(response *Response, err error) {
	if response.Committed() {
		return
	}

	statusCode := http.StatusInternalServerError
	message := http.StatusText(http.StatusInternalServerError)
	var statusErr servlet.StatusError
	if errors.As(err, &statusErr) {
		statusCode = statusErr.StatusCode()
		message = statusErr.PublicMessage()
	}
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
