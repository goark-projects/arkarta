package servlet

import (
	"context"

	"goark.dev/arkarta/servlet/upgrade"
	"goark.dev/arkarta/websocket"
)

// Handler 处理 Servlet Upgrade 完成后的 WebSocket 连接。
type Handler interface {
	ServeWebSocket(ctx context.Context, handshake websocket.Handshake, conn upgrade.Connection) error
}

// HandlerFunc 将函数适配为 WebSocket 升级处理器。
type HandlerFunc func(ctx context.Context, handshake websocket.Handshake, conn upgrade.Connection) error

// ServeWebSocket 执行 WebSocket 升级处理函数。
func (f HandlerFunc) ServeWebSocket(ctx context.Context, handshake websocket.Handshake, conn upgrade.Connection) error {
	return f(ctx, handshake, conn)
}
