package registration

import "goark.dev/arkarta/servlet"

// DispatcherType 表示一次请求进入过滤器链的分发来源。
type DispatcherType = servlet.DispatchType

const (
	// DispatcherRequest 表示客户端原始请求。
	DispatcherRequest = servlet.DispatchRequest
	// DispatcherForward 表示服务端 forward 分发。
	DispatcherForward = servlet.DispatchForward
	// DispatcherInclude 表示服务端 include 分发。
	DispatcherInclude = servlet.DispatchInclude
	// DispatcherAsync 表示异步分发。
	DispatcherAsync = servlet.DispatchAsync
	// DispatcherError 表示错误页分发。
	DispatcherError = servlet.DispatchError
)

// DispatcherTypes 是 DispatcherType 的紧凑位集合。
type DispatcherTypes = servlet.DispatchTypes

const (
	// DispatchOnRequest 匹配客户端原始请求。
	DispatchOnRequest = servlet.DispatchOnRequest
	// DispatchOnForward 匹配 forward 分发。
	DispatchOnForward = servlet.DispatchOnForward
	// DispatchOnInclude 匹配 include 分发。
	DispatchOnInclude = servlet.DispatchOnInclude
	// DispatchOnAsync 匹配异步分发。
	DispatchOnAsync = servlet.DispatchOnAsync
	// DispatchOnError 匹配错误页分发。
	DispatchOnError = servlet.DispatchOnError
)

// NewDispatcherTypes 从枚举值构造位集合。
func NewDispatcherTypes(types ...DispatcherType) (DispatcherTypes, error) {
	dispatchers, err := servlet.NewDispatchTypes(types...)
	if err != nil {
		return 0, ErrInvalidDispatcherTypes
	}
	return dispatchers, nil
}

func normalizeDispatcherTypes(dispatchers DispatcherTypes) DispatcherTypes {
	return servlet.NormalizeDispatchTypes(dispatchers)
}

func validateDispatcherTypes(dispatchers DispatcherTypes) error {
	if err := servlet.ValidateDispatchTypes(dispatchers); err != nil {
		return ErrInvalidDispatcherTypes
	}
	return nil
}
