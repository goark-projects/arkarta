package container

import (
	"goark.dev/arkarta/servlet"
)

// Mapping 表示一个 Servlet 路径映射。
type Mapping struct {
	pattern string
	handler servlet.Handler
	filters []servlet.Filter
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
	return Mapping{
		pattern: pattern,
		handler: handler,
		filters: cloneFilters(filters),
	}, nil
}

// Pattern 返回路径映射模式。
func (m Mapping) Pattern() string {
	return m.pattern
}

// Handler 返回目标处理器。
func (m Mapping) Handler() servlet.Handler {
	return m.handler
}

// Filters 返回过滤器副本。
func (m Mapping) Filters() []servlet.Filter {
	return cloneFilters(m.filters)
}

func (m Mapping) servletHandler() servlet.Handler {
	return servlet.ChainFilters(m.handler, m.filters...)
}
