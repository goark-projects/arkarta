package frame

import "strconv"

// OpCode 表示 WebSocket 帧操作码。
type OpCode byte

const (
	// OpContinuation 表示延续帧。
	OpContinuation OpCode = 0x0
	// OpText 表示文本数据帧。
	OpText OpCode = 0x1
	// OpBinary 表示二进制数据帧。
	OpBinary OpCode = 0x2
	// OpClose 表示关闭控制帧。
	OpClose OpCode = 0x8
	// OpPing 表示 Ping 控制帧。
	OpPing OpCode = 0x9
	// OpPong 表示 Pong 控制帧。
	OpPong OpCode = 0xA
)

// Valid 判断操作码是否属于 RFC 6455 已定义集合。
func (o OpCode) Valid() bool {
	return o == OpContinuation || o == OpText || o == OpBinary || o == OpClose || o == OpPing || o == OpPong
}

// Control 判断操作码是否为控制帧。
func (o OpCode) Control() bool {
	return o == OpClose || o == OpPing || o == OpPong
}

// Data 判断操作码是否为数据帧。
func (o OpCode) Data() bool {
	return o == OpText || o == OpBinary
}

// String 返回操作码名称。
func (o OpCode) String() string {
	switch o {
	case OpContinuation:
		return "continuation"
	case OpText:
		return "text"
	case OpBinary:
		return "binary"
	case OpClose:
		return "close"
	case OpPing:
		return "ping"
	case OpPong:
		return "pong"
	default:
		return "opcode(" + strconv.Itoa(int(o)) + ")"
	}
}
