package servlet

import (
	"context"

	"goark.dev/arkarta/servlet/upgrade"
	"goark.dev/arkarta/websocket"
)

// EndpointHandler 将 WebSocket Endpoint 适配为 Servlet Upgrade Handler。
func EndpointHandler(sessionID string, endpoint websocket.Endpoint, options ...FrameConnectionOption) Handler {
	return HandlerFunc(func(ctx context.Context, handshake websocket.Handshake, conn upgrade.Connection) error {
		return ServeEndpoint(ctx, sessionID, handshake, conn, endpoint, options...)
	})
}

// ServeEndpoint 基于升级连接运行标准 WebSocket Endpoint。
func ServeEndpoint(ctx context.Context, sessionID string, handshake websocket.Handshake, conn upgrade.Connection, endpoint websocket.Endpoint, options ...FrameConnectionOption) error {
	if endpoint == nil {
		return websocket.ErrNilEndpoint
	}
	frameConn, err := NewFrameConnection(conn, options...)
	if err != nil {
		return err
	}
	session, err := websocket.NewSession(sessionID, frameConn, websocket.WithSubprotocol(handshake.Subprotocol()))
	if err != nil {
		return err
	}
	return websocket.Serve(ctx, session, endpoint)
}
