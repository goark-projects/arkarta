package websocket

import "fmt"

// HandshakeError 表示可映射为 HTTP 状态码的 WebSocket 握手错误。
type HandshakeError struct {
	statusCode int
	message    string
	cause      error
}

func newHandshakeError(statusCode int, message string, cause error) *HandshakeError {
	return &HandshakeError{
		statusCode: statusCode,
		message:    message,
		cause:      cause,
	}
}

// Error 返回握手错误文本。
func (e *HandshakeError) Error() string {
	if e == nil {
		return ""
	}
	if e.cause == nil {
		return fmt.Sprintf("%d %s", e.statusCode, e.message)
	}
	return fmt.Sprintf("%d %s: %v", e.statusCode, e.message, e.cause)
}

// Unwrap 返回底层错误。
func (e *HandshakeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// StatusCode 返回 HTTP 状态码。
func (e *HandshakeError) StatusCode() int {
	if e == nil || e.statusCode == 0 {
		return 400
	}
	return e.statusCode
}

// PublicMessage 返回可以写给客户端的安全错误信息。
func (e *HandshakeError) PublicMessage() string {
	if e == nil || e.message == "" {
		return "invalid websocket handshake"
	}
	return e.message
}
