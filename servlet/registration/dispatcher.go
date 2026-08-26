package registration

// DispatcherType 表示一次请求进入过滤器链的分发来源。
type DispatcherType uint8

const (
	// DispatcherRequest 表示客户端原始请求。
	DispatcherRequest DispatcherType = iota
	// DispatcherForward 表示服务端 forward 分发。
	DispatcherForward
	// DispatcherInclude 表示服务端 include 分发。
	DispatcherInclude
	// DispatcherAsync 表示异步分发。
	DispatcherAsync
	// DispatcherError 表示错误页分发。
	DispatcherError
)

// DispatcherTypes 是 DispatcherType 的紧凑位集合。
type DispatcherTypes uint8

const (
	// DispatchOnRequest 匹配客户端原始请求。
	DispatchOnRequest DispatcherTypes = 1 << iota
	// DispatchOnForward 匹配 forward 分发。
	DispatchOnForward
	// DispatchOnInclude 匹配 include 分发。
	DispatchOnInclude
	// DispatchOnAsync 匹配异步分发。
	DispatchOnAsync
	// DispatchOnError 匹配错误页分发。
	DispatchOnError
)

const allDispatcherTypes = DispatchOnRequest |
	DispatchOnForward |
	DispatchOnInclude |
	DispatchOnAsync |
	DispatchOnError

// NewDispatcherTypes 从枚举值构造位集合。
func NewDispatcherTypes(types ...DispatcherType) (DispatcherTypes, error) {
	if len(types) == 0 {
		return DispatchOnRequest, nil
	}
	var result DispatcherTypes
	for _, item := range types {
		mask, ok := dispatcherMask(item)
		if !ok {
			return 0, ErrInvalidDispatcherTypes
		}
		result |= mask
	}
	return result, nil
}

// Contains 判断位集合是否包含指定 DispatcherType。
func (d DispatcherTypes) Contains(dispatcherType DispatcherType) bool {
	mask, ok := dispatcherMask(dispatcherType)
	return ok && d&mask != 0
}

// List 按 Servlet 规范顺序返回 DispatcherType 切片。
func (d DispatcherTypes) List() []DispatcherType {
	d = normalizeDispatcherTypes(d)
	result := make([]DispatcherType, 0, 5)
	for _, item := range []DispatcherType{
		DispatcherRequest,
		DispatcherForward,
		DispatcherInclude,
		DispatcherAsync,
		DispatcherError,
	} {
		if d.Contains(item) {
			result = append(result, item)
		}
	}
	return result
}

func normalizeDispatcherTypes(dispatchers DispatcherTypes) DispatcherTypes {
	if dispatchers == 0 {
		return DispatchOnRequest
	}
	return dispatchers
}

func validateDispatcherTypes(dispatchers DispatcherTypes) error {
	if normalizeDispatcherTypes(dispatchers)&^allDispatcherTypes != 0 {
		return ErrInvalidDispatcherTypes
	}
	return nil
}

func dispatcherMask(dispatcherType DispatcherType) (DispatcherTypes, bool) {
	switch dispatcherType {
	case DispatcherRequest:
		return DispatchOnRequest, true
	case DispatcherForward:
		return DispatchOnForward, true
	case DispatcherInclude:
		return DispatchOnInclude, true
	case DispatcherAsync:
		return DispatchOnAsync, true
	case DispatcherError:
		return DispatchOnError, true
	default:
		return 0, false
	}
}
