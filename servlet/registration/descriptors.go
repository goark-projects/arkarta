package registration

import (
	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/security"
)

// ServletDescriptor 是 Servlet 注册元数据快照。
type ServletDescriptor struct {
	name             string
	className        string
	initParam        map[string]string
	asyncSupported   bool
	handler          servlet.Handler
	mappings         []string
	loadOnStartup    int
	hasLoadOnStartup bool
	runAsRole        string
	multipartConfig  *multipart.Config
	securityConfig   *security.Constraint
}

// FilterDescriptor 是 Filter 注册元数据快照。
type FilterDescriptor struct {
	name                string
	className           string
	initParam           map[string]string
	asyncSupported      bool
	filter              servlet.Filter
	urlPatternMappings  []URLPatternMapping
	servletNameMappings []ServletNameMapping
}

// ListenerDescriptor 是 Listener 注册元数据快照。
type ListenerDescriptor struct {
	kind      ListenerKind
	className string
	listener  any
	order     int
}

// Name 返回 Servlet 名称。
func (d ServletDescriptor) Name() string {
	return d.name
}

// ClassName 返回 Servlet 实现类型名。
func (d ServletDescriptor) ClassName() string {
	return d.className
}

// InitParams 返回初始化参数副本。
func (d ServletDescriptor) InitParams() map[string]string {
	return cloneStringMap(d.initParam)
}

// AsyncSupported 返回 Servlet 是否支持异步处理。
func (d ServletDescriptor) AsyncSupported() bool {
	return d.asyncSupported
}

// Handler 返回 Servlet 处理器实例。
func (d ServletDescriptor) Handler() servlet.Handler {
	return d.handler
}

// Mappings 返回 Servlet URL 映射副本。
func (d ServletDescriptor) Mappings() []string {
	return cloneStrings(d.mappings)
}

// LoadOnStartup 返回启动初始化顺序。
func (d ServletDescriptor) LoadOnStartup() (int, bool) {
	return d.loadOnStartup, d.hasLoadOnStartup
}

// RunAsRole 返回 Servlet 执行身份角色。
func (d ServletDescriptor) RunAsRole() string {
	return d.runAsRole
}

// MultipartConfig 返回 Servlet multipart 解析配置。
func (d ServletDescriptor) MultipartConfig() (multipart.Config, bool) {
	if d.multipartConfig == nil {
		return multipart.Config{}, false
	}
	return *d.multipartConfig, true
}

// SecurityConfig 返回 Servlet 声明式安全约束。
func (d ServletDescriptor) SecurityConfig() (security.Constraint, bool) {
	if d.securityConfig == nil {
		return security.Constraint{}, false
	}
	return *d.securityConfig, true
}

func (d ServletDescriptor) clone() ServletDescriptor {
	d.initParam = cloneStringMap(d.initParam)
	d.mappings = cloneStrings(d.mappings)
	d.multipartConfig = cloneMultipartConfig(d.multipartConfig)
	d.securityConfig = cloneSecurityConfig(d.securityConfig)
	return d
}

// Name 返回 Filter 名称。
func (d FilterDescriptor) Name() string {
	return d.name
}

// ClassName 返回 Filter 实现类型名。
func (d FilterDescriptor) ClassName() string {
	return d.className
}

// InitParams 返回初始化参数副本。
func (d FilterDescriptor) InitParams() map[string]string {
	return cloneStringMap(d.initParam)
}

// AsyncSupported 返回 Filter 是否支持异步处理。
func (d FilterDescriptor) AsyncSupported() bool {
	return d.asyncSupported
}

// Filter 返回 Filter 实例。
func (d FilterDescriptor) Filter() servlet.Filter {
	return d.filter
}

// URLPatternMappings 返回 URL 模式映射副本。
func (d FilterDescriptor) URLPatternMappings() []URLPatternMapping {
	return cloneURLPatternMappings(d.urlPatternMappings)
}

// ServletNameMappings 返回 Servlet 名称映射副本。
func (d FilterDescriptor) ServletNameMappings() []ServletNameMapping {
	return cloneServletNameMappings(d.servletNameMappings)
}

func (d FilterDescriptor) clone() FilterDescriptor {
	d.initParam = cloneStringMap(d.initParam)
	d.urlPatternMappings = cloneURLPatternMappings(d.urlPatternMappings)
	d.servletNameMappings = cloneServletNameMappings(d.servletNameMappings)
	return d
}

// Kind 返回 Listener 类型。
func (d ListenerDescriptor) Kind() ListenerKind {
	return d.kind
}

// ClassName 返回 Listener 实现类型名。
func (d ListenerDescriptor) ClassName() string {
	return d.className
}

// Listener 返回 Listener 实例。
func (d ListenerDescriptor) Listener() any {
	return d.listener
}

// Order 返回 Listener 注册顺序。
func (d ListenerDescriptor) Order() int {
	return d.order
}

func (r *ServletRegistration) descriptorLocked() ServletDescriptor {
	return ServletDescriptor{
		name:             r.name,
		className:        r.className,
		initParam:        cloneStringMap(r.initParam),
		asyncSupported:   r.asyncSupported,
		handler:          r.handler,
		mappings:         cloneStrings(r.mappings),
		loadOnStartup:    r.loadOnStartup,
		hasLoadOnStartup: r.hasLoadOnStartup,
		runAsRole:        r.runAsRole,
		multipartConfig:  cloneMultipartConfig(r.multipartConfig),
		securityConfig:   cloneSecurityConfig(r.securityConfig),
	}
}

func (r *FilterRegistration) descriptorLocked() FilterDescriptor {
	return FilterDescriptor{
		name:                r.name,
		className:           r.className,
		initParam:           cloneStringMap(r.initParam),
		asyncSupported:      r.asyncSupported,
		filter:              r.filter,
		urlPatternMappings:  cloneURLPatternMappings(r.urlPatternMappings),
		servletNameMappings: cloneServletNameMappings(r.servletNameMappings),
	}
}

func cloneMultipartConfig(src *multipart.Config) *multipart.Config {
	if src == nil {
		return nil
	}
	dst := *src
	return &dst
}

func cloneSecurityConfig(src *security.Constraint) *security.Constraint {
	if src == nil {
		return nil
	}
	dst := src.Clone()
	return &dst
}

func (r *ListenerRegistration) descriptorLocked() ListenerDescriptor {
	return ListenerDescriptor{
		kind:      r.kind,
		className: r.className,
		listener:  r.listener,
		order:     r.order,
	}
}
