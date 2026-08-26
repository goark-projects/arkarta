package async

import (
	"context"

	"goark.dev/arkarta/servlet"
)

// Event 表示异步请求生命周期事件。
type Event struct {
	Request  *servlet.Request
	Response servlet.Response
	Err      error
}

// Listener 监听异步请求生命周期。
type Listener interface {
	OnStartAsync(ctx context.Context, event Event)
	OnComplete(ctx context.Context, event Event)
	OnTimeout(ctx context.Context, event Event)
	OnError(ctx context.Context, event Event)
}

// ListenerFunc 将函数组适配为 Listener。
type ListenerFunc struct {
	Start    func(ctx context.Context, event Event)
	Complete func(ctx context.Context, event Event)
	Timeout  func(ctx context.Context, event Event)
	Error    func(ctx context.Context, event Event)
}

// OnStartAsync 触发异步开始回调。
func (f ListenerFunc) OnStartAsync(ctx context.Context, event Event) {
	if f.Start != nil {
		f.Start(ctx, event)
	}
}

// OnComplete 触发异步完成回调。
func (f ListenerFunc) OnComplete(ctx context.Context, event Event) {
	if f.Complete != nil {
		f.Complete(ctx, event)
	}
}

// OnTimeout 触发异步超时回调。
func (f ListenerFunc) OnTimeout(ctx context.Context, event Event) {
	if f.Timeout != nil {
		f.Timeout(ctx, event)
	}
}

// OnError 触发异步错误回调。
func (f ListenerFunc) OnError(ctx context.Context, event Event) {
	if f.Error != nil {
		f.Error(ctx, event)
	}
}
