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
	listeners           []Listener
	attributeListeners  []AttributeListener
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
		session, exists := m.createWithIDLocked(id)
		if exists {
			continue
		}
		if err := m.fireSessionCreated(ctx, session); err != nil {
			_ = session.Invalidate()
			return nil, err
		}
		return session, nil
	}
}

// CreateWithID 使用容器提供的稳定 ID 创建会话。
func (m *MemoryManager) CreateWithID(ctx context.Context, id string) (Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, ErrSessionNotFound
	}
	session, exists := m.createWithIDLocked(id)
	if exists {
		return nil, ErrDuplicateSessionID
	}
	if err := m.fireSessionCreated(ctx, session); err != nil {
		_ = session.Invalidate()
		return nil, err
	}
	return session, nil
}

func (m *MemoryManager) createWithIDLocked(id string) (*memorySession, bool) {
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
	defer m.mu.Unlock()
	if _, exists := m.sessions[id]; exists {
		return nil, true
	}
	m.sessions[id] = session
	return session, false
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
		oldID := id
		delete(m.sessions, oldID)
		m.sessions[newID] = session
		m.mu.Unlock()
		if err := m.fireSessionIDChanged(ctx, session, oldID, newID); err != nil {
			return nil, err
		}
		return session, nil
	}
}

// Destroy 删除会话。
func (m *MemoryManager) Destroy(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	session := m.invalidateID(id)
	if session == nil {
		return nil
	}
	return m.fireSessionDestroyed(ctx, session)
}

// Passivate 将当前会话保存到 Store，并触发属性值钝化回调。
func (m *MemoryManager) Passivate(ctx context.Context, id string, store Store) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if store == nil {
		return ErrNilStore
	}
	session, ok, err := m.Get(ctx, id)
	if err != nil || !ok {
		return err
	}
	record, err := Snapshot(session)
	if err != nil {
		return err
	}
	fireSessionWillPassivate(session, record.Attributes)
	return store.Save(ctx, record)
}

// Activate 从 Store 恢复会话，并触发属性值激活回调。
func (m *MemoryManager) Activate(ctx context.Context, id string, store Store) (Session, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if store == nil {
		return nil, false, ErrNilStore
	}
	record, ok, err := store.Load(ctx, id)
	if err != nil || !ok {
		return nil, ok, err
	}
	now := m.clock()
	attributes := cloneAttributes(record.Attributes)
	session := &memorySession{
		manager:             m,
		id:                  record.ID,
		creationTime:        record.CreationTime,
		lastAccessedTime:    record.LastAccessedTime,
		maxInactiveInterval: record.MaxInactiveInterval,
		isNew:               false,
		valid:               true,
		attribute:           attributes,
	}
	if session.expired(now) {
		_ = store.Delete(ctx, id)
		return nil, false, nil
	}
	m.mu.Lock()
	if existing, exists := m.sessions[record.ID]; exists {
		m.mu.Unlock()
		return existing, true, nil
	}
	m.sessions[record.ID] = session
	m.mu.Unlock()
	fireSessionDidActivate(session, attributes)
	return session, true, nil
}

func (m *MemoryManager) invalidateSession(session *memorySession) error {
	m.mu.Lock()
	session.mu.Lock()
	if !session.valid {
		session.mu.Unlock()
		m.mu.Unlock()
		return ErrInvalidSession
	}
	session.valid = false
	delete(m.sessions, session.id)
	attributes := session.attribute
	session.attribute = make(map[string]any)
	session.mu.Unlock()
	m.mu.Unlock()
	m.fireAttributesRemoved(session, attributes)
	return m.fireSessionDestroyed(context.Background(), session)
}

func (m *MemoryManager) invalidateID(id string) *memorySession {
	m.mu.Lock()
	session, ok := m.sessions[id]
	if !ok {
		m.mu.Unlock()
		return nil
	}
	session.mu.Lock()
	session.valid = false
	attributes := session.attribute
	session.attribute = make(map[string]any)
	session.mu.Unlock()
	delete(m.sessions, id)
	m.mu.Unlock()
	m.fireAttributesRemoved(session, attributes)
	return session
}
