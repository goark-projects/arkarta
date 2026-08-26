package upgrade

import (
	"bufio"
	"context"
	"net"
	"time"
)

// Connection 表示升级后由应用接管的网络连接。
type Connection interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// BufferedConnection 表示带缓冲读写器的升级连接。
type BufferedConnection interface {
	Connection
	Reader() *bufio.Reader
	Writer() *bufio.ReadWriter
}

// Handler 处理升级后的连接生命周期。
type Handler interface {
	ServeUpgrade(ctx context.Context, conn Connection) error
}

// HandlerFunc 将函数适配为 Handler。
type HandlerFunc func(ctx context.Context, conn Connection) error

// ServeUpgrade 执行协议升级处理函数。
func (f HandlerFunc) ServeUpgrade(ctx context.Context, conn Connection) error {
	return f(ctx, conn)
}
