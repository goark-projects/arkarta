package tck

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"goark.dev/arkarta/servlet/nativeio"
)

// NativeIOSenderFactory 创建 Native I/O 文件发送器。
type NativeIOSenderFactory func() nativeio.Sender

// RunNativeIO 执行 Native I/O Profile 兼容性测试。
func RunNativeIO(t *testing.T, factory NativeIOSenderFactory) {
	t.Helper()
	t.Run("sends_file_region", func(t *testing.T) {
		sender := factory()
		region, err := nativeio.NewFileRegion(strings.NewReader("0123456789"), 3, 4)
		if err != nil {
			t.Fatalf("NewFileRegion failed: %v", err)
		}
		var dst bytes.Buffer
		result, err := sender.SendFile(context.Background(), &dst, region)
		if err != nil {
			t.Fatalf("SendFile failed: %v", err)
		}
		if dst.String() != "3456" || result.Bytes() != 4 {
			t.Fatalf("body/result = %q/%d, want 3456/4", dst.String(), result.Bytes())
		}
	})
	t.Run("rejects_invalid_region", func(t *testing.T) {
		sender := factory()
		_, err := sender.SendFile(context.Background(), io.Discard, nativeio.FileRegion{})
		if !errors.Is(err, nativeio.ErrNilSource) {
			t.Fatalf("SendFile err = %v, want ErrNilSource", err)
		}
	})
	t.Run("observes_context_cancellation", func(t *testing.T) {
		sender := factory()
		region, err := nativeio.NewFileRegion(strings.NewReader("abc"), 0, 3)
		if err != nil {
			t.Fatalf("NewFileRegion failed: %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = sender.SendFile(ctx, io.Discard, region)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("SendFile err = %v, want context.Canceled", err)
		}
	})
}
