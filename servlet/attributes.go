package servlet

import (
	"context"
)

// ContextAttributeEvent 表示 WebApp 属性变更事件。
type ContextAttributeEvent struct {
	WebApp   *WebApp
	Name     string
	Value    any
	OldValue any
}

// RequestAttributeEvent 表示 Request 属性变更事件。
type RequestAttributeEvent struct {
	Request  *Request
	Name     string
	Value    any
	OldValue any
}

// ContextAttributeListener 监听 WebApp 属性变更。
type ContextAttributeListener interface {
	AttributeAdded(ctx context.Context, event ContextAttributeEvent)
	AttributeReplaced(ctx context.Context, event ContextAttributeEvent)
	AttributeRemoved(ctx context.Context, event ContextAttributeEvent)
}

// RequestAttributeListener 监听 Request 属性变更。
type RequestAttributeListener interface {
	AttributeAdded(ctx context.Context, event RequestAttributeEvent)
	AttributeReplaced(ctx context.Context, event RequestAttributeEvent)
	AttributeRemoved(ctx context.Context, event RequestAttributeEvent)
}

// ContextAttributeListenerFunc 将函数组适配为 ContextAttributeListener。
type ContextAttributeListenerFunc struct {
	Added    func(ctx context.Context, event ContextAttributeEvent)
	Replaced func(ctx context.Context, event ContextAttributeEvent)
	Removed  func(ctx context.Context, event ContextAttributeEvent)
}

// AttributeAdded 触发属性新增回调。
func (f ContextAttributeListenerFunc) AttributeAdded(ctx context.Context, event ContextAttributeEvent) {
	if f.Added != nil {
		f.Added(ctx, event)
	}
}

// AttributeReplaced 触发属性替换回调。
func (f ContextAttributeListenerFunc) AttributeReplaced(ctx context.Context, event ContextAttributeEvent) {
	if f.Replaced != nil {
		f.Replaced(ctx, event)
	}
}

// AttributeRemoved 触发属性移除回调。
func (f ContextAttributeListenerFunc) AttributeRemoved(ctx context.Context, event ContextAttributeEvent) {
	if f.Removed != nil {
		f.Removed(ctx, event)
	}
}

// RequestAttributeListenerFunc 将函数组适配为 RequestAttributeListener。
type RequestAttributeListenerFunc struct {
	Added    func(ctx context.Context, event RequestAttributeEvent)
	Replaced func(ctx context.Context, event RequestAttributeEvent)
	Removed  func(ctx context.Context, event RequestAttributeEvent)
}

// AttributeAdded 触发属性新增回调。
func (f RequestAttributeListenerFunc) AttributeAdded(ctx context.Context, event RequestAttributeEvent) {
	if f.Added != nil {
		f.Added(ctx, event)
	}
}

// AttributeReplaced 触发属性替换回调。
func (f RequestAttributeListenerFunc) AttributeReplaced(ctx context.Context, event RequestAttributeEvent) {
	if f.Replaced != nil {
		f.Replaced(ctx, event)
	}
}

// AttributeRemoved 触发属性移除回调。
func (f RequestAttributeListenerFunc) AttributeRemoved(ctx context.Context, event RequestAttributeEvent) {
	if f.Removed != nil {
		f.Removed(ctx, event)
	}
}

// WithRequestAttributeListener 设置请求属性监听器。
func WithRequestAttributeListener(listener RequestAttributeListener) RequestOption {
	return func(req *Request) {
		if listener != nil {
			req.attributeListeners = append(req.attributeListeners, listener)
		}
	}
}

// AddRequestAttributeListener 为已创建请求追加属性监听器。
func (r *Request) AddRequestAttributeListener(listener RequestAttributeListener) {
	if listener == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.attributeListeners = append(r.attributeListeners, listener)
}

func (r *Request) setAttribute(ctx context.Context, key string, value any) {
	r.mu.Lock()
	oldValue, existed := r.attribute[key]
	switch {
	case value == nil:
		delete(r.attribute, key)
	case existed:
		r.attribute[key] = value
	default:
		r.attribute[key] = value
	}
	listeners := append([]RequestAttributeListener(nil), r.attributeListeners...)
	r.mu.Unlock()

	if value == nil && existed {
		fireRequestAttributeRemoved(ctx, listeners, r, key, oldValue)
		return
	}
	if value != nil && existed {
		fireRequestAttributeReplaced(ctx, listeners, r, key, value, oldValue)
		return
	}
	if value != nil {
		fireRequestAttributeAdded(ctx, listeners, r, key, value)
	}
}

func fireRequestAttributeAdded(ctx context.Context, listeners []RequestAttributeListener, req *Request, name string, value any) {
	event := RequestAttributeEvent{Request: req, Name: name, Value: value}
	for _, listener := range listeners {
		listener.AttributeAdded(ctx, event)
	}
}

func fireRequestAttributeReplaced(ctx context.Context, listeners []RequestAttributeListener, req *Request, name string, value, oldValue any) {
	event := RequestAttributeEvent{Request: req, Name: name, Value: value, OldValue: oldValue}
	for _, listener := range listeners {
		listener.AttributeReplaced(ctx, event)
	}
}

func fireRequestAttributeRemoved(ctx context.Context, listeners []RequestAttributeListener, req *Request, name string, oldValue any) {
	event := RequestAttributeEvent{Request: req, Name: name, OldValue: oldValue}
	for _, listener := range listeners {
		listener.AttributeRemoved(ctx, event)
	}
}

// WithContextAttributeListener 添加应用上下文属性监听器。
func WithContextAttributeListener(listener ContextAttributeListener) WebAppOption {
	return func(app *WebApp) error {
		if listener != nil {
			app.contextAttributeListeners = append(app.contextAttributeListeners, listener)
		}
		return nil
	}
}

// WithRequestAttributeListener 添加请求属性监听器。
func WithWebAppRequestAttributeListener(listener RequestAttributeListener) WebAppOption {
	return func(app *WebApp) error {
		if listener != nil {
			app.requestAttributeListeners = append(app.requestAttributeListeners, listener)
		}
		return nil
	}
}

// AddContextAttributeListener 在应用初始化前追加上下文属性监听器。
func (a *WebApp) AddContextAttributeListener(listener ContextAttributeListener) error {
	if listener == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state != WebAppStateNew {
		return ErrInvalidWebAppState
	}
	a.contextAttributeListeners = append(a.contextAttributeListeners, listener)
	return nil
}

// AddRequestAttributeListener 在应用初始化前追加请求属性监听器。
func (a *WebApp) AddRequestAttributeListener(listener RequestAttributeListener) error {
	if listener == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state != WebAppStateNew {
		return ErrInvalidWebAppState
	}
	a.requestAttributeListeners = append(a.requestAttributeListeners, listener)
	return nil
}

// AttachRequestAttributeListeners 将应用级请求属性监听器挂到单个请求。
func (a *WebApp) AttachRequestAttributeListeners(req *Request) {
	if a == nil || req == nil {
		return
	}
	a.mu.RLock()
	listeners := append([]RequestAttributeListener(nil), a.requestAttributeListeners...)
	a.mu.RUnlock()
	for _, listener := range listeners {
		req.AddRequestAttributeListener(listener)
	}
}

// SetAttributeContext 设置应用属性并触发属性事件。
func (a *WebApp) SetAttributeContext(ctx context.Context, key string, value any) {
	a.mu.Lock()
	oldValue, existed := a.attribute[key]
	switch {
	case value == nil:
		delete(a.attribute, key)
	case existed:
		a.attribute[key] = value
	default:
		a.attribute[key] = value
	}
	listeners := append([]ContextAttributeListener(nil), a.contextAttributeListeners...)
	a.mu.Unlock()

	if value == nil && existed {
		fireContextAttributeRemoved(ctx, listeners, a, key, oldValue)
		return
	}
	if value != nil && existed {
		fireContextAttributeReplaced(ctx, listeners, a, key, value, oldValue)
		return
	}
	if value != nil {
		fireContextAttributeAdded(ctx, listeners, a, key, value)
	}
}

func fireContextAttributeAdded(ctx context.Context, listeners []ContextAttributeListener, app *WebApp, name string, value any) {
	event := ContextAttributeEvent{WebApp: app, Name: name, Value: value}
	for _, listener := range listeners {
		listener.AttributeAdded(ctx, event)
	}
}

func fireContextAttributeReplaced(ctx context.Context, listeners []ContextAttributeListener, app *WebApp, name string, value, oldValue any) {
	event := ContextAttributeEvent{WebApp: app, Name: name, Value: value, OldValue: oldValue}
	for _, listener := range listeners {
		listener.AttributeReplaced(ctx, event)
	}
}

func fireContextAttributeRemoved(ctx context.Context, listeners []ContextAttributeListener, app *WebApp, name string, oldValue any) {
	event := ContextAttributeEvent{WebApp: app, Name: name, OldValue: oldValue}
	for _, listener := range listeners {
		listener.AttributeRemoved(ctx, event)
	}
}
