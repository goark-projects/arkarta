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

// AddContextAttributeListener 注册上下文属性监听器。
func (r *Registry) AddContextAttributeListener(listener servlet.ContextAttributeListener) (*ListenerRegistration, error) {
	return r.addListener(ListenerContextAttribute, listener)
}

// AddRequestAttributeListener 注册请求属性监听器。
func (r *Registry) AddRequestAttributeListener(listener servlet.RequestAttributeListener) (*ListenerRegistration, error) {
	return r.addListener(ListenerRequestAttribute, listener)
}

// AddSessionAttributeListener 注册会话属性监听器。
func (r *Registry) AddSessionAttributeListener(listener any) (*ListenerRegistration, error) {
	if !isSessionAttributeListener(listener) {
		return nil, ErrNilListener
	}
	return r.addListener(ListenerSessionAttribute, listener)
}

// AddListener 按监听器接口类型注册实例；多接口监听器应使用显式方法注册。
func (r *Registry) AddListener(listener any) (*ListenerRegistration, error) {
	switch item := listener.(type) {
	case servlet.ContextListener:
		return r.AddContextListener(item)
	case servlet.RequestListener:
		return r.AddRequestListener(item)
	case servlet.ContextAttributeListener:
		return r.AddContextAttributeListener(item)
	case servlet.RequestAttributeListener:
		return r.AddRequestAttributeListener(item)
	default:
		if isSessionListener(item) {
			return r.AddSessionListener(item)
		}
		if isSessionAttributeListener(item) {
			return r.AddSessionAttributeListener(item)
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
