package websocket

import "errors"

// ErrNilConnection 表示 WebSocket 会话缺少底层连接。
var ErrNilConnection = errors.New("arkarta/websocket: connection is nil")

// ErrNilEndpoint 表示 WebSocket 会话缺少 Endpoint。
var ErrNilEndpoint = errors.New("arkarta/websocket: endpoint is nil")

// ErrSessionClosed 表示 WebSocket 会话已经关闭。
var ErrSessionClosed = errors.New("arkarta/websocket: session is closed")

// ErrNilHTTPRequest 表示 HTTP 握手请求为空。
var ErrNilHTTPRequest = errors.New("arkarta/websocket: http request is nil")

// ErrNilResponseWriter 表示 HTTP 握手响应写出器为空。
var ErrNilResponseWriter = errors.New("arkarta/websocket: response writer is nil")

// ErrInvalidHandshake 表示 WebSocket HTTP 握手非法。
var ErrInvalidHandshake = errors.New("arkarta/websocket: invalid handshake")

// ErrUnsupportedVersion 表示客户端 WebSocket 版本不受支持。
var ErrUnsupportedVersion = errors.New("arkarta/websocket: unsupported websocket version")

// ErrMessageTooLarge 表示解压后的消息超过限制。
var ErrMessageTooLarge = errors.New("arkarta/websocket: message too large")
