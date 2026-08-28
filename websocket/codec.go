package websocket

import arkjson "goark.dev/arkarta/json"

// TextEncoder 将 Go 值编码为 WebSocket 文本消息。
type TextEncoder[T any] interface {
	EncodeText(value T) (string, error)
}

// TextDecoder 将 WebSocket 文本消息解码为 Go 值。
type TextDecoder[T any] interface {
	DecodeText(text string) (T, error)
}

// BinaryEncoder 将 Go 值编码为 WebSocket 二进制消息。
type BinaryEncoder[T any] interface {
	EncodeBinary(value T) ([]byte, error)
}

// BinaryDecoder 将 WebSocket 二进制消息解码为 Go 值。
type BinaryDecoder[T any] interface {
	DecodeBinary(data []byte) (T, error)
}

// JSONTextCodec 使用 Arkarta 默认 JSON 编解码器处理文本消息。
type JSONTextCodec[T any] struct{}

// EncodeText 编码 JSON 文本。
func (JSONTextCodec[T]) EncodeText(value T) (string, error) {
	data, err := arkjson.Marshal(nil, value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DecodeText 解码 JSON 文本。
func (JSONTextCodec[T]) DecodeText(text string) (T, error) {
	var value T
	err := arkjson.Unmarshal(nil, []byte(text), &value)
	return value, err
}
