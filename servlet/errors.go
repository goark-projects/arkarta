package servlet

import (
	"fmt"
	"net/http"
)

// StatusError 表示可映射到 HTTP 状态码的处理错误。
type StatusError interface {
	error
	StatusCode() int
	PublicMessage() string
}

// HTTPError 是标准 HTTP 状态错误实现。
type HTTPError struct {
	statusCode    int
	publicMessage string
	cause         error
}

// NewHTTPError 创建带 HTTP 状态码的错误。
func NewHTTPError(statusCode int, publicMessage string, cause error) *HTTPError {
	if statusCode < 100 || statusCode > 999 {
		statusCode = http.StatusInternalServerError
	}
	if publicMessage == "" {
		publicMessage = http.StatusText(statusCode)
	}
	if publicMessage == "" {
		publicMessage = "HTTP error"
	}
	return &HTTPError{
		statusCode:    statusCode,
		publicMessage: publicMessage,
		cause:         cause,
	}
}

// Error 返回内部错误文本。
func (e *HTTPError) Error() string {
	if e.cause == nil {
		return fmt.Sprintf("%d %s", e.statusCode, e.publicMessage)
	}
	return fmt.Sprintf("%d %s: %v", e.statusCode, e.publicMessage, e.cause)
}

// Unwrap 返回底层错误。
func (e *HTTPError) Unwrap() error {
	return e.cause
}

// StatusCode 返回 HTTP 状态码。
func (e *HTTPError) StatusCode() int {
	return e.statusCode
}

// PublicMessage 返回可以写给客户端的安全错误信息。
func (e *HTTPError) PublicMessage() string {
	return e.publicMessage
}
