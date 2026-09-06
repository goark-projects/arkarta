package session

import (
	"sort"
	"sync"
	"time"
)

type memorySession struct {
	manager             *MemoryManager
	id                  string
	creationTime        time.Time
	lastAccessedTime    time.Time
	maxInactiveInterval time.Duration
	isNew               bool
	valid               bool
	attribute           map[string]any
	mu                  sync.RWMutex
}

func (s *memorySession) ID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.id
}

func (s *memorySession) CreationTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.creationTime
}

func (s *memorySession) LastAccessedTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastAccessedTime
}

func (s *memorySession) MaxInactiveInterval() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.maxInactiveInterval
}

func (s *memorySession) SetMaxInactiveInterval(interval time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.valid {
		return ErrInvalidSession
	}
	s.maxInactiveInterval = interval
	return nil
}

func (s *memorySession) IsNew() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isNew
}

func (s *memorySession) IsValid() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.valid
}

func (s *memorySession) Attribute(name string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.valid {
		return nil, false
	}
	value, ok := s.attribute[name]
	return value, ok
}

func (s *memorySession) AttributeNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.valid {
		return nil
	}
	names := make([]string, 0, len(s.attribute))
	for name := range s.attribute {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *memorySession) SetAttribute(name string, value any) error {
	s.mu.Lock()
	if !s.valid {
		s.mu.Unlock()
		return ErrInvalidSession
	}
	oldValue, existed := s.attribute[name]
	if value == nil {
		delete(s.attribute, name)
	} else {
		s.attribute[name] = value
	}
	s.mu.Unlock()

	if value == nil && existed {
		fireValueUnbound(s, name, oldValue)
		s.manager.fireAttributeRemoved(s, name, oldValue)
		return nil
	}
	if value != nil && existed {
		fireValueUnbound(s, name, oldValue)
		fireValueBound(s, name, value)
		s.manager.fireAttributeReplaced(s, name, value, oldValue)
		return nil
	}
	if value != nil {
		fireValueBound(s, name, value)
		s.manager.fireAttributeAdded(s, name, value)
	}
	return nil
}

func (s *memorySession) RemoveAttribute(name string) error {
	return s.SetAttribute(name, nil)
}

func (s *memorySession) Invalidate() error {
	return s.manager.invalidateSession(s)
}

func (s *memorySession) access(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.valid {
		return
	}
	s.lastAccessedTime = now
	s.isNew = false
}

func (s *memorySession) expired(now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.valid {
		return !s.valid
	}
	if s.maxInactiveInterval < 0 {
		return false
	}
	if s.maxInactiveInterval == 0 {
		return true
	}
	return now.Sub(s.lastAccessedTime) > s.maxInactiveInterval
}

func (s *memorySession) renewIDLockedByManager(oldID, newID string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.valid || s.id != oldID {
		return false
	}
	s.id = newID
	s.isNew = false
	s.lastAccessedTime = now
	return true
}

// MemoryManagerOption 定制内存会话管理器。
type MemoryManagerOption func(*MemoryManager)

// WithIDGenerator 设置会话 ID 生成器。
func WithIDGenerator(generator IDGenerator) MemoryManagerOption {
	return func(manager *MemoryManager) {
		if generator != nil {
			manager.idGenerator = generator
		}
	}
}

// WithClock 设置时间源，测试中用于控制过期行为。
func WithClock(clock func() time.Time) MemoryManagerOption {
	return func(manager *MemoryManager) {
		if clock != nil {
			manager.clock = clock
		}
	}
}

// WithMaxInactiveInterval 设置默认空闲超时。
func WithMaxInactiveInterval(interval time.Duration) MemoryManagerOption {
	return func(manager *MemoryManager) {
		manager.maxInactiveInterval = interval
	}
}

// WithListener 添加会话生命周期监听器。
func WithListener(listener Listener) MemoryManagerOption {
	return func(manager *MemoryManager) {
		if listener != nil {
			manager.listeners = append(manager.listeners, listener)
		}
	}
}

// WithAttributeListener 添加会话属性监听器。
func WithAttributeListener(listener AttributeListener) MemoryManagerOption {
	return func(manager *MemoryManager) {
		if listener != nil {
			manager.attributeListeners = append(manager.attributeListeners, listener)
		}
	}
}
