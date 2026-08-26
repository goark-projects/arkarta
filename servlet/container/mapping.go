package container

import (
	"goark.dev/arkarta/servlet"
)

// Mapping 表示一个 Servlet 路径映射。
type Mapping struct {
	pattern   string
	name      string
	handler   servlet.Handler
	filters   []servlet.Filter
	initParam map[string]string
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
		name:    pattern,
		handler: handler,
		filters: cloneFilters(filters),
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
	return cloneFilters(m.filters)
}

func (m Mapping) servletHandler() servlet.Handler {
	return servlet.ChainFilters(m.handler, m.filters...)
}

func (m Mapping) servletConfig(app *servlet.WebApp) servlet.ServletConfig {
	return servlet.NewServletConfig(m.name, app, m.initParam)
}
