package session

import (
	"context"
	"errors"
)

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

func (m *MemoryManager) fireSessionDestroyed(ctx context.Context, target Session) error {
	event := Event{Session: target}
	var result error
	listeners := m.listenerSnapshot()
	for i := len(listeners) - 1; i >= 0; i-- {
		result = errors.Join(result, listeners[i].SessionDestroyed(ctx, event))
	}
	return result
}

func (m *MemoryManager) fireSessionIDChanged(ctx context.Context, target Session, oldID, newID string) error {
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
