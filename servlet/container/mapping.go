package container

import (
	"context"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/security"
)

// Mapping 表示一个 Servlet 路径映射。
type Mapping struct {
	pattern          string
	name             string
	handler          servlet.Handler
	filterBindings   []servlet.FilterBinding
	initParam        map[string]string
	loadOnStartup    int
	hasLoadOnStartup bool
	runAsRole        string
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
		initParam:      make(map[string]string),
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

func newRegistrationMapping(pattern, name string, handler servlet.Handler, initParam map[string]string, loadOnStartup int, hasLoadOnStartup bool, runAsRole string) (Mapping, error) {
	mapping, err := NewMapping(pattern, handler)
	if err != nil {
		return Mapping{}, err
	}
	if name != "" {
		mapping.name = name
	}
	mapping.initParam = cloneStringMap(initParam)
	mapping.loadOnStartup = loadOnStartup
	mapping.hasLoadOnStartup = hasLoadOnStartup
	mapping.runAsRole = runAsRole
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

// InitParams 返回 Servlet 初始化参数副本。
func (m Mapping) InitParams() map[string]string {
	return cloneStringMap(m.initParam)
}

// LoadOnStartup 返回注册声明的启动初始化顺序。
func (m Mapping) LoadOnStartup() (int, bool) {
	return m.loadOnStartup, m.hasLoadOnStartup
}

func (m Mapping) servletHandler() servlet.Handler {
	handler := m.handler
	if m.runAsRole != "" {
		handler = servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
			return security.RunAs(req, m.runAsRole, func() error {
				return m.handler.Serve(ctx, req, res)
			})
		})
	}
	target := servlet.ChainFilterBindings(handler, m.filterBindings...)
	return servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		previous, hadPrevious := req.Attribute(servlet.AttributeServletName)
		req.SetAttribute(servlet.AttributeServletName, m.name)
		defer func() {
			if hadPrevious {
				req.SetAttribute(servlet.AttributeServletName, previous)
				return
			}
			req.SetAttribute(servlet.AttributeServletName, nil)
		}()
		return target.Serve(ctx, req, res)
	})
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
