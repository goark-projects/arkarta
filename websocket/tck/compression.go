package tck

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"goark.dev/arkarta/websocket"
)

// CompressionFactory 创建 permessage-deflate 压缩协商器。
type CompressionFactory func(options ...websocket.PerMessageDeflateOption) *websocket.PerMessageDeflate

// RunCompression 执行 WebSocket 压缩扩展兼容性测试。
func RunCompression(t *testing.T, factory CompressionFactory) {
	t.Helper()
	t.Run("round_trip_message", func(t *testing.T) {
		codec := factory()
		payload := []byte(strings.Repeat("arkarta-websocket-", 8))
		compressed, err := codec.CompressMessage(payload)
		if err != nil {
			t.Fatalf("CompressMessage failed: %v", err)
		}
		if bytes.HasSuffix(compressed, []byte{0x00, 0x00, 0xff, 0xff}) {
			t.Fatalf("compressed payload keeps flush tail: %x", compressed)
		}
		got, err := codec.DecompressMessage(compressed, int64(len(payload)))
		if err != nil {
			t.Fatalf("DecompressMessage failed: %v", err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("payload = %q, want original", string(got))
		}
	})
	t.Run("enforces_message_limit", func(t *testing.T) {
		codec := factory()
		compressed, err := codec.CompressMessage([]byte(strings.Repeat("x", 32)))
		if err != nil {
			t.Fatalf("CompressMessage failed: %v", err)
		}
		if _, err := codec.DecompressMessage(compressed, 8); !errors.Is(err, websocket.ErrMessageTooLarge) {
			t.Fatalf("DecompressMessage err = %v, want ErrMessageTooLarge", err)
		}
	})
}
