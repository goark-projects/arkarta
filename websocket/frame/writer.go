package frame

import (
	"encoding/binary"
	"io"
)

// Writer 向字节流写出 WebSocket 帧。
type Writer struct {
	writer io.Writer
}

// NewWriter 创建帧写出器。
func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

// Write 写出单个 WebSocket 帧。
func Write(writer io.Writer, frame Frame) error {
	return NewWriter(writer).WriteFrame(frame)
}

// WriteFrame 写出单个 WebSocket 帧。
func (w *Writer) WriteFrame(frame Frame) error {
	if w == nil || w.writer == nil {
		return ErrNilWriter
	}
	if err := validateFrame(frame, true); err != nil {
		return err
	}
	header := make([]byte, 0, 14+len(frame.payload))
	first := byte(frame.opcode)
	if frame.fin {
		first |= 0x80
	}
	if frame.rsv1 {
		first |= 0x40
	}
	if frame.rsv2 {
		first |= 0x20
	}
	if frame.rsv3 {
		first |= 0x10
	}
	header = append(header, first)

	maskBit := byte(0)
	if frame.masked {
		maskBit = 0x80
	}
	payloadLength := len(frame.payload)
	switch {
	case payloadLength <= 125:
		header = append(header, maskBit|byte(payloadLength))
	case payloadLength <= 0xffff:
		header = append(header, maskBit|126)
		var extended [2]byte
		binary.BigEndian.PutUint16(extended[:], uint16(payloadLength))
		header = append(header, extended[:]...)
	default:
		header = append(header, maskBit|127)
		var extended [8]byte
		binary.BigEndian.PutUint64(extended[:], uint64(payloadLength))
		header = append(header, extended[:]...)
	}
	if frame.masked {
		header = append(header, frame.maskKey[:]...)
	}
	if err := writeFull(w.writer, header); err != nil {
		return err
	}
	payload := frame.payload
	if frame.masked {
		payload = MaskPayload(payload, frame.maskKey)
	}
	return writeFull(w.writer, payload)
}

func writeFull(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := writer.Write(data)
		if err != nil {
			return err
		}
		if n <= 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}
