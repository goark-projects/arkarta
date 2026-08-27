package sonic

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"strings"
	"testing"

	arkjson "goark.dev/arkarta/json"
)

func TestCodecMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}
	codec := NewCodec(WithEscapeHTML(false))
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

func TestCodecOptions(t *testing.T) {
	t.Parallel()

	var value map[string]any
	codec := NewCodec(WithUseNumber(true), WithMaxBytes(128))
	if err := codec.Unmarshal([]byte(`{"id":9223372036854775807}`), &value); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if _, ok := value["id"].(stdjson.Number); !ok {
		t.Fatalf("id type = %T, want json.Number", value["id"])
	}

	var target struct {
		Name string `json:"name"`
	}
	err := NewCodec(WithDisallowUnknownFields(true)).Unmarshal([]byte(`{"name":"a","extra":1}`), &target)
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("unknown field err = %v", err)
	}
}

func TestCodecStreamAndLimit(t *testing.T) {
	t.Parallel()

	codec := NewCodec(WithMaxBytes(64))
	var buffer bytes.Buffer
	encoder, err := codec.NewEncoder(&buffer)
	if err != nil {
		t.Fatalf("NewEncoder failed: %v", err)
	}
	if err := encoder.Encode(map[string]string{"name": "arkarta"}); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	var target map[string]string
	decoder, err := codec.NewDecoder(&buffer)
	if err != nil {
		t.Fatalf("NewDecoder failed: %v", err)
	}
	if err := decoder.Decode(&target); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if target["name"] != "arkarta" {
		t.Fatalf("target = %#v", target)
	}

	if err := NewCodec(WithMaxBytes(4)).Unmarshal([]byte(`{"name":"arkarta"}`), &target); !errors.Is(err, arkjson.ErrPayloadTooLarge) {
		t.Fatalf("large payload err = %v, want ErrPayloadTooLarge", err)
	}
}
