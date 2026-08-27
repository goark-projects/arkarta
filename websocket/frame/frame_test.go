package frame_test

import (
	"bytes"
	"errors"
	"testing"

	"goark.dev/arkarta/websocket/frame"
)

func TestReadWriteMaskedTextFrame(t *testing.T) {
	t.Parallel()

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
		t.Fatalf("frame = opcode %v payload %q masked %v", got.OpCode(), string(got.Payload()), got.Masked())
	}
}

func TestReadRejectsInvalidControlFrame(t *testing.T) {
	t.Parallel()

	data := []byte{byte(frame.OpPing), 0}
	if _, err := frame.Read(bytes.NewReader(data)); !errors.Is(err, frame.ErrProtocol) {
		t.Fatalf("Read err = %v, want ErrProtocol", err)
	}
}

func TestReadWriteExtendedPayload(t *testing.T) {
	t.Parallel()

	payload := bytes.Repeat([]byte("a"), 130)
	var buffer bytes.Buffer
	if err := frame.Write(&buffer, frame.New(frame.OpBinary, payload)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	got, err := frame.Read(&buffer, frame.WithMaskPolicy(frame.MaskForbidden))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if got.OpCode() != frame.OpBinary || !bytes.Equal(got.Payload(), payload) {
		t.Fatalf("payload len = %d, want %d", len(got.Payload()), len(payload))
	}
}

func TestAssemblerHandlesFragmentsAndControlFrames(t *testing.T) {
	t.Parallel()

	assembler := frame.NewAssembler()
	if message, complete, err := assembler.Add(frame.New(frame.OpText, []byte("hel"), frame.WithFin(false))); err != nil || complete || len(message.Payload()) != 0 {
		t.Fatalf("first fragment = %v/%v/%v, want pending", message, complete, err)
	}
	ping, complete, err := assembler.Add(frame.New(frame.OpPing, []byte("p")))
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}
	if !complete || ping.OpCode() != frame.OpPing || string(ping.Payload()) != "p" {
		t.Fatalf("ping message = %#v complete=%v", ping, complete)
	}
	message, complete, err := assembler.Add(frame.New(frame.OpContinuation, []byte("lo")))
	if err != nil {
		t.Fatalf("continuation failed: %v", err)
	}
	if !complete || message.OpCode() != frame.OpText || string(message.Payload()) != "hello" {
		t.Fatalf("message = %#v complete=%v", message, complete)
	}
}

func TestClosePayloadRoundTrip(t *testing.T) {
	t.Parallel()

	payload, err := frame.ClosePayload(1000, "bye")
	if err != nil {
		t.Fatalf("ClosePayload failed: %v", err)
	}
	code, reason, err := frame.ParseClosePayload(payload)
	if err != nil {
		t.Fatalf("ParseClosePayload failed: %v", err)
	}
	if code != 1000 || reason != "bye" {
		t.Fatalf("close = %d/%q, want 1000/bye", code, reason)
	}
	if _, _, err := frame.ParseClosePayload([]byte{1}); !errors.Is(err, frame.ErrInvalidClosePayload) {
		t.Fatalf("short close err = %v, want ErrInvalidClosePayload", err)
	}
}
