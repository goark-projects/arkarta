package container

import (
	"context"

	"goark.dev/arkarta/servlet"
)

// Container 是 Web 容器必须实现的入口。
type Container interface {
	Metadata() Metadata
	Deploy(ctx context.Context, deployment *Deployment) (Application, error)
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// Application 表示一个已部署应用。
type Application interface {
	WebApp() *servlet.WebApp
	Handler() servlet.Handler
	Stop(ctx context.Context) error
}
