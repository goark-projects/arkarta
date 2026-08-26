package session

import (
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

func (s *memorySession) SetAttribute(name string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.valid {
		return ErrInvalidSession
	}
	if value == nil {
		delete(s.attribute, name)
		return nil
	}
	s.attribute[name] = value
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
