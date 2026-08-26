package session

import (
	"context"
	"sync"
	"time"
)

const defaultMaxInactiveInterval = 30 * time.Minute

// MemoryManager 是面向测试和轻量容器的内存会话管理器。
type MemoryManager struct {
	mu                  sync.RWMutex
	sessions            map[string]*memorySession
	idGenerator         IDGenerator
	clock               func() time.Time
	maxInactiveInterval time.Duration
}

// NewMemoryManager 创建内存会话管理器。
func NewMemoryManager(options ...MemoryManagerOption) *MemoryManager {
	manager := &MemoryManager{
		sessions:            make(map[string]*memorySession),
		idGenerator:         SecureID,
		clock:               time.Now,
		maxInactiveInterval: defaultMaxInactiveInterval,
	}
	for _, option := range options {
		if option != nil {
			option(manager)
		}
	}
	return manager
}

// Create 创建新会话。
func (m *MemoryManager) Create(ctx context.Context) (Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	for {
		id, err := m.idGenerator()
		if err != nil {
			return nil, err
		}
		now := m.clock()
		session := &memorySession{
			manager:             m,
			id:                  id,
			creationTime:        now,
			lastAccessedTime:    now,
			maxInactiveInterval: m.maxInactiveInterval,
			isNew:               true,
			valid:               true,
			attribute:           make(map[string]any),
		}

		m.mu.Lock()
		if _, exists := m.sessions[id]; !exists {
			m.sessions[id] = session
			m.mu.Unlock()
			return session, nil
		}
		m.mu.Unlock()
	}
}

// Get 查找并访问会话。
func (m *MemoryManager) Get(ctx context.Context, id string) (Session, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return nil, false, nil
	}
	if session.expired(m.clock()) {
		_ = session.Invalidate()
		return nil, false, nil
	}
	session.access(m.clock())
	return session, true, nil
}

// RenewID 轮换会话 ID。
func (m *MemoryManager) RenewID(ctx context.Context, id string) (Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	for {
		newID, err := m.idGenerator()
		if err != nil {
			return nil, err
		}

		m.mu.Lock()
		session, ok := m.sessions[id]
		if !ok {
			m.mu.Unlock()
			return nil, ErrSessionNotFound
		}
		if _, exists := m.sessions[newID]; exists {
			m.mu.Unlock()
			continue
		}
		if !session.renewIDLockedByManager(id, newID, m.clock()) {
			m.mu.Unlock()
			return nil, ErrSessionNotFound
		}
		delete(m.sessions, id)
		m.sessions[newID] = session
		m.mu.Unlock()
		return session, nil
	}
}

// Destroy 删除会话。
func (m *MemoryManager) Destroy(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.invalidateID(id)
	return nil
}

func (m *MemoryManager) invalidateSession(session *memorySession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	session.mu.Lock()
	defer session.mu.Unlock()
	if !session.valid {
		return ErrInvalidSession
	}
	session.valid = false
	delete(m.sessions, session.id)
	session.attribute = make(map[string]any)
	return nil
}

func (m *MemoryManager) invalidateID(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	if !ok {
		return
	}
	session.mu.Lock()
	session.valid = false
	session.attribute = make(map[string]any)
	session.mu.Unlock()
	delete(m.sessions, id)
}
