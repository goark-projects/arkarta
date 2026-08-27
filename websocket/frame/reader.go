package frame

import (
	"encoding/binary"
	"io"
)

const defaultMaxPayloadBytes int64 = 32 << 20

// MaskPolicy 表示读取方向的 Mask 校验策略。
type MaskPolicy uint8

const (
	// MaskOptional 表示允许 Mask 或非 Mask 帧。
	MaskOptional MaskPolicy = iota
	// MaskRequired 表示要求读取客户端到服务端的 Mask 帧。
	MaskRequired
	// MaskForbidden 表示要求读取服务端到客户端的非 Mask 帧。
	MaskForbidden
)

// ReaderOption 定制帧读取器。
type ReaderOption func(*readerConfig)

type readerConfig struct {
	maxPayloadBytes int64
	maskPolicy      MaskPolicy
	allowReserved   bool
}

// WithMaxPayloadBytes 设置单帧最大载荷；负数表示不限制。
func WithMaxPayloadBytes(maxBytes int64) ReaderOption {
	return func(config *readerConfig) {
		config.maxPayloadBytes = maxBytes
	}
}

// WithMaskPolicy 设置读取方向的 Mask 策略。
func WithMaskPolicy(policy MaskPolicy) ReaderOption {
	return func(config *readerConfig) {
		config.maskPolicy = policy
	}
}

// WithReservedBits 设置是否允许扩展协商后的 RSV 位。
func WithReservedBits(enabled bool) ReaderOption {
	return func(config *readerConfig) {
		config.allowReserved = enabled
	}
}

// Reader 从字节流读取 WebSocket 帧。
type Reader struct {
	reader io.Reader
	config readerConfig
}

// NewReader 创建帧读取器。
func NewReader(reader io.Reader, options ...ReaderOption) *Reader {
	config := readerConfig{maxPayloadBytes: defaultMaxPayloadBytes}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	return &Reader{reader: reader, config: config}
}

// Read 读取单个 WebSocket 帧。
func Read(reader io.Reader, options ...ReaderOption) (Frame, error) {
	return NewReader(reader, options...).ReadFrame()
}

// ReadFrame 读取单个 WebSocket 帧。
func (r *Reader) ReadFrame() (Frame, error) {
	if r == nil || r.reader == nil {
		return Frame{}, ErrNilReader
	}
	header := [2]byte{}
	if _, err := io.ReadFull(r.reader, header[:]); err != nil {
		return Frame{}, err
	}

	opcode := OpCode(header[0] & 0x0f)
	masked := header[1]&0x80 != 0
	payloadLength := uint64(header[1] & 0x7f)
	switch payloadLength {
	case 126:
		var extended [2]byte
		if _, err := io.ReadFull(r.reader, extended[:]); err != nil {
			return Frame{}, err
		}
		payloadLength = uint64(binary.BigEndian.Uint16(extended[:]))
		if payloadLength < 126 {
			return Frame{}, ErrProtocol
		}
	case 127:
		var extended [8]byte
		if _, err := io.ReadFull(r.reader, extended[:]); err != nil {
			return Frame{}, err
		}
		payloadLength = binary.BigEndian.Uint64(extended[:])
		if payloadLength < 1<<16 || payloadLength&(1<<63) != 0 {
			return Frame{}, ErrProtocol
		}
	}
	if err := r.validateMask(masked); err != nil {
		return Frame{}, err
	}
	if err := r.validateLength(payloadLength); err != nil {
		return Frame{}, err
	}

	var key MaskKey
	if masked {
		if _, err := io.ReadFull(r.reader, key[:]); err != nil {
			return Frame{}, err
		}
	}
	payload := make([]byte, int(payloadLength))
	if _, err := io.ReadFull(r.reader, payload); err != nil {
		return Frame{}, err
	}
	if masked {
		ApplyMask(payload, key)
	}
	frame := Frame{
		fin:     header[0]&0x80 != 0,
		rsv1:    header[0]&0x40 != 0,
		rsv2:    header[0]&0x20 != 0,
		rsv3:    header[0]&0x10 != 0,
		opcode:  opcode,
		payload: payload,
		masked:  masked,
		maskKey: key,
	}
	if err := validateFrame(frame, r.config.allowReserved); err != nil {
		return Frame{}, err
	}
	return frame, nil
}

func (r *Reader) validateMask(masked bool) error {
	switch r.config.maskPolicy {
	case MaskRequired:
		if !masked {
			return ErrMaskRequired
		}
	case MaskForbidden:
		if masked {
			return ErrMaskForbidden
		}
	}
	return nil
}

func (r *Reader) validateLength(length uint64) error {
	if r.config.maxPayloadBytes >= 0 && length > uint64(r.config.maxPayloadBytes) {
		return ErrPayloadTooLarge
	}
	if length > uint64(int(^uint(0)>>1)) {
		return ErrPayloadTooLarge
	}
	return nil
}
