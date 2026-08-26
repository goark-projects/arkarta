package servlet

import "context"

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
