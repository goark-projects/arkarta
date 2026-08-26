package servlet

import "errors"

// RegisterErrorType 注册基于 errors.As 匹配的错误类型错误页。
func RegisterErrorType[T error](registry *ErrorPageRegistry, handler Handler) error {
	return registry.registerErrorType(func(err error) bool {
		var target T
		return errors.As(err, &target)
	}, handler)
}
