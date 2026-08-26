package servlet

import "errors"

// ErrNilFilter 表示过滤器为空。
var ErrNilFilter = errors.New("arkarta/servlet: filter is nil")

// ErrInvalidFilterConfig 表示 Filter 配置非法。
var ErrInvalidFilterConfig = errors.New("arkarta/servlet: invalid filter config")

// FilterBindingOption 定制 FilterBinding。
type FilterBindingOption func(*FilterBinding) error

// FilterBinding 表示一个 Filter 在运行时链路中的映射约束。
type FilterBinding struct {
	name          string
	filter        Filter
	dispatchTypes DispatchTypes
	initParam     map[string]string
}

// NewFilterBinding 创建 Filter 运行时映射。
func NewFilterBinding(name string, filter Filter, options ...FilterBindingOption) (FilterBinding, error) {
	if isNilFilter(filter) {
		return FilterBinding{}, ErrNilFilter
	}
	binding := FilterBinding{
		name:          name,
		filter:        filter,
		dispatchTypes: DispatchOnRequest,
		initParam:     make(map[string]string),
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&binding); err != nil {
			return FilterBinding{}, err
		}
	}
	if err := ValidateDispatchTypes(binding.dispatchTypes); err != nil {
		return FilterBinding{}, err
	}
	binding.dispatchTypes = NormalizeDispatchTypes(binding.dispatchTypes)
	return binding, nil
}

// BindFilter 以指定分发类型快速创建 FilterBinding。
func BindFilter(filter Filter, dispatchTypes ...DispatchType) (FilterBinding, error) {
	dispatchers, err := NewDispatchTypes(dispatchTypes...)
	if err != nil {
		return FilterBinding{}, err
	}
	return NewFilterBinding("", filter, WithFilterDispatchTypes(dispatchers))
}

// WithFilterDispatchTypes 设置 Filter 匹配的分发类型集合。
func WithFilterDispatchTypes(dispatchTypes DispatchTypes) FilterBindingOption {
	return func(binding *FilterBinding) error {
		if err := ValidateDispatchTypes(dispatchTypes); err != nil {
			return err
		}
		binding.dispatchTypes = NormalizeDispatchTypes(dispatchTypes)
		return nil
	}
}

// WithFilterInitParam 设置 Filter 初始化参数。
func WithFilterInitParam(name, value string) FilterBindingOption {
	return func(binding *FilterBinding) error {
		if name == "" {
			return ErrInvalidFilterConfig
		}
		binding.initParam[name] = value
		return nil
	}
}

// Name 返回 Filter 名称。
func (b FilterBinding) Name() string {
	return b.name
}

// Filter 返回 Filter 实例。
func (b FilterBinding) Filter() Filter {
	return b.filter
}

// DispatchTypes 返回匹配的分发类型集合。
func (b FilterBinding) DispatchTypes() DispatchTypes {
	return b.dispatchTypes
}

// InitParams 返回初始化参数副本。
func (b FilterBinding) InitParams() map[string]string {
	return cloneStringMap(b.initParam)
}

// Matches 判断当前分发类型是否应该执行该 Filter。
func (b FilterBinding) Matches(dispatchType DispatchType) bool {
	return b.dispatchTypes.Contains(dispatchType)
}
