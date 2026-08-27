package tck

import (
	"bytes"
	"errors"
	"testing"

	"goark.dev/arkarta/websocket/frame"
)

// RunFrameCodec 执行 WebSocket 帧层兼容性测试。
func RunFrameCodec(t *testing.T) {
	t.Helper()
	t.Run("reads_and_writes_masked_text_frame", func(t *testing.T) {
		var buffer bytes.Buffer
		key := frame.MaskKey{1, 2, 3, 4}
		if err := frame.Write(&buffer, frame.New(frame.OpText, []byte("hello"), frame.WithMask(key))); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		got, err := frame.Read(&buffer, frame.WithMaskPolicy(frame.MaskRequired))
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if got.OpCode() != frame.OpText || string(got.Payload()) != "hello" || !got.Masked() {
			t.Fatalf("frame = %v/%q/%v", got.OpCode(), string(got.Payload()), got.Masked())
		}
	})
	t.Run("rejects_unmasked_client_frame", func(t *testing.T) {
		var buffer bytes.Buffer
		if err := frame.Write(&buffer, frame.New(frame.OpText, []byte("hello"))); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if _, err := frame.Read(&buffer, frame.WithMaskPolicy(frame.MaskRequired)); !errors.Is(err, frame.ErrMaskRequired) {
			t.Fatalf("Read err = %v, want ErrMaskRequired", err)
		}
	})
	t.Run("aggregates_fragments", func(t *testing.T) {
		assembler := frame.NewAssembler()
		if _, complete, err := assembler.Add(frame.New(frame.OpText, []byte("ar"), frame.WithFin(false))); err != nil || complete {
			t.Fatalf("first fragment = complete %v err %v, want pending", complete, err)
		}
		message, complete, err := assembler.Add(frame.New(frame.OpContinuation, []byte("karta")))
		if err != nil {
			t.Fatalf("continuation failed: %v", err)
		}
		if !complete || message.OpCode() != frame.OpText || string(message.Payload()) != "arkarta" {
			t.Fatalf("message = %v/%q complete=%v", message.OpCode(), string(message.Payload()), complete)
		}
	})
	t.Run("round_trips_close_payload", func(t *testing.T) {
		payload, err := frame.ClosePayload(1000, "normal")
		if err != nil {
			t.Fatalf("ClosePayload failed: %v", err)
		}
		code, reason, err := frame.ParseClosePayload(payload)
		if err != nil {
			t.Fatalf("ParseClosePayload failed: %v", err)
		}
		if code != 1000 || reason != "normal" {
			t.Fatalf("close = %d/%q, want 1000/normal", code, reason)
		}
	})
}
