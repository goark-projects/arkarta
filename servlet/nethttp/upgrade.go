package nethttp

import (
	"bufio"
	"context"
	"net"
	"net/http"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/upgrade"
)

// UpgradeHTTP 将标准库 HTTP 连接升级并交给协议处理器。
func (r *Response) UpgradeHTTP(ctx context.Context, req *servlet.Request, handler upgrade.Handler) error {
	if handler == nil {
		return upgrade.ErrNilHandler
	}
	if r.Committed() {
		return upgrade.ErrAlreadyCommitted
	}
	conn, rw, err := http.NewResponseController(r.writer).Hijack()
	if err != nil {
		return err
	}
	r.committed = true
	return handler.ServeUpgrade(ctx, upgradedConnection{Conn: conn, rw: rw})
}

type upgradedConnection struct {
	net.Conn
	rw *bufio.ReadWriter
}

func (c upgradedConnection) Reader() *bufio.Reader {
	if c.rw == nil {
		return nil
	}
	return c.rw.Reader
}

func (c upgradedConnection) Writer() *bufio.ReadWriter {
	return c.rw
}
