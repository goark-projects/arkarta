package frame

// Frame 表示一个完整 WebSocket 帧。
type Frame struct {
	fin     bool
	rsv1    bool
	rsv2    bool
	rsv3    bool
	opcode  OpCode
	payload []byte
	masked  bool
	maskKey MaskKey
}

// FrameOption 定制帧元数据。
type FrameOption func(*Frame)

// New 创建 WebSocket 帧。
func New(opcode OpCode, payload []byte, options ...FrameOption) Frame {
	frame := Frame{
		fin:     true,
		opcode:  opcode,
		payload: cloneBytes(payload),
	}
	for _, option := range options {
		if option != nil {
			option(&frame)
		}
	}
	return frame
}

// WithFin 设置 FIN 标记。
func WithFin(fin bool) FrameOption {
	return func(frame *Frame) {
		frame.fin = fin
	}
}

// WithRSV1 设置 RSV1 标记。
func WithRSV1(enabled bool) FrameOption {
	return func(frame *Frame) {
		frame.rsv1 = enabled
	}
}

// WithRSV2 设置 RSV2 标记。
func WithRSV2(enabled bool) FrameOption {
	return func(frame *Frame) {
		frame.rsv2 = enabled
	}
}

// WithRSV3 设置 RSV3 标记。
func WithRSV3(enabled bool) FrameOption {
	return func(frame *Frame) {
		frame.rsv3 = enabled
	}
}

// WithMask 设置客户端 Mask。
func WithMask(key MaskKey) FrameOption {
	return func(frame *Frame) {
		frame.masked = true
		frame.maskKey = key
	}
}

// Fin 返回 FIN 标记。
func (f Frame) Fin() bool {
	return f.fin
}

// RSV1 返回 RSV1 标记。
func (f Frame) RSV1() bool {
	return f.rsv1
}

// RSV2 返回 RSV2 标记。
func (f Frame) RSV2() bool {
	return f.rsv2
}

// RSV3 返回 RSV3 标记。
func (f Frame) RSV3() bool {
	return f.rsv3
}

// OpCode 返回帧操作码。
func (f Frame) OpCode() OpCode {
	return f.opcode
}

// Payload 返回帧载荷副本。
func (f Frame) Payload() []byte {
	return cloneBytes(f.payload)
}

// Masked 表示帧是否携带客户端 Mask。
func (f Frame) Masked() bool {
	return f.masked
}

// MaskKey 返回帧 MaskKey。
func (f Frame) MaskKey() MaskKey {
	return f.maskKey
}

func validateFrame(frame Frame, allowReservedBits bool) error {
	if !frame.opcode.Valid() {
		return ErrInvalidOpcode
	}
	if !allowReservedBits && (frame.rsv1 || frame.rsv2 || frame.rsv3) {
		return ErrProtocol
	}
	if frame.opcode.Control() {
		if !frame.fin {
			return ErrProtocol
		}
		if len(frame.payload) > 125 {
			return ErrProtocol
		}
		if frame.rsv1 || frame.rsv2 || frame.rsv3 {
			return ErrProtocol
		}
		if frame.opcode == OpClose {
			if _, _, err := ParseClosePayload(frame.payload); err != nil {
				return err
			}
		}
	}
	return nil
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
