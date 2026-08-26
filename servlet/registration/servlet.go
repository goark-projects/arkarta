package registration

import (
	"context"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/security"
)

// ServletRegistration 保存单个 Servlet 的动态注册信息。
type ServletRegistration struct {
	commonRegistration

	handler          servlet.Handler
	mappings         []string
	loadOnStartup    int
	hasLoadOnStartup bool
	runAsRole        string
	multipartConfig  *multipart.Config
	securityConfig   *security.Constraint
}

// Handler 返回注册的 Servlet 处理器。
func (r *ServletRegistration) Handler() servlet.Handler {
	if r == nil || r.owner == nil {
		return nil
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.handler
}

// AddMapping 为 Servlet 增加 URL 映射；存在冲突时不写入任何映射。
func (r *ServletRegistration) AddMapping(patterns ...string) ([]string, error) {
	if r == nil || r.owner == nil {
		return nil, ErrNilRegistry
	}
	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		if err := validateURLPattern(pattern); err != nil {
			return nil, err
		}
		normalized = appendMissingString(normalized, pattern)
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return nil, err
	}
	conflicts := make([]string, 0)
	for _, pattern := range normalized {
		owner := r.owner.servletMappings[pattern]
		if owner != "" && owner != r.name {
			conflicts = append(conflicts, pattern)
		}
	}
	if len(conflicts) > 0 {
		return conflicts, nil
	}
	for _, pattern := range normalized {
		r.mappings = appendMissingString(r.mappings, pattern)
		r.owner.servletMappings[pattern] = r.name
	}
	return nil, nil
}

// Mappings 返回当前 Servlet URL 映射副本。
func (r *ServletRegistration) Mappings() []string {
	if r == nil || r.owner == nil {
		return nil
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return cloneStrings(r.mappings)
}

// SetLoadOnStartup 设置容器启动时初始化顺序。
func (r *ServletRegistration) SetLoadOnStartup(order int) error {
	if r == nil || r.owner == nil {
		return ErrNilRegistry
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return err
	}
	r.loadOnStartup = order
	r.hasLoadOnStartup = true
	return nil
}

// LoadOnStartup 返回启动初始化顺序。
func (r *ServletRegistration) LoadOnStartup() (int, bool) {
	if r == nil || r.owner == nil {
		return 0, false
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.loadOnStartup, r.hasLoadOnStartup
}

// SetRunAsRole 设置 Servlet 执行身份角色。
func (r *ServletRegistration) SetRunAsRole(role string) error {
	if r == nil || r.owner == nil {
		return ErrNilRegistry
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return err
	}
	r.runAsRole = role
	return nil
}

// RunAsRole 返回 Servlet 执行身份角色。
func (r *ServletRegistration) RunAsRole() string {
	if r == nil || r.owner == nil {
		return ""
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.runAsRole
}

// SetMultipartConfig 设置 Servlet multipart 解析配置。
func (r *ServletRegistration) SetMultipartConfig(config multipart.Config) error {
	if r == nil || r.owner == nil {
		return ErrNilRegistry
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return err
	}
	r.multipartConfig = &config
	return nil
}

// MultipartConfig 返回 Servlet multipart 解析配置。
func (r *ServletRegistration) MultipartConfig() (multipart.Config, bool) {
	if r == nil || r.owner == nil {
		return multipart.Config{}, false
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	if r.multipartConfig == nil {
		return multipart.Config{}, false
	}
	return *r.multipartConfig, true
}

// SetSecurityConfig 设置 Servlet 声明式安全约束。
func (r *ServletRegistration) SetSecurityConfig(config security.Constraint) error {
	if r == nil || r.owner == nil {
		return ErrNilRegistry
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return err
	}
	r.securityConfig = &config
	return nil
}

// SecurityConfig 返回 Servlet 声明式安全约束。
func (r *ServletRegistration) SecurityConfig() (security.Constraint, bool) {
	if r == nil || r.owner == nil {
		return security.Constraint{}, false
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	if r.securityConfig == nil {
		return security.Constraint{}, false
	}
	return *r.securityConfig, true
}

func validateURLPattern(pattern string) error {
	router := servlet.NewRouter()
	return router.Handle(pattern, servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return nil
	}))
}
