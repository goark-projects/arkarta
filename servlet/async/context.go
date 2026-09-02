package async

import (
	"context"
	"errors"
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
	quiescent chan struct{}
	listeners []Listener

	mu         sync.RWMutex
	completed  bool
	err        error
	dispatches int
	workers    int
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
	quiescent := make(chan struct{})
	close(quiescent)
	async := &Context{
		ctx:       ctx,
		cancel:    cancel,
		req:       req,
		res:       res,
		done:      make(chan struct{}),
		quiescent: quiescent,
	}
	for _, option := range options {
		if option != nil {
			option(async)
		}
	}
	async.fireStart()
	if async.timeout > 0 {
		timer := time.AfterFunc(async.timeout, func() {
			_ = async.complete(ErrTimeout, true)
		})
		async.mu.Lock()
		if async.completed {
			timer.Stop()
		} else {
			async.timer = timer
		}
		async.mu.Unlock()
	}
	return async, nil
}

// Go 在独立 goroutine 中执行任务并在结束时完成异步上下文。
func (a *Context) Go(fn func(context.Context) error) error {
	if a == nil {
		return ErrCompleted
	}
	a.mu.Lock()
	if a.completed {
		a.mu.Unlock()
		return ErrCompleted
	}
	if a.workers == 0 {
		a.quiescent = make(chan struct{})
	}
	a.workers++
	a.mu.Unlock()

	go func() {
		defer a.workerDone()
		var err error
		if fn != nil {
			err = fn(a.Context())
		}
		_ = a.Complete(err)
	}()
	return nil
}

// Dispatch 以 ASYNC 分发类型执行处理器。
func (a *Context) Dispatch(handler servlet.Handler) error {
	if handler == nil {
		return ErrNilHandler
	}
	if err := a.markDispatch(); err != nil {
		return err
	}
	err := a.req.RunWithDispatchType(servlet.DispatchAsync, func() error {
		return handler.Serve(a.Context(), a.req, a.res)
	})
	if err != nil {
		a.fireError(err)
	}
	return err
}

// Complete 标记异步请求完成。
func (a *Context) Complete(err error) error {
	return a.complete(err, false)
}

// Await 等待异步请求完成，并返回完成时记录的错误。
func (a *Context) Await(ctx context.Context) error {
	if a == nil {
		return ErrCompleted
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-a.done:
		return a.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// AwaitQuiescence 等待已登记的异步任务全部退出。
func (a *Context) AwaitQuiescence(ctx context.Context) error {
	if a == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	a.mu.RLock()
	quiescent := a.quiescent
	a.mu.RUnlock()
	select {
	case <-quiescent:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Completed 判断异步请求是否已经完成。
func (a *Context) Completed() bool {
	if a == nil {
		return true
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.completed
}

// DispatchCount 返回该异步上下文已经执行的 ASYNC 分发次数。
func (a *Context) DispatchCount() int {
	if a == nil {
		return 0
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.dispatches
}

func (a *Context) complete(err error, timeout bool) error {
	a.mu.Lock()
	if a.completed {
		a.mu.Unlock()
		return ErrCompleted
	}
	a.completed = true
	a.err = err
	timer := a.timer
	a.mu.Unlock()

	if timer != nil && !timeout {
		timer.Stop()
	}
	if timeout {
		a.fireTimeout()
	} else if err != nil && !errors.Is(err, context.Canceled) {
		a.fireError(err)
	}
	a.fireComplete(err)
	a.cancel()
	close(a.done)
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

func (a *Context) markDispatch() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.completed {
		return ErrCompleted
	}
	if err := a.ctx.Err(); err != nil {
		return err
	}
	a.dispatches++
	return nil
}

func (a *Context) workerDone() {
	a.mu.Lock()
	a.workers--
	if a.workers == 0 {
		close(a.quiescent)
	}
	a.mu.Unlock()
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
