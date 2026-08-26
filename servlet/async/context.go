package async

import (
	"context"
	"sync"
	"time"

	"goark.dev/arkarta/servlet"
)

// Option 定制异步上下文。
type Option func(*Context)

// WithTimeout 设置异步超时。
func WithTimeout(timeout time.Duration) Option {
	return func(async *Context) {
		async.timeout = timeout
	}
}

// WithListener 添加异步生命周期监听器。
func WithListener(listener Listener) Option {
	return func(async *Context) {
		if listener != nil {
			async.listeners = append(async.listeners, listener)
		}
	}
}

// Context 表示一次显式异步请求生命周期。
type Context struct {
	ctx    context.Context
	cancel context.CancelFunc

	req *servlet.Request
	res servlet.Response

	timeout   time.Duration
	timer     *time.Timer
	done      chan struct{}
	listeners []Listener

	mu        sync.RWMutex
	completed bool
	err       error
}

// NewContext 创建异步上下文。
func NewContext(parent context.Context, req *servlet.Request, res servlet.Response, options ...Option) (*Context, error) {
	if parent == nil {
		parent = context.Background()
	}
	if req == nil {
		return nil, ErrNilRequest
	}
	if res == nil {
		return nil, ErrNilResponse
	}
	ctx, cancel := context.WithCancel(parent)
	async := &Context{
		ctx:    ctx,
		cancel: cancel,
		req:    req,
		res:    res,
		done:   make(chan struct{}),
	}
	for _, option := range options {
		if option != nil {
			option(async)
		}
	}
	async.fireStart()
	if async.timeout > 0 {
		async.timer = time.AfterFunc(async.timeout, func() {
			async.fireTimeout()
			_ = async.Complete(ErrTimeout)
		})
	}
	return async, nil
}

// Go 在独立 goroutine 中执行任务并在结束时完成异步上下文。
func (a *Context) Go(fn func(context.Context) error) {
	go func() {
		var err error
		if fn != nil {
			err = fn(a.Context())
		}
		if err != nil {
			a.fireError(err)
		}
		_ = a.Complete(err)
	}()
}

// Dispatch 以 ASYNC 分发类型执行处理器。
func (a *Context) Dispatch(handler servlet.Handler) error {
	if handler == nil {
		return ErrNilHandler
	}
	if err := a.ensureActive(); err != nil {
		return err
	}
	return a.req.RunWithDispatchType(servlet.DispatchAsync, func() error {
		return handler.Serve(a.Context(), a.req, a.res)
	})
}

// Complete 标记异步请求完成。
func (a *Context) Complete(err error) error {
	a.mu.Lock()
	if a.completed {
		a.mu.Unlock()
		return ErrCompleted
	}
	a.completed = true
	a.err = err
	timer := a.timer
	a.mu.Unlock()

	if timer != nil {
		timer.Stop()
	}
	a.cancel()
	close(a.done)
	a.fireComplete(err)
	return nil
}

// Context 返回异步生命周期上下文。
func (a *Context) Context() context.Context {
	return a.ctx
}

// Done 返回完成通知通道。
func (a *Context) Done() <-chan struct{} {
	return a.done
}

// Err 返回异步任务错误。
func (a *Context) Err() error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.err
}

// Request 返回异步请求视图。
func (a *Context) Request() *servlet.Request {
	return a.req
}

// Response 返回异步响应。
func (a *Context) Response() servlet.Response {
	return a.res
}

func (a *Context) ensureActive() error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.completed {
		return ErrCompleted
	}
	return a.ctx.Err()
}

func (a *Context) fireStart() {
	event := Event{Request: a.req, Response: a.res}
	for _, listener := range a.listeners {
		listener.OnStartAsync(a.ctx, event)
	}
}

func (a *Context) fireComplete(err error) {
	event := Event{Request: a.req, Response: a.res, Err: err}
	for _, listener := range a.listeners {
		listener.OnComplete(a.ctx, event)
	}
}

func (a *Context) fireTimeout() {
	event := Event{Request: a.req, Response: a.res, Err: ErrTimeout}
	for _, listener := range a.listeners {
		listener.OnTimeout(a.ctx, event)
	}
}

func (a *Context) fireError(err error) {
	event := Event{Request: a.req, Response: a.res, Err: err}
	for _, listener := range a.listeners {
		listener.OnError(a.ctx, event)
	}
}
