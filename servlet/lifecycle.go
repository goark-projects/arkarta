package servlet

import (
	"context"
	"errors"
	"fmt"
)

// ErrInvalidWebAppState 表示 Web 应用生命周期状态不允许当前操作。
var ErrInvalidWebAppState = errors.New("arkarta/servlet: invalid web app state")

// WebAppState 表示 Web 应用生命周期状态。
type WebAppState uint8

const (
	// WebAppStateNew 表示应用尚未初始化。
	WebAppStateNew WebAppState = iota
	// WebAppStateInitialized 表示应用已完成初始化。
	WebAppStateInitialized
	// WebAppStateStarted 表示应用已启动并可接收请求。
	WebAppStateStarted
	// WebAppStateStopped 表示应用已停止接收请求。
	WebAppStateStopped
	// WebAppStateDestroyed 表示应用已销毁。
	WebAppStateDestroyed
)

// State 返回当前生命周期状态。
func (a *WebApp) State() WebAppState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// Initialize 初始化 Web 应用上下文并触发 ContextListener。
func (a *WebApp) Initialize(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	listeners, err := a.transitionToInitialized()
	if err != nil {
		return err
	}
	event := ContextEvent{WebApp: a}
	for _, listener := range listeners {
		if err := listener.ContextInitialized(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Start 启动 Web 应用。
func (a *WebApp) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	switch a.state {
	case WebAppStateInitialized, WebAppStateStopped:
		a.state = WebAppStateStarted
		return nil
	case WebAppStateStarted:
		return nil
	default:
		return fmt.Errorf("%w: cannot start from %v", ErrInvalidWebAppState, a.state)
	}
}

// Stop 停止 Web 应用接收新请求。
func (a *WebApp) Stop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	switch a.state {
	case WebAppStateStarted:
		a.state = WebAppStateStopped
		return nil
	case WebAppStateStopped, WebAppStateDestroyed:
		return nil
	default:
		return fmt.Errorf("%w: cannot stop from %v", ErrInvalidWebAppState, a.state)
	}
}

// Destroy 销毁 Web 应用上下文并触发 ContextListener。
func (a *WebApp) Destroy(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	listeners, fire, err := a.transitionToDestroyed()
	if err != nil {
		return err
	}
	if !fire {
		return nil
	}
	event := ContextEvent{WebApp: a}
	var result error
	for i := len(listeners) - 1; i >= 0; i-- {
		result = errors.Join(result, listeners[i].ContextDestroyed(ctx, event))
	}
	return result
}

// RequestInitialized 触发请求初始化事件。
func (a *WebApp) RequestInitialized(ctx context.Context, req *Request) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	event := RequestEvent{WebApp: a, Request: req}
	for _, listener := range a.requestListenerSnapshot() {
		if err := listener.RequestInitialized(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// RequestDestroyed 触发请求销毁事件。
func (a *WebApp) RequestDestroyed(ctx context.Context, req *Request, cause error) error {
	event := RequestEvent{WebApp: a, Request: req, Err: cause}
	var result error
	listeners := a.requestListenerSnapshot()
	for i := len(listeners) - 1; i >= 0; i-- {
		result = errors.Join(result, listeners[i].RequestDestroyed(ctx, event))
	}
	return result
}

func (a *WebApp) transitionToInitialized() ([]ContextListener, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch a.state {
	case WebAppStateNew:
		a.state = WebAppStateInitialized
		return cloneContextListeners(a.contextListeners), nil
	case WebAppStateInitialized, WebAppStateStarted:
		return nil, nil
	default:
		return nil, fmt.Errorf("%w: cannot initialize from %v", ErrInvalidWebAppState, a.state)
	}
}

func (a *WebApp) transitionToDestroyed() ([]ContextListener, bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch a.state {
	case WebAppStateDestroyed:
		return nil, false, nil
	case WebAppStateNew:
		a.state = WebAppStateDestroyed
		return nil, false, nil
	case WebAppStateInitialized, WebAppStateStarted, WebAppStateStopped:
		a.state = WebAppStateDestroyed
		return cloneContextListeners(a.contextListeners), true, nil
	default:
		return nil, false, fmt.Errorf("%w: cannot destroy from %v", ErrInvalidWebAppState, a.state)
	}
}

func (a *WebApp) requestListenerSnapshot() []RequestListener {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return cloneRequestListeners(a.requestListeners)
}

func cloneContextListeners(src []ContextListener) []ContextListener {
	if len(src) == 0 {
		return nil
	}
	dst := make([]ContextListener, len(src))
	copy(dst, src)
	return dst
}

func cloneRequestListeners(src []RequestListener) []RequestListener {
	if len(src) == 0 {
		return nil
	}
	dst := make([]RequestListener, len(src))
	copy(dst, src)
	return dst
}

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
