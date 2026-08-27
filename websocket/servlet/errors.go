package servlet

import "errors"

// ErrNilHandler 表示 WebSocket 升级后处理器为空。
var ErrNilHandler = errors.New("arkarta/websocket/servlet: handler is nil")

// ErrNilConnection 表示 WebSocket 升级后连接为空。
var ErrNilConnection = errors.New("arkarta/websocket/servlet: upgrade connection is nil")
