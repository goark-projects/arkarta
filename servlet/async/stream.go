package async

import (
	"context"
	"sync"

	"goark.dev/arkarta/servlet"
)

// Stream 提供带上下文检查的流式响应写入。
type Stream struct {
	res servlet.Response

	mu     sync.RWMutex
	closed bool
}

// NewStream 创建流式响应写入器。
func NewStream(res servlet.Response) (*Stream, error) {
	if res == nil {
		return nil, ErrNilResponse
	}
	return &Stream{res: res}, nil
}

// Write 写出响应片段。
func (s *Stream) Write(ctx context.Context, data []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()
	if closed {
		return 0, ErrCompleted
	}
	return s.res.Write(data)
}

// Flush 刷新响应片段。
func (s *Stream) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()
	if closed {
		return ErrCompleted
	}
	return s.res.Flush()
}

// Close 关闭流式写入。
func (s *Stream) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrCompleted
	}
	s.closed = true
	s.mu.Unlock()
	return s.res.Flush()
}
