package registration

// ListenerKind 表示监听器标准类型。
type ListenerKind string

const (
	// ListenerContext 表示 ServletContextListener。
	ListenerContext ListenerKind = "context"
	// ListenerRequest 表示 ServletRequestListener。
	ListenerRequest ListenerKind = "request"
	// ListenerSession 表示 HttpSessionListener 与 HttpSessionIdListener。
	ListenerSession ListenerKind = "session"
)

// ListenerRegistration 保存单个 Listener 的动态注册信息。
type ListenerRegistration struct {
	owner     *Registry
	kind      ListenerKind
	className string
	listener  any
	order     int
}

// Kind 返回监听器类型。
func (r *ListenerRegistration) Kind() ListenerKind {
	if r == nil || r.owner == nil {
		return ""
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.kind
}

// ClassName 返回监听器实现类型名。
func (r *ListenerRegistration) ClassName() string {
	if r == nil || r.owner == nil {
		return ""
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.className
}

// SetClassName 设置监听器实现类型名。
func (r *ListenerRegistration) SetClassName(className string) error {
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

// Listener 返回注册的监听器实例。
func (r *ListenerRegistration) Listener() any {
	if r == nil || r.owner == nil {
		return nil
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.listener
}

// Order 返回监听器注册顺序。
func (r *ListenerRegistration) Order() int {
	if r == nil || r.owner == nil {
		return -1
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.order
}
