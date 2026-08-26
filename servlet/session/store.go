package session

import (
	"context"
	"sync"
	"time"
)

// Record 是会话可持久化快照。
type Record struct {
	ID                  string
	CreationTime        time.Time
	LastAccessedTime    time.Time
	MaxInactiveInterval time.Duration
	Attributes          map[string]any
}

// Store 定义会话持久化或分布式复制边界。
type Store interface {
	Load(ctx context.Context, id string) (Record, bool, error)
	Save(ctx context.Context, record Record) error
	Delete(ctx context.Context, id string) error
	Rename(ctx context.Context, oldID, newID string, record Record) error
}

// Snapshot 创建会话持久化快照。
func Snapshot(target Session) (Record, error) {
	if target == nil || !target.IsValid() {
		return Record{}, ErrInvalidSession
	}
	attributes := make(map[string]any)
	for _, name := range target.AttributeNames() {
		value, ok := target.Attribute(name)
		if ok {
			attributes[name] = value
		}
	}
	return Record{
		ID:                  target.ID(),
		CreationTime:        target.CreationTime(),
		LastAccessedTime:    target.LastAccessedTime(),
		MaxInactiveInterval: target.MaxInactiveInterval(),
		Attributes:          attributes,
	}, nil
}

func cloneRecord(record Record) Record {
	record.Attributes = cloneAttributes(record.Attributes)
	return record
}

func cloneAttributes(src map[string]any) map[string]any {
	if len(src) == 0 {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

// MemoryStore 是面向测试和单进程容器的 Store 实现。
type MemoryStore struct {
	mu      sync.RWMutex
	records map[string]Record
}

// NewMemoryStore 创建内存 Store。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: make(map[string]Record)}
}

// Load 读取会话快照。
func (s *MemoryStore) Load(ctx context.Context, id string) (Record, bool, error) {
	if err := ctx.Err(); err != nil {
		return Record{}, false, err
	}
	if s == nil {
		return Record{}, false, ErrNilStore
	}
	s.mu.RLock()
	record, ok := s.records[id]
	s.mu.RUnlock()
	if !ok {
		return Record{}, false, nil
	}
	return cloneRecord(record), true, nil
}

// Save 保存会话快照。
func (s *MemoryStore) Save(ctx context.Context, record Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil {
		return ErrNilStore
	}
	if record.ID == "" {
		return ErrInvalidSession
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[record.ID] = cloneRecord(record)
	return nil
}

// Delete 删除会话快照。
func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil {
		return ErrNilStore
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, id)
	return nil
}

// Rename 原子替换会话 ID。
func (s *MemoryStore) Rename(ctx context.Context, oldID, newID string, record Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil {
		return ErrNilStore
	}
	if oldID == "" || newID == "" || record.ID == "" {
		return ErrInvalidSession
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, oldID)
	record.ID = newID
	s.records[newID] = cloneRecord(record)
	return nil
}
