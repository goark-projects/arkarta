package session

import "context"

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
