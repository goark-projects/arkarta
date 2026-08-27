package web

import (
	"errors"
	"fmt"
)

// ErrNilContext 表示 Web 上下文为空。
var ErrNilContext = errors.New("arkarta/web: context is nil")

// ErrNilHandler 表示路由处理器为空。
var ErrNilHandler = errors.New("arkarta/web: handler is nil")

// ErrInvalidRoutePattern 表示路由路径模式非法。
var ErrInvalidRoutePattern = errors.New("arkarta/web: invalid route pattern")

// ErrDuplicateRoute 表示同一 HTTP 方法和路径模式被重复注册。
var ErrDuplicateRoute = errors.New("arkarta/web: duplicate route")

// ErrUnsupportedMediaType 表示请求 Content-Type 不是 JSON 兼容媒体类型。
var ErrUnsupportedMediaType = errors.New("arkarta/web: unsupported media type")

// BindError 表示请求绑定失败。
type BindError struct {
	cause error
}

func newBindError(cause error) *BindError {
	return &BindError{cause: cause}
}

// Error 返回绑定失败摘要。
func (e *BindError) Error() string {
	if e == nil || e.cause == nil {
		return "arkarta/web: bind failed"
	}
	return fmt.Sprintf("arkarta/web: bind failed: %v", e.cause)
}

// Unwrap 返回底层绑定错误。
func (e *BindError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}
