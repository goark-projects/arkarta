package tck

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	arkjson "goark.dev/arkarta/json"
)

// CodecFactory 创建不同配置下的 JSON 编解码器。
type CodecFactory struct {
	New                  func() arkjson.Codec
	WithEscapeHTML       func(enabled bool) arkjson.Codec
	WithMaxBytes         func(maxBytes int64) arkjson.Codec
	WithUnknownFieldGate func(enabled bool) arkjson.Codec
	WithUseNumber        func(enabled bool) arkjson.Codec
}

// RunCodec 执行 Arkarta JSON Codec 兼容性测试。
func RunCodec(t *testing.T, factory CodecFactory) {
	t.Helper()
	t.Run("marshals_unmarshals_payload", func(t *testing.T) {
		runMarshalUnmarshal(t, factory)
	})
	t.Run("streams_payload", func(t *testing.T) {
		runStreaming(t, factory)
	})
	t.Run("honors_nil_guards_and_size_limit", func(t *testing.T) {
		runNilAndLimit(t, factory)
	})
	t.Run("honors_unknown_fields_and_use_number", func(t *testing.T) {
		runDecodeOptions(t, factory)
	})
}

func runMarshalUnmarshal(t *testing.T, factory CodecFactory) {
	t.Helper()
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	codec := requireCodec(t, "WithEscapeHTML", factory.WithEscapeHTML(false))
	data, err := codec.Marshal(payload{Name: "<arkarta>", Age: 1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var got payload
	if err := codec.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Name != "<arkarta>" || got.Age != 1 {
		t.Fatalf("payload = %#v, want decoded payload", got)
	}
}

func runStreaming(t *testing.T, factory CodecFactory) {
	t.Helper()
	codec := requireCodec(t, "New", factory.New())
	var buffer bytes.Buffer
	encoder, err := codec.NewEncoder(&buffer)
	if err != nil {
		t.Fatalf("NewEncoder failed: %v", err)
	}
	if err := encoder.Encode(map[string]string{"name": "arkarta"}); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	decoder, err := codec.NewDecoder(&buffer)
	if err != nil {
		t.Fatalf("NewDecoder failed: %v", err)
	}
	var got map[string]string
	if err := decoder.Decode(&got); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if got["name"] != "arkarta" {
		t.Fatalf("payload = %#v, want name", got)
	}
}

func runNilAndLimit(t *testing.T, factory CodecFactory) {
	t.Helper()
	codec := requireCodec(t, "WithMaxBytes", factory.WithMaxBytes(4))
	if _, err := codec.NewEncoder(nil); !errors.Is(err, arkjson.ErrNilWriter) {
		t.Fatalf("nil writer err = %v, want ErrNilWriter", err)
	}
	if _, err := codec.NewDecoder(nil); !errors.Is(err, arkjson.ErrNilReader) {
		t.Fatalf("nil reader err = %v, want ErrNilReader", err)
	}
	var target map[string]any
	if err := codec.Unmarshal([]byte(`{"name":"arkarta"}`), &target); !errors.Is(err, arkjson.ErrPayloadTooLarge) {
		t.Fatalf("large payload err = %v, want ErrPayloadTooLarge", err)
	}
	if err := codec.Unmarshal([]byte(`{}`), nil); !errors.Is(err, arkjson.ErrNilTarget) {
		t.Fatalf("nil target err = %v, want ErrNilTarget", err)
	}
}

func runDecodeOptions(t *testing.T, factory CodecFactory) {
	t.Helper()
	type payload struct {
		Name string `json:"name"`
	}
	codec := requireCodec(t, "WithUnknownFieldGate", factory.WithUnknownFieldGate(true))
	var got payload
	err := codec.Unmarshal([]byte(`{"name":"arkarta","extra":true}`), &got)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unknown") {
		t.Fatalf("unknown field err = %v, want unknown field failure", err)
	}

	var numberPayload map[string]any
	codec = requireCodec(t, "WithUseNumber", factory.WithUseNumber(true))
	if err := codec.Unmarshal([]byte(`{"id":9223372036854775807}`), &numberPayload); err != nil {
		t.Fatalf("UseNumber Unmarshal failed: %v", err)
	}
	if _, ok := numberPayload["id"].(stdjson.Number); !ok {
		t.Fatalf("id type = %T, want json.Number", numberPayload["id"])
	}
}

func requireCodec(t *testing.T, name string, codec arkjson.Codec) arkjson.Codec {
	t.Helper()
	if codec == nil {
		t.Fatal(fmt.Sprintf("%s returned nil codec", name))
	}
	return codec
}
