package json

import "testing"

func TestDefaultCodecUsesSonic(t *testing.T) {
	t.Parallel()

	codec := DefaultCodec()
	if codec.Name() != "sonic" {
		t.Fatalf("default codec = %q, want sonic", codec.Name())
	}
}
