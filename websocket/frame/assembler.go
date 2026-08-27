package frame

// Message 表示帧层聚合后的 WebSocket 消息或控制帧。
type Message struct {
	opcode     OpCode
	payload    []byte
	compressed bool
}

// OpCode 返回消息操作码。
func (m Message) OpCode() OpCode {
	return m.opcode
}

// Payload 返回消息载荷副本。
func (m Message) Payload() []byte {
	return cloneBytes(m.payload)
}

// Compressed 表示消息首帧是否带 RSV1 压缩标记。
func (m Message) Compressed() bool {
	return m.compressed
}

// AssemblerOption 定制碎片聚合器。
type AssemblerOption func(*Assembler)

// WithMaxMessageBytes 设置聚合消息最大载荷；负数表示不限制。
func WithMaxMessageBytes(maxBytes int64) AssemblerOption {
	return func(assembler *Assembler) {
		assembler.maxMessageBytes = maxBytes
	}
}

// Assembler 聚合 WebSocket 碎片帧。
type Assembler struct {
	maxMessageBytes int64
	fragmented      bool
	opcode          OpCode
	compressed      bool
	payload         []byte
}

// NewAssembler 创建碎片聚合器。
func NewAssembler(options ...AssemblerOption) *Assembler {
	assembler := &Assembler{maxMessageBytes: defaultMaxPayloadBytes}
	for _, option := range options {
		if option != nil {
			option(assembler)
		}
	}
	return assembler
}

// Add 增加一个帧；complete 为 true 时返回完整消息或控制帧。
func (a *Assembler) Add(frame Frame) (Message, bool, error) {
	if a == nil {
		return Message{}, false, ErrProtocol
	}
	if err := validateFrame(frame, true); err != nil {
		return Message{}, false, err
	}
	if frame.opcode.Control() {
		return Message{opcode: frame.opcode, payload: cloneBytes(frame.payload)}, true, nil
	}
	switch frame.opcode {
	case OpText, OpBinary:
		return a.addData(frame)
	case OpContinuation:
		return a.addContinuation(frame)
	default:
		return Message{}, false, ErrInvalidOpcode
	}
}

func (a *Assembler) addData(frame Frame) (Message, bool, error) {
	if a.fragmented {
		return Message{}, false, ErrProtocol
	}
	if err := a.ensureMessageSize(int64(len(frame.payload))); err != nil {
		return Message{}, false, err
	}
	if frame.fin {
		return Message{opcode: frame.opcode, payload: cloneBytes(frame.payload), compressed: frame.rsv1}, true, nil
	}
	a.fragmented = true
	a.opcode = frame.opcode
	a.compressed = frame.rsv1
	a.payload = cloneBytes(frame.payload)
	return Message{}, false, nil
}

func (a *Assembler) addContinuation(frame Frame) (Message, bool, error) {
	if !a.fragmented || frame.rsv1 || frame.rsv2 || frame.rsv3 {
		return Message{}, false, ErrProtocol
	}
	if err := a.ensureMessageSize(int64(len(a.payload) + len(frame.payload))); err != nil {
		return Message{}, false, err
	}
	a.payload = append(a.payload, frame.payload...)
	if !frame.fin {
		return Message{}, false, nil
	}
	message := Message{opcode: a.opcode, payload: cloneBytes(a.payload), compressed: a.compressed}
	a.reset()
	return message, true, nil
}

func (a *Assembler) ensureMessageSize(size int64) error {
	if a.maxMessageBytes >= 0 && size > a.maxMessageBytes {
		return ErrPayloadTooLarge
	}
	return nil
}

func (a *Assembler) reset() {
	a.fragmented = false
	a.opcode = 0
	a.compressed = false
	a.payload = nil
}
