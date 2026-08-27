package websocket

// MessageType 表示 WebSocket 消息类型。
type MessageType uint8

const (
	MessageText MessageType = iota + 1
	MessageBinary
	MessagePing
	MessagePong
	MessageClose
)

// Message 表示容器与 Endpoint 之间传递的消息。
type Message struct {
	typ    MessageType
	data   []byte
	reason CloseReason
}

// TextMessage 创建文本消息。
func TextMessage(text string) Message {
	return Message{typ: MessageText, data: []byte(text)}
}

// BinaryMessage 创建二进制消息。
func BinaryMessage(data []byte) Message {
	return Message{typ: MessageBinary, data: cloneBytes(data)}
}

// PingMessage 创建 Ping 消息。
func PingMessage(data []byte) Message {
	return Message{typ: MessagePing, data: cloneBytes(data)}
}

// PongMessage 创建 Pong 消息。
func PongMessage(data []byte) Message {
	return Message{typ: MessagePong, data: cloneBytes(data)}
}

// CloseMessage 创建关闭消息。
func CloseMessage(reason CloseReason) Message {
	return Message{typ: MessageClose, reason: reason}
}

// Type 返回消息类型。
func (m Message) Type() MessageType {
	return m.typ
}

// Text 返回文本内容。
func (m Message) Text() string {
	return string(m.data)
}

// Binary 返回二进制内容副本。
func (m Message) Binary() []byte {
	return cloneBytes(m.data)
}

// CloseReason 返回关闭原因。
func (m Message) CloseReason() CloseReason {
	return m.reason
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
