package websocket

import "errors"

// ErrNilConnection 表示 WebSocket 会话缺少底层连接。
var ErrNilConnection = errors.New("arkarta/websocket: connection is nil")

// ErrNilEndpoint 表示 WebSocket 会话缺少 Endpoint。
var ErrNilEndpoint = errors.New("arkarta/websocket: endpoint is nil")

// ErrSessionClosed 表示 WebSocket 会话已经关闭。
var ErrSessionClosed = errors.New("arkarta/websocket: session is closed")
