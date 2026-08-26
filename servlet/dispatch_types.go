package servlet

// DispatchTypes 是 DispatchType 的紧凑位集合。
type DispatchTypes uint8

const (
	// DispatchOnRequest 匹配客户端原始请求。
	DispatchOnRequest DispatchTypes = 1 << iota
	// DispatchOnForward 匹配服务端 forward。
	DispatchOnForward
	// DispatchOnInclude 匹配服务端 include。
	DispatchOnInclude
	// DispatchOnError 匹配错误分发。
	DispatchOnError
	// DispatchOnAsync 匹配异步分发。
	DispatchOnAsync
)

const allDispatchTypes = DispatchOnRequest |
	DispatchOnForward |
	DispatchOnInclude |
	DispatchOnError |
	DispatchOnAsync

// NewDispatchTypes 从枚举值构造位集合；未传入时默认匹配 REQUEST。
func NewDispatchTypes(types ...DispatchType) (DispatchTypes, error) {
	if len(types) == 0 {
		return DispatchOnRequest, nil
	}
	var result DispatchTypes
	for _, item := range types {
		mask, ok := dispatchMask(item)
		if !ok {
			return 0, ErrInvalidDispatchTypes
		}
		result |= mask
	}
	return result, nil
}

// Contains 判断位集合是否包含指定 DispatchType。
func (d DispatchTypes) Contains(dispatchType DispatchType) bool {
	mask, ok := dispatchMask(dispatchType)
	return ok && d&mask != 0
}

// List 按 Arkarta Servlet 运行时顺序返回 DispatchType 切片。
func (d DispatchTypes) List() []DispatchType {
	d = NormalizeDispatchTypes(d)
	result := make([]DispatchType, 0, 5)
	for _, item := range []DispatchType{
		DispatchRequest,
		DispatchForward,
		DispatchInclude,
		DispatchError,
		DispatchAsync,
	} {
		if d.Contains(item) {
			result = append(result, item)
		}
	}
	return result
}

// NormalizeDispatchTypes 将空位集合归一化为 REQUEST。
func NormalizeDispatchTypes(dispatchers DispatchTypes) DispatchTypes {
	if dispatchers == 0 {
		return DispatchOnRequest
	}
	return dispatchers
}

// ValidateDispatchTypes 校验位集合是否只包含合法 DispatchType。
func ValidateDispatchTypes(dispatchers DispatchTypes) error {
	if NormalizeDispatchTypes(dispatchers)&^allDispatchTypes != 0 {
		return ErrInvalidDispatchTypes
	}
	return nil
}

func dispatchMask(dispatchType DispatchType) (DispatchTypes, bool) {
	switch dispatchType {
	case DispatchRequest:
		return DispatchOnRequest, true
	case DispatchForward:
		return DispatchOnForward, true
	case DispatchInclude:
		return DispatchOnInclude, true
	case DispatchError:
		return DispatchOnError, true
	case DispatchAsync:
		return DispatchOnAsync, true
	default:
		return 0, false
	}
}
