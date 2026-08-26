package registration

import (
	"errors"

	"goark.dev/arkarta/servlet"
)

// ErrNilWebApp 表示动态注册上下文缺少 WebApp。
var ErrNilWebApp = errors.New("arkarta/servlet/registration: web app is nil")

// Context 将 WebApp 与动态注册元模型组合为 ServletContext 风格入口。
type Context struct {
	app      *servlet.WebApp
	registry *Registry
}

// NewContext 创建动态注册上下文。
func NewContext(app *servlet.WebApp, registry *Registry) (*Context, error) {
	if app == nil {
		return nil, ErrNilWebApp
	}
	if registry == nil {
		registry = NewRegistry()
	}
	return &Context{app: app, registry: registry}, nil
}

// WebApp 返回关联的应用上下文。
func (c *Context) WebApp() *servlet.WebApp {
	if c == nil {
		return nil
	}
	return c.app
}

// Registry 返回底层动态注册表。
func (c *Context) Registry() *Registry {
	if c == nil {
		return nil
	}
	return c.registry
}

// AddServlet 注册 Servlet 处理器。
func (c *Context) AddServlet(name string, target servlet.Handler) (*ServletRegistration, error) {
	if c == nil || c.registry == nil {
		return nil, ErrNilRegistry
	}
	if err := c.ensureOpen(); err != nil {
		return nil, err
	}
	return c.registry.AddServlet(name, target)
}

// AddFilter 注册 Filter。
func (c *Context) AddFilter(name string, target servlet.Filter) (*FilterRegistration, error) {
	if c == nil || c.registry == nil {
		return nil, ErrNilRegistry
	}
	if err := c.ensureOpen(); err != nil {
		return nil, err
	}
	return c.registry.AddFilter(name, target)
}

// AddListener 按监听器接口类型注册实例。
func (c *Context) AddListener(listener any) (*ListenerRegistration, error) {
	if c == nil || c.registry == nil {
		return nil, ErrNilRegistry
	}
	if err := c.ensureOpen(); err != nil {
		return nil, err
	}
	return c.registry.AddListener(listener)
}

// SetInitParam 设置 WebApp 初始化参数；返回 false 表示名称已存在。
func (c *Context) SetInitParam(name, value string) (bool, error) {
	if c == nil || c.app == nil {
		return false, ErrNilWebApp
	}
	if err := c.ensureOpen(); err != nil {
		return false, err
	}
	return c.app.SetInitParam(name, value)
}

// SetInitParams 批量设置 WebApp 初始化参数；存在冲突时不写入任何参数。
func (c *Context) SetInitParams(params map[string]string) ([]string, error) {
	if c == nil || c.app == nil {
		return nil, ErrNilWebApp
	}
	if err := c.ensureOpen(); err != nil {
		return nil, err
	}
	return c.app.SetInitParams(params)
}

// ServletRegistration 返回指定 Servlet 注册项。
func (c *Context) ServletRegistration(name string) (*ServletRegistration, bool) {
	if c == nil || c.registry == nil {
		return nil, false
	}
	return c.registry.Servlet(name)
}

// FilterRegistration 返回指定 Filter 注册项。
func (c *Context) FilterRegistration(name string) (*FilterRegistration, bool) {
	if c == nil || c.registry == nil {
		return nil, false
	}
	return c.registry.Filter(name)
}

// Snapshot 返回动态注册快照。
func (c *Context) Snapshot() Snapshot {
	if c == nil || c.registry == nil {
		return Snapshot{}
	}
	return c.registry.Snapshot()
}

// Freeze 关闭动态注册阶段并返回不可变部署快照。
func (c *Context) Freeze() (Snapshot, error) {
	if c == nil || c.registry == nil {
		return Snapshot{}, ErrNilRegistry
	}
	if err := c.ensureOpen(); err != nil {
		return Snapshot{}, err
	}
	return c.registry.Freeze()
}

// Frozen 表示动态注册上下文是否已经冻结。
func (c *Context) Frozen() bool {
	if c == nil || c.registry == nil {
		return false
	}
	return c.registry.Frozen()
}

// RequestDispatcher 返回指定路径的请求分发器。
func (c *Context) RequestDispatcher(path string) (servlet.RequestDispatcher, error) {
	if c == nil || c.app == nil {
		return nil, ErrNilWebApp
	}
	return c.app.RequestDispatcher(path)
}

// NamedDispatcher 返回指定名称的请求分发器。
func (c *Context) NamedDispatcher(name string) (servlet.RequestDispatcher, error) {
	if c == nil || c.app == nil {
		return nil, ErrNilWebApp
	}
	return c.app.NamedDispatcher(name)
}

func (c *Context) ensureOpen() error {
	if c == nil || c.app == nil {
		return ErrNilWebApp
	}
	switch c.app.State() {
	case servlet.WebAppStateNew, servlet.WebAppStateInitialized:
		return nil
	default:
		return ErrRegistrationClosed
	}
}
