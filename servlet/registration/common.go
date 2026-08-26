package registration

import "sort"

type commonRegistration struct {
	owner          *Registry
	name           string
	className      string
	initParam      map[string]string
	asyncSupported bool
}

func newCommonRegistration(owner *Registry, name string, target any) commonRegistration {
	return commonRegistration{
		owner:     owner,
		name:      name,
		className: typeName(target),
		initParam: make(map[string]string),
	}
}

// Name 返回注册项名称。
func (r *commonRegistration) Name() string {
	if r == nil || r.owner == nil {
		return ""
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.name
}

// ClassName 返回实现类型名；Go 实例注册时使用包路径加类型名。
func (r *commonRegistration) ClassName() string {
	if r == nil || r.owner == nil {
		return ""
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.className
}

// SetClassName 设置实现类型名；空字符串用于表达仅声明但未绑定类型。
func (r *commonRegistration) SetClassName(className string) error {
	if r == nil || r.owner == nil {
		return ErrNilRegistry
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return err
	}
	r.className = className
	return nil
}

// SetAsyncSupported 设置注册项是否声明支持异步处理。
func (r *commonRegistration) SetAsyncSupported(asyncSupported bool) error {
	if r == nil || r.owner == nil {
		return ErrNilRegistry
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return err
	}
	r.asyncSupported = asyncSupported
	return nil
}

// AsyncSupported 返回注册项是否声明支持异步处理。
func (r *commonRegistration) AsyncSupported() bool {
	if r == nil || r.owner == nil {
		return false
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.asyncSupported
}

// SetInitParam 设置单个初始化参数；返回 false 表示参数名已存在。
func (r *commonRegistration) SetInitParam(name, value string) (bool, error) {
	if name == "" {
		return false, ErrInvalidInitParamName
	}
	if r == nil || r.owner == nil {
		return false, ErrNilRegistry
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return false, err
	}
	if _, exists := r.initParam[name]; exists {
		return false, nil
	}
	r.initParam[name] = value
	return true, nil
}

// SetInitParams 批量设置初始化参数；存在冲突时不写入任何参数。
func (r *commonRegistration) SetInitParams(params map[string]string) ([]string, error) {
	for name := range params {
		if name == "" {
			return nil, ErrInvalidInitParamName
		}
	}
	if r == nil || r.owner == nil {
		return nil, ErrNilRegistry
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return nil, err
	}
	conflicts := make([]string, 0)
	for name := range params {
		if _, exists := r.initParam[name]; exists {
			conflicts = append(conflicts, name)
		}
	}
	sort.Strings(conflicts)
	if len(conflicts) > 0 {
		return conflicts, nil
	}
	for name, value := range params {
		r.initParam[name] = value
	}
	return nil, nil
}

// InitParam 返回指定初始化参数。
func (r *commonRegistration) InitParam(name string) (string, bool) {
	if r == nil || r.owner == nil {
		return "", false
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	value, ok := r.initParam[name]
	return value, ok
}

// InitParams 返回初始化参数副本。
func (r *commonRegistration) InitParams() map[string]string {
	if r == nil || r.owner == nil {
		return map[string]string{}
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return cloneStringMap(r.initParam)
}
