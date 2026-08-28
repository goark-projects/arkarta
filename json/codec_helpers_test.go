package json

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNewCodecUsesSonicOptions(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}
	codec := NewCodec(WithEscapeHTML(false), WithSortMapKeys(true))
	data, err := codec.Marshal(payload{Name: "<arkarta>"})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(data) != `{"name":"<arkarta>"}` {
		t.Fatalf("json = %s", data)
	}
	var got payload
	if err := codec.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Name != "<arkarta>" {
		t.Fatalf("payload = %#v", got)
	}
}

func TestNewCodecDecodeOptions(t *testing.T) {
	t.Parallel()

	var value map[string]any
	codec := NewCodec(WithUseNumber(true), WithMaxBytes(128))
	if err := codec.Unmarshal([]byte(`{"id":9223372036854775807}`), &value); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if _, ok := value["id"].(float64); ok {
		t.Fatalf("id type = %T, want precision-preserving number", value["id"])
	}
	if got := fmt.Sprint(value["id"]); got != "9223372036854775807" {
		t.Fatalf("id = %s, want full precision", got)
	}

	var target struct {
		Name string `json:"name"`
	}
	err := NewCodec(WithDisallowUnknownFields(true)).Unmarshal([]byte(`{"name":"a","extra":1}`), &target)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unknown") {
		t.Fatalf("unknown field err = %v", err)
	}
}

func TestNewCodecLimitAndNilGuards(t *testing.T) {
	t.Parallel()

	codec := NewCodec(WithMaxBytes(4))
	var target map[string]any
	if err := codec.Unmarshal([]byte(`{"name":"arkarta"}`), &target); !errors.Is(err, ErrPayloadTooLarge) {
		t.Fatalf("large payload err = %v, want ErrPayloadTooLarge", err)
	}
	if err := codec.Unmarshal([]byte(`{}`), nil); !errors.Is(err, ErrNilTarget) {
		t.Fatalf("nil target err = %v, want ErrNilTarget", err)
	}
	if _, err := codec.NewEncoder(nil); !errors.Is(err, ErrNilWriter) {
		t.Fatalf("nil writer err = %v, want ErrNilWriter", err)
	}
	if _, err := codec.NewDecoder(nil); !errors.Is(err, ErrNilReader) {
		t.Fatalf("nil reader err = %v, want ErrNilReader", err)
	}
}

func TestPackageHelpersUseDefaultCodec(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	if err := Encode(nil, &buffer, map[string]string{"name": "arkarta"}); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	var target map[string]string
	if err := Decode(nil, &buffer, &target); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if target["name"] != "arkarta" {
		t.Fatalf("target = %#v", target)
	}
}
