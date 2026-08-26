package session

import "time"

// Session 表示 Servlet 会话。
type Session interface {
	ID() string
	CreationTime() time.Time
	LastAccessedTime() time.Time
	MaxInactiveInterval() time.Duration
	SetMaxInactiveInterval(interval time.Duration) error
	IsNew() bool
	IsValid() bool
	Attribute(name string) (any, bool)
	AttributeNames() []string
	SetAttribute(name string, value any) error
	RemoveAttribute(name string) error
	Invalidate() error
}
