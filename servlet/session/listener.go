package session

import "context"

// Event 表示会话生命周期事件。
type Event struct {
	Session Session
}

// IDChangedEvent 表示会话 ID 轮换事件。
type IDChangedEvent struct {
	Session Session
	OldID   string
	NewID   string
}

// Listener 监听会话生命周期。
type Listener interface {
	SessionCreated(ctx context.Context, event Event) error
	SessionDestroyed(ctx context.Context, event Event) error
	SessionIDChanged(ctx context.Context, event IDChangedEvent) error
}

// ListenerFunc 将函数组适配为 Listener。
type ListenerFunc struct {
	Created   func(ctx context.Context, event Event) error
	Destroyed func(ctx context.Context, event Event) error
	IDChanged func(ctx context.Context, event IDChangedEvent) error
}

// SessionCreated 触发会话创建回调。
func (f ListenerFunc) SessionCreated(ctx context.Context, event Event) error {
	if f.Created == nil {
		return nil
	}
	return f.Created(ctx, event)
}

// SessionDestroyed 触发会话销毁回调。
func (f ListenerFunc) SessionDestroyed(ctx context.Context, event Event) error {
	if f.Destroyed == nil {
		return nil
	}
	return f.Destroyed(ctx, event)
}

// SessionIDChanged 触发会话 ID 轮换回调。
func (f ListenerFunc) SessionIDChanged(ctx context.Context, event IDChangedEvent) error {
	if f.IDChanged == nil {
		return nil
	}
	return f.IDChanged(ctx, event)
}
