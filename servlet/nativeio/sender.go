package nativeio

import (
	"context"
	"io"
)

// Strategy 表示一次文件发送实际采用的传输策略。
type Strategy string

const (
	// StrategyBufferedCopy 表示使用用户态缓冲复制。
	StrategyBufferedCopy Strategy = "buffered-copy"
	// StrategyReaderFrom 表示使用目标 Writer 的 ReaderFrom 优化路径。
	StrategyReaderFrom Strategy = "reader-from"
	// StrategyNative 表示容器使用平台原生零拷贝能力。
	StrategyNative Strategy = "native"
)

// SendResult 描述一次文件区段发送结果。
type SendResult struct {
	bytes    int64
	strategy Strategy
}

// NewSendResult 创建发送结果。
func NewSendResult(bytes int64, strategy Strategy) SendResult {
	return SendResult{
		bytes:    bytes,
		strategy: strategy,
	}
}

// Bytes 返回已发送字节数。
func (r SendResult) Bytes() int64 {
	return r.bytes
}

// Strategy 返回传输策略。
func (r SendResult) Strategy() Strategy {
	if r.strategy == "" {
		return StrategyBufferedCopy
	}
	return r.strategy
}

// Native 表示本次发送是否使用容器原生零拷贝能力。
func (r SendResult) Native() bool {
	return r.strategy == StrategyNative
}

// Sender 定义 Native I/O Profile 的文件区段发送契约。
type Sender interface {
	SendFile(ctx context.Context, dst io.Writer, region FileRegion) (SendResult, error)
}

// SenderFunc 将函数适配为 Sender。
type SenderFunc func(ctx context.Context, dst io.Writer, region FileRegion) (SendResult, error)

// SendFile 执行文件区段发送。
func (f SenderFunc) SendFile(ctx context.Context, dst io.Writer, region FileRegion) (SendResult, error) {
	return f(ctx, dst, region)
}
