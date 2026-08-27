package nativeio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestStandardSenderSendsFileRegion(t *testing.T) {
	t.Parallel()

	region, err := NewFileRegion(strings.NewReader("0123456789"), 2, 4)
	if err != nil {
		t.Fatalf("NewFileRegion failed: %v", err)
	}
	var dst bytes.Buffer
	result, err := NewStandardSender().SendFile(context.Background(), &dst, region)
	if err != nil {
		t.Fatalf("SendFile failed: %v", err)
	}
	if dst.String() != "2345" {
		t.Fatalf("body = %q, want 2345", dst.String())
	}
	if result.Bytes() != 4 || result.Strategy() != StrategyReaderFrom {
		t.Fatalf("result = %d/%s, want 4/reader-from", result.Bytes(), result.Strategy())
	}
}

func TestStandardSenderUsesBufferedFallback(t *testing.T) {
	t.Parallel()

	region, err := NewFileRegion(strings.NewReader("abcdef"), 1, 3)
	if err != nil {
		t.Fatalf("NewFileRegion failed: %v", err)
	}
	dst := writerOnly{writer: new(strings.Builder)}
	result, err := NewStandardSender().SendFile(context.Background(), dst, region)
	if err != nil {
		t.Fatalf("SendFile failed: %v", err)
	}
	if dst.writer.String() != "bcd" {
		t.Fatalf("body = %q, want bcd", dst.writer.String())
	}
	if result.Bytes() != 3 || result.Strategy() != StrategyBufferedCopy {
		t.Fatalf("result = %d/%s, want 3/buffered-copy", result.Bytes(), result.Strategy())
	}
}

func TestStandardSenderValidatesInputsAndContext(t *testing.T) {
	t.Parallel()

	sender := NewStandardSender()
	region, err := NewFileRegion(strings.NewReader("abc"), 0, 1)
	if err != nil {
		t.Fatalf("NewFileRegion failed: %v", err)
	}
	if _, err := sender.SendFile(context.Background(), nil, region); !errors.Is(err, ErrNilWriter) {
		t.Fatalf("nil writer err = %v, want ErrNilWriter", err)
	}
	if _, err := sender.SendFile(context.Background(), io.Discard, FileRegion{}); !errors.Is(err, ErrNilSource) {
		t.Fatalf("nil source err = %v, want ErrNilSource", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := sender.SendFile(ctx, io.Discard, region); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled err = %v, want context.Canceled", err)
	}
}

func TestCapabilitiesAreStable(t *testing.T) {
	t.Parallel()

	capabilities := NewCapabilities(CapabilityKqueue, CapabilitySendfile, CapabilitySendfile)
	if !capabilities.Has(CapabilitySendfile) || capabilities.Has(CapabilityIOUring) {
		t.Fatalf("capability lookup mismatch")
	}
	values := capabilities.Values()
	if len(values) != 2 || values[0] != CapabilityKqueue || values[1] != CapabilitySendfile {
		t.Fatalf("values = %#v, want stable sorted list", values)
	}
}

type writerOnly struct {
	writer *strings.Builder
}

func (w writerOnly) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}
