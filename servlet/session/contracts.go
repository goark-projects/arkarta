// Package session 定义 Servlet Session Profile 的会话契约与默认实现。
package session

import (
	"context"
	"errors"
	"time"

	"goark.dev/arkarta/servlet"
)

// ErrInvalidSession 表示会话已经失效。
var ErrInvalidSession = errors.New("arkarta/servlet/session: session is invalid")

// ErrSessionNotFound 表示指定会话不存在。
var ErrSessionNotFound = errors.New("arkarta/servlet/session: session not found")

// ErrNilManager 表示会话访问器缺少 Manager。
var ErrNilManager = errors.New("arkarta/servlet/session: manager is nil")

// ErrNilRequest 表示会话访问器缺少请求对象。
var ErrNilRequest = errors.New("arkarta/servlet/session: request is nil")

// ErrInvalidCookieConfig 表示会话 Cookie 配置非法。
var ErrInvalidCookieConfig = errors.New("arkarta/servlet/session: invalid cookie config")

// ErrCookieConfigLocked 表示应用已经启动，不能再修改 Session Cookie 配置。
var ErrCookieConfigLocked = errors.New("arkarta/servlet/session: cookie config is locked")

// ErrInvalidURLRewriteConfig 表示 URL 重写配置非法。
var ErrInvalidURLRewriteConfig = errors.New("arkarta/servlet/session: invalid url rewrite config")

// ErrNilStore 表示会话持久化存储为空。
var ErrNilStore = errors.New("arkarta/servlet/session: store is nil")

// ErrDuplicateSessionID 表示指定会话 ID 已存在。
var ErrDuplicateSessionID = errors.New("arkarta/servlet/session: duplicate session id")

// Manager 管理 Servlet 会话生命周期。
type Manager interface {
	Create(ctx context.Context) (Session, error)
	Get(ctx context.Context, id string) (Session, bool, error)
	RenewID(ctx context.Context, id string) (Session, error)
	Destroy(ctx context.Context, id string) error
}

// IDBoundManager 支持使用容器提供的稳定 ID 创建会话。
type IDBoundManager interface {
	Manager
	CreateWithID(ctx context.Context, id string) (Session, error)
}

// Current 返回当前请求已经关联的 Session。
func Current(req *servlet.Request) (Session, bool) {
	if req == nil {
		return nil, false
	}
	value, ok := req.Attribute(AttributeCurrentSession)
	if !ok {
		return nil, false
	}
	current, ok := value.(Session)
	if !ok || current == nil || !current.IsValid() {
		req.SetAttribute(AttributeCurrentSession, nil)
		return nil, false
	}
	return current, true
}

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
