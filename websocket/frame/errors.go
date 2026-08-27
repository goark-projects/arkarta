package frame

import "errors"

// ErrNilReader 表示帧读取器为空。
var ErrNilReader = errors.New("arkarta/websocket/frame: reader is nil")

// ErrNilWriter 表示帧写出器为空。
var ErrNilWriter = errors.New("arkarta/websocket/frame: writer is nil")

// ErrProtocol 表示 WebSocket 帧违反 RFC 6455 协议约束。
var ErrProtocol = errors.New("arkarta/websocket/frame: protocol error")

// ErrInvalidOpcode 表示帧操作码非法。
var ErrInvalidOpcode = errors.New("arkarta/websocket/frame: invalid opcode")

// ErrMaskRequired 表示当前方向要求客户端 Mask。
var ErrMaskRequired = errors.New("arkarta/websocket/frame: mask required")

// ErrMaskForbidden 表示当前方向禁止服务端 Mask。
var ErrMaskForbidden = errors.New("arkarta/websocket/frame: mask forbidden")

// ErrPayloadTooLarge 表示帧或消息超过安全上限。
var ErrPayloadTooLarge = errors.New("arkarta/websocket/frame: payload too large")

// ErrInvalidClosePayload 表示关闭帧载荷非法。
var ErrInvalidClosePayload = errors.New("arkarta/websocket/frame: invalid close payload")
