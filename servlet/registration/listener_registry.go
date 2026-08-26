package registration

import (
	"goark.dev/arkarta/servlet"
)

// AddContextListener 注册上下文生命周期监听器。
func (r *Registry) AddContextListener(listener servlet.ContextListener) (*ListenerRegistration, error) {
	return r.addListener(ListenerContext, listener)
}

// AddRequestListener 注册请求生命周期监听器。
func (r *Registry) AddRequestListener(listener servlet.RequestListener) (*ListenerRegistration, error) {
	return r.addListener(ListenerRequest, listener)
}

// AddSessionListener 注册会话生命周期监听器。
func (r *Registry) AddSessionListener(listener any) (*ListenerRegistration, error) {
	if !isSessionListener(listener) {
		return nil, ErrNilListener
	}
	return r.addListener(ListenerSession, listener)
}

// AddListener 按监听器接口类型注册实例；多接口监听器应使用显式方法注册。
func (r *Registry) AddListener(listener any) (*ListenerRegistration, error) {
	switch item := listener.(type) {
	case servlet.ContextListener:
		return r.AddContextListener(item)
	case servlet.RequestListener:
		return r.AddRequestListener(item)
	default:
		if isSessionListener(item) {
			return r.AddSessionListener(item)
		}
		return nil, ErrNilListener
	}
}

func (r *Registry) addListener(kind ListenerKind, listener any) (*ListenerRegistration, error) {
	if isNil(listener) {
		return nil, ErrNilListener
	}
	if r == nil {
		return nil, ErrNilRegistry
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.ensureMutableLocked(); err != nil {
		return nil, err
	}
	item := &ListenerRegistration{
		owner:     r,
		kind:      kind,
		className: typeName(listener),
		listener:  listener,
		order:     len(r.listeners),
	}
	r.listeners = append(r.listeners, item)
	return item, nil
}
