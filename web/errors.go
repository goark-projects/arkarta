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

// ErrInvalidBindTarget 表示绑定目标必须是非空结构体指针。
var ErrInvalidBindTarget = errors.New("arkarta/web: invalid bind target")

// ErrInvalidParameter 表示请求参数不能转换为目标类型。
var ErrInvalidParameter = errors.New("arkarta/web: invalid parameter")

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

// ParameterError 表示单个请求参数转换失败。
type ParameterError struct {
	Name  string
	Value string
	Type  string
	Cause error
}

// Error 返回参数转换失败摘要。
func (e *ParameterError) Error() string {
	if e == nil {
		return "arkarta/web: invalid parameter"
	}
	if e.Cause == nil {
		return fmt.Sprintf("arkarta/web: invalid parameter %q as %s", e.Name, e.Type)
	}
	return fmt.Sprintf("arkarta/web: invalid parameter %q=%q as %s: %v", e.Name, e.Value, e.Type, e.Cause)
}

// Unwrap 返回底层转换错误。
func (e *ParameterError) Unwrap() error {
	if e == nil || e.Cause == nil {
		return ErrInvalidParameter
	}
	return e.Cause
}
