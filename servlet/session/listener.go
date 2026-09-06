package session

import (
	"context"
	"errors"
)

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
func (f ListenerFunc) SessionIDChanged(
	ctx context.Context,
	event IDChangedEvent,
) error {
	if f.IDChanged == nil {
		return nil
	}
	return f.IDChanged(ctx, event)
}

func (m *MemoryManager) listenerSnapshot() []Listener {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.listeners) == 0 {
		return nil
	}
	result := make([]Listener, len(m.listeners))
	copy(result, m.listeners)
	return result
}

func (m *MemoryManager) fireSessionCreated(ctx context.Context, target Session) error {
	event := Event{Session: target}
	var result error
	for _, listener := range m.listenerSnapshot() {
		result = errors.Join(result, listener.SessionCreated(ctx, event))
	}
	return result
}

func (m *MemoryManager) fireSessionDestroyed(
	ctx context.Context,
	target Session,
) error {
	event := Event{Session: target}
	var result error
	listeners := m.listenerSnapshot()
	for i := len(listeners) - 1; i >= 0; i-- {
		result = errors.Join(result, listeners[i].SessionDestroyed(ctx, event))
	}
	return result
}

func (m *MemoryManager) fireSessionIDChanged(
	ctx context.Context,
	target Session,
	oldID, newID string,
) error {
	event := IDChangedEvent{
		Session: target,
		OldID:   oldID,
		NewID:   newID,
	}
	var result error
	for _, listener := range m.listenerSnapshot() {
		result = errors.Join(result, listener.SessionIDChanged(ctx, event))
	}
	return result
}
