package websocket

import (
	"context"
	"errors"
	"io"
)

// Endpoint 处理一个 WebSocket 会话的生命周期与消息。
type Endpoint interface {
	OnOpen(ctx context.Context, session Session) error
	OnText(ctx context.Context, session Session, text string) error
	OnBinary(ctx context.Context, session Session, data []byte) error
	OnPong(ctx context.Context, session Session, data []byte) error
	OnClose(ctx context.Context, session Session, reason CloseReason) error
	OnError(ctx context.Context, session Session, err error)
}

// EndpointFunc 将函数组适配为 Endpoint。
type EndpointFunc struct {
	Open   func(ctx context.Context, session Session) error
	Text   func(ctx context.Context, session Session, text string) error
	Binary func(ctx context.Context, session Session, data []byte) error
	Pong   func(ctx context.Context, session Session, data []byte) error
	Close  func(ctx context.Context, session Session, reason CloseReason) error
	Error  func(ctx context.Context, session Session, err error)
}

func (f EndpointFunc) OnOpen(ctx context.Context, session Session) error {
	if f.Open == nil {
		return nil
	}
	return f.Open(ctx, session)
}

func (f EndpointFunc) OnText(ctx context.Context, session Session, text string) error {
	if f.Text == nil {
		return nil
	}
	return f.Text(ctx, session, text)
}

func (f EndpointFunc) OnBinary(ctx context.Context, session Session, data []byte) error {
	if f.Binary == nil {
		return nil
	}
	return f.Binary(ctx, session, data)
}

func (f EndpointFunc) OnPong(ctx context.Context, session Session, data []byte) error {
	if f.Pong == nil {
		return nil
	}
	return f.Pong(ctx, session, data)
}

func (f EndpointFunc) OnClose(ctx context.Context, session Session, reason CloseReason) error {
	if f.Close == nil {
		return nil
	}
	return f.Close(ctx, session, reason)
}

func (f EndpointFunc) OnError(ctx context.Context, session Session, err error) {
	if f.Error != nil {
		f.Error(ctx, session, err)
	}
}

// Serve 运行一个 WebSocket Endpoint 消息循环。
func Serve(ctx context.Context, session *StandardSession, endpoint Endpoint) error {
	if session == nil || session.connection == nil {
		return ErrNilConnection
	}
	if endpoint == nil {
		return ErrNilEndpoint
	}
	if err := endpoint.OnOpen(ctx, session); err != nil {
		endpoint.OnError(ctx, session, err)
		_ = session.Close(ctx, NewCloseReason(CloseUnexpectedCondition, "open failed"))
		return err
	}
	for {
		if err := ctx.Err(); err != nil {
			endpoint.OnError(ctx, session, err)
			_ = session.Close(context.Background(), NewCloseReason(CloseGoingAway, "context canceled"))
			return err
		}
		message, err := session.connection.Read(ctx)
		if errors.Is(err, io.EOF) {
			session.markClosed()
			return endpoint.OnClose(ctx, session, NewCloseReason(CloseAbnormal, "connection closed"))
		}
		if err != nil {
			endpoint.OnError(ctx, session, err)
			_ = session.Close(context.Background(), NewCloseReason(CloseUnexpectedCondition, "read failed"))
			return err
		}
		if message.Type() == MessageClose {
			session.markClosed()
			return endpoint.OnClose(ctx, session, message.CloseReason())
		}
		if err := dispatchMessage(ctx, session, endpoint, message); err != nil {
			endpoint.OnError(ctx, session, err)
			_ = session.Close(context.Background(), NewCloseReason(CloseUnexpectedCondition, "endpoint failed"))
			return err
		}
	}
}

func dispatchMessage(ctx context.Context, session Session, endpoint Endpoint, message Message) error {
	switch message.Type() {
	case MessageText:
		return endpoint.OnText(ctx, session, message.Text())
	case MessageBinary:
		return endpoint.OnBinary(ctx, session, message.Binary())
	case MessagePong:
		return endpoint.OnPong(ctx, session, message.Binary())
	default:
		return nil
	}
}
