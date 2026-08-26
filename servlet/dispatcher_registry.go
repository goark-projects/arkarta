package servlet

import "sync"

// DispatcherRegistry 基于 Router 提供路径和名称分发器。
type DispatcherRegistry struct {
	router *Router

	mu    sync.RWMutex
	names map[string]string
}

// NewDispatcherRegistry 创建分发器注册表。
func NewDispatcherRegistry(router *Router) *DispatcherRegistry {
	return &DispatcherRegistry{
		router: router,
		names:  make(map[string]string),
	}
}

// RegisterName 绑定 Servlet 名称到可分发路径。
func (r *DispatcherRegistry) RegisterName(name, path string) {
	if r == nil || name == "" || path == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.names[name] = path
}

// RequestDispatcher 返回指定路径的请求分发器。
func (r *DispatcherRegistry) RequestDispatcher(path string) (RequestDispatcher, error) {
	if r == nil || r.router == nil {
		return nil, ErrNilRouter
	}
	return NewRequestDispatcher(r.router, path)
}

// NamedDispatcher 返回指定 Servlet 名称的请求分发器。
func (r *DispatcherRegistry) NamedDispatcher(name string) (RequestDispatcher, error) {
	if r == nil || r.router == nil {
		return nil, ErrNilRouter
	}
	r.mu.RLock()
	path, ok := r.names[name]
	r.mu.RUnlock()
	if !ok {
		return nil, ErrDispatcherTargetNotFound
	}
	return NewRequestDispatcher(r.router, path)
}
