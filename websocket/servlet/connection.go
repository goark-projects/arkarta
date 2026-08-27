package servlet

import (
	"context"

	"goark.dev/arkarta/servlet/upgrade"
	"goark.dev/arkarta/websocket"
	"goark.dev/arkarta/websocket/frame"
)

// FrameConnection 将 Servlet Upgrade 连接适配为 WebSocket 消息连接。
type FrameConnection struct {
	conn      upgrade.Connection
	reader    *frame.Reader
	writer    *frame.Writer
	assembler *frame.Assembler
}

// FrameConnectionOption 定制帧连接适配器。
type FrameConnectionOption func(*frameConnectionConfig)

type frameConnectionConfig struct {
	maxFrameBytes   int64
	maxMessageBytes int64
}

// WithMaxFrameBytes 设置单帧最大载荷。
func WithMaxFrameBytes(maxBytes int64) FrameConnectionOption {
	return func(config *frameConnectionConfig) {
		config.maxFrameBytes = maxBytes
	}
}

// WithMaxMessageBytes 设置聚合消息最大载荷。
func WithMaxMessageBytes(maxBytes int64) FrameConnectionOption {
	return func(config *frameConnectionConfig) {
		config.maxMessageBytes = maxBytes
	}
}

// NewFrameConnection 创建基于 RFC 6455 帧层的 WebSocket 连接适配器。
func NewFrameConnection(conn upgrade.Connection, options ...FrameConnectionOption) (*FrameConnection, error) {
	if conn == nil {
		return nil, ErrNilConnection
	}
	config := frameConnectionConfig{
		maxFrameBytes:   32 << 20,
		maxMessageBytes: 32 << 20,
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	return &FrameConnection{
		conn: conn,
		reader: frame.NewReader(
			conn,
			frame.WithMaskPolicy(frame.MaskRequired),
			frame.WithMaxPayloadBytes(config.maxFrameBytes),
		),
		writer:    frame.NewWriter(conn),
		assembler: frame.NewAssembler(frame.WithMaxMessageBytes(config.maxMessageBytes)),
	}, nil
}

// Read 读取并聚合一个 WebSocket 消息。
func (c *FrameConnection) Read(ctx context.Context) (websocket.Message, error) {
	if c == nil || c.reader == nil || c.writer == nil || c.assembler == nil {
		return websocket.Message{}, ErrNilConnection
	}
	for {
		if err := ctx.Err(); err != nil {
			return websocket.Message{}, err
		}
		next, err := c.reader.ReadFrame()
		if err != nil {
			return websocket.Message{}, err
		}
		message, complete, err := c.assembler.Add(next)
		if err != nil {
			return websocket.Message{}, err
		}
		if !complete {
			continue
		}
		switch message.OpCode() {
		case frame.OpPing:
			if err := c.writeFrame(ctx, frame.New(frame.OpPong, message.Payload())); err != nil {
				return websocket.Message{}, err
			}
			continue
		case frame.OpPong:
			return websocket.PongMessage(message.Payload()), nil
		case frame.OpText:
			return websocket.TextMessage(string(message.Payload())), nil
		case frame.OpBinary:
			return websocket.BinaryMessage(message.Payload()), nil
		case frame.OpClose:
			code, reason, err := frame.ParseClosePayload(message.Payload())
			if err != nil {
				return websocket.Message{}, err
			}
			if code == 0 {
				return websocket.CloseMessage(websocket.NewCloseReason(websocket.CloseNoStatus, reason)), nil
			}
			return websocket.CloseMessage(websocket.NewCloseReason(websocket.CloseCode(code), reason)), nil
		default:
			return websocket.Message{}, frame.ErrInvalidOpcode
		}
	}
}

// Write 写出一个 WebSocket 消息。
func (c *FrameConnection) Write(ctx context.Context, message websocket.Message) error {
	if c == nil || c.writer == nil {
		return ErrNilConnection
	}
	switch message.Type() {
	case websocket.MessageText:
		return c.writeFrame(ctx, frame.New(frame.OpText, []byte(message.Text())))
	case websocket.MessageBinary:
		return c.writeFrame(ctx, frame.New(frame.OpBinary, message.Binary()))
	case websocket.MessagePing:
		return c.writeFrame(ctx, frame.New(frame.OpPing, message.Binary()))
	case websocket.MessagePong:
		return c.writeFrame(ctx, frame.New(frame.OpPong, message.Binary()))
	case websocket.MessageClose:
		return c.Close(ctx, message.CloseReason())
	default:
		return frame.ErrInvalidOpcode
	}
}

// Close 写出关闭帧并关闭底层连接。
func (c *FrameConnection) Close(ctx context.Context, reason websocket.CloseReason) error {
	if c == nil || c.writer == nil || c.conn == nil {
		return ErrNilConnection
	}
	payload, err := closePayload(reason)
	if err != nil {
		return err
	}
	writeErr := c.writeFrame(ctx, frame.New(frame.OpClose, payload))
	closeErr := c.conn.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func (c *FrameConnection) writeFrame(ctx context.Context, next frame.Frame) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.writer.WriteFrame(next)
}

func closePayload(reason websocket.CloseReason) ([]byte, error) {
	code := reason.Code()
	if code == websocket.CloseNoStatus || code == websocket.CloseAbnormal || code == websocket.CloseTLSFailure {
		if reason.Reason() == "" {
			return nil, nil
		}
		return nil, frame.ErrInvalidClosePayload
	}
	return frame.ClosePayload(uint16(code), reason.Reason())
}

var _ websocket.Connection = (*FrameConnection)(nil)
