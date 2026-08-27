package nativeio

import (
	"context"
	"io"
	"sync"
)

const defaultBufferSize = 32 << 10

var defaultBufferPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, defaultBufferSize)
		return &buffer
	},
}

// StandardSender 是跨平台 Native I/O Profile 参考实现。
type StandardSender struct {
	bufferSize int
}

// Option 定制标准发送器。
type Option func(*StandardSender)

// NewStandardSender 创建标准发送器。
func NewStandardSender(options ...Option) *StandardSender {
	sender := &StandardSender{bufferSize: defaultBufferSize}
	for _, option := range options {
		if option != nil {
			option(sender)
		}
	}
	return sender
}

// WithBufferSize 设置 fallback 缓冲大小。
func WithBufferSize(size int) Option {
	return func(sender *StandardSender) {
		if size > 0 {
			sender.bufferSize = size
		}
	}
}

// SendFile 发送指定文件区段。
func (s *StandardSender) SendFile(ctx context.Context, dst io.Writer, region FileRegion) (SendResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return SendResult{}, err
	}
	if dst == nil {
		return SendResult{}, ErrNilWriter
	}
	if err := region.validate(); err != nil {
		return SendResult{}, err
	}
	if region.count == 0 {
		return NewSendResult(0, StrategyBufferedCopy), nil
	}

	reader := &contextReader{
		ctx:    ctx,
		reader: io.NewSectionReader(region.source, region.offset, region.count),
	}
	if optimized, ok := dst.(io.ReaderFrom); ok {
		n, err := optimized.ReadFrom(reader)
		return NewSendResult(n, StrategyReaderFrom), err
	}
	n, err := copyBuffered(dst, reader, s.bufferSize)
	return NewSendResult(n, StrategyBufferedCopy), err
}

func copyBuffered(dst io.Writer, src io.Reader, size int) (int64, error) {
	if size == defaultBufferSize {
		value := defaultBufferPool.Get()
		buffer := *(value.(*[]byte))
		defer defaultBufferPool.Put(&buffer)
		return io.CopyBuffer(dst, src, buffer)
	}
	return io.CopyBuffer(dst, src, make([]byte, size))
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(buffer)
}
