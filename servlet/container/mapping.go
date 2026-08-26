package container

import (
	"goark.dev/arkarta/servlet"
)

// Mapping 表示一个 Servlet 路径映射。
type Mapping struct {
	pattern        string
	name           string
	handler        servlet.Handler
	filterBindings []servlet.FilterBinding
	initParam      map[string]string
}

// NewMapping 创建路径映射。
func NewMapping(pattern string, handler servlet.Handler, filters ...servlet.Filter) (Mapping, error) {
	if handler == nil {
		return Mapping{}, servlet.ErrNilHandler
	}
	router := servlet.NewRouter()
	if err := router.Handle(pattern, handler); err != nil {
		return Mapping{}, err
	}
	bindings, err := requestFilterBindings(filters)
	if err != nil {
		return Mapping{}, err
	}
	return Mapping{
		pattern:        pattern,
		name:           pattern,
		handler:        handler,
		filterBindings: bindings,
	}, nil
}

func newServletMapping(pattern, name string, handler servlet.Servlet, filters ...servlet.Filter) (Mapping, error) {
	if name == "" {
		name = pattern
	}
	mapping, err := NewMapping(pattern, handler, filters...)
	if err != nil {
		return Mapping{}, err
	}
	mapping.name = name
	return mapping, nil
}

// Pattern 返回路径映射模式。
func (m Mapping) Pattern() string {
	return m.pattern
}

// Name 返回映射名称。
func (m Mapping) Name() string {
	return m.name
}

// Handler 返回目标处理器。
func (m Mapping) Handler() servlet.Handler {
	return m.handler
}

// Filters 返回过滤器副本。
func (m Mapping) Filters() []servlet.Filter {
	filters := make([]servlet.Filter, 0, len(m.filterBindings))
	for _, binding := range m.filterBindings {
		if binding.Filter() != nil {
			filters = append(filters, binding.Filter())
		}
	}
	return filters
}

// FilterBindings 返回带 DispatcherType 约束的过滤器映射副本。
func (m Mapping) FilterBindings() []servlet.FilterBinding {
	return cloneFilterBindings(m.filterBindings)
}

func (m Mapping) servletHandler() servlet.Handler {
	return servlet.ChainFilterBindings(m.handler, m.filterBindings...)
}

func (m Mapping) servletConfig(app *servlet.WebApp) servlet.ServletConfig {
	return servlet.NewServletConfig(m.name, app, m.initParam)
}

func requestFilterBindings(filters []servlet.Filter) ([]servlet.FilterBinding, error) {
	bindings := make([]servlet.FilterBinding, 0, len(filters))
	for _, filter := range filters {
		if filter == nil {
			continue
		}
		binding, err := servlet.NewFilterBinding("", filter)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, nil
}
