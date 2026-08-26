package websocket

import (
	"context"
	"sync"
)

// Session 表示一个 WebSocket 会话。
type Session interface {
	ID() string
	Subprotocol() string
	Attribute(name string) (any, bool)
	SetAttribute(name string, value any)
	SendText(ctx context.Context, text string) error
	SendBinary(ctx context.Context, data []byte) error
	Close(ctx context.Context, reason CloseReason) error
	Closed() bool
}

// SessionOption 定制 WebSocket 会话。
type SessionOption func(*StandardSession)

// StandardSession 是标准 Session 实现。
type StandardSession struct {
	id          string
	subprotocol string
	connection  Connection

	mu        sync.RWMutex
	attribute map[string]any
	closed    bool
}

// NewSession 创建标准 WebSocket 会话。
func NewSession(id string, connection Connection, options ...SessionOption) (*StandardSession, error) {
	if connection == nil {
		return nil, ErrNilConnection
	}
	session := &StandardSession{
		id:         id,
		connection: connection,
		attribute:  make(map[string]any),
	}
	for _, option := range options {
		if option != nil {
			option(session)
		}
	}
	return session, nil
}

// WithSubprotocol 设置协商后的子协议。
func WithSubprotocol(subprotocol string) SessionOption {
	return func(session *StandardSession) {
		session.subprotocol = subprotocol
	}
}

// WithAttribute 设置初始用户属性。
func WithAttribute(name string, value any) SessionOption {
	return func(session *StandardSession) {
		if name != "" && value != nil {
			session.attribute[name] = value
		}
	}
}

// ID 返回会话 ID。
func (s *StandardSession) ID() string {
	if s == nil {
		return ""
	}
	return s.id
}

// Subprotocol 返回协商后的子协议。
func (s *StandardSession) Subprotocol() string {
	if s == nil {
		return ""
	}
	return s.subprotocol
}

// Attribute 返回用户属性。
func (s *StandardSession) Attribute(name string) (any, bool) {
	if s == nil {
		return nil, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.attribute[name]
	return value, ok
}

// SetAttribute 设置用户属性；nil 表示删除。
func (s *StandardSession) SetAttribute(name string, value any) {
	if s == nil || name == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if value == nil {
		delete(s.attribute, name)
		return
	}
	s.attribute[name] = value
}

// SendText 发送文本消息。
func (s *StandardSession) SendText(ctx context.Context, text string) error {
	return s.write(ctx, TextMessage(text))
}

// SendBinary 发送二进制消息。
func (s *StandardSession) SendBinary(ctx context.Context, data []byte) error {
	return s.write(ctx, BinaryMessage(data))
}

func (s *StandardSession) write(ctx context.Context, message Message) error {
	if s == nil || s.connection == nil {
		return ErrNilConnection
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSessionClosed
	}
	return s.connection.Write(ctx, message)
}

// Close 关闭会话。
func (s *StandardSession) Close(ctx context.Context, reason CloseReason) error {
	if s == nil || s.connection == nil {
		return ErrNilConnection
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSessionClosed
	}
	s.closed = true
	return s.connection.Close(ctx, reason)
}

// Closed 判断会话是否已经关闭。
func (s *StandardSession) Closed() bool {
	if s == nil {
		return true
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

func (s *StandardSession) markClosed() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
}
