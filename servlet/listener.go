package servlet

import "context"

// ContextEvent 表示 Web 应用上下文事件。
type ContextEvent struct {
	WebApp *WebApp
}

// RequestEvent 表示一次请求的生命周期事件。
type RequestEvent struct {
	WebApp  *WebApp
	Request *Request
	Err     error
}

// ContextListener 监听 Web 应用上下文生命周期。
type ContextListener interface {
	ContextInitialized(ctx context.Context, event ContextEvent) error
	ContextDestroyed(ctx context.Context, event ContextEvent) error
}

// ContextListenerFunc 将函数组适配为 ContextListener。
type ContextListenerFunc struct {
	Initialized func(ctx context.Context, event ContextEvent) error
	Destroyed   func(ctx context.Context, event ContextEvent) error
}

// ContextInitialized 触发上下文初始化回调。
func (f ContextListenerFunc) ContextInitialized(ctx context.Context, event ContextEvent) error {
	if f.Initialized == nil {
		return nil
	}
	return f.Initialized(ctx, event)
}

// ContextDestroyed 触发上下文销毁回调。
func (f ContextListenerFunc) ContextDestroyed(ctx context.Context, event ContextEvent) error {
	if f.Destroyed == nil {
		return nil
	}
	return f.Destroyed(ctx, event)
}

// RequestListener 监听请求生命周期。
type RequestListener interface {
	RequestInitialized(ctx context.Context, event RequestEvent) error
	RequestDestroyed(ctx context.Context, event RequestEvent) error
}

// RequestListenerFunc 将函数组适配为 RequestListener。
type RequestListenerFunc struct {
	Initialized func(ctx context.Context, event RequestEvent) error
	Destroyed   func(ctx context.Context, event RequestEvent) error
}

// RequestInitialized 触发请求初始化回调。
func (f RequestListenerFunc) RequestInitialized(ctx context.Context, event RequestEvent) error {
	if f.Initialized == nil {
		return nil
	}
	return f.Initialized(ctx, event)
}

// RequestDestroyed 触发请求销毁回调。
func (f RequestListenerFunc) RequestDestroyed(ctx context.Context, event RequestEvent) error {
	if f.Destroyed == nil {
		return nil
	}
	return f.Destroyed(ctx, event)
}
