package websocket

import "context"

// Connection 是容器提供给 WebSocket 标准层的连接端口。
type Connection interface {
	Read(ctx context.Context) (Message, error)
	Write(ctx context.Context, message Message) error
	Close(ctx context.Context, reason CloseReason) error
}
