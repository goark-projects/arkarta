package json

import "io"

const (
	// ContentType 是 Arkarta JSON 默认响应媒体类型。
	ContentType = "application/json"

	// DefaultMaxBytes 是 JSON 请求体默认安全读取上限。
	DefaultMaxBytes int64 = 1 << 20
)

// Codec 定义 JSON 编解码标准端口。
type Codec interface {
	Name() string
	Marshal(value any) ([]byte, error)
	Unmarshal(data []byte, target any) error
	NewEncoder(writer io.Writer) (Encoder, error)
	NewDecoder(reader io.Reader) (Decoder, error)
}

// Encoder 表示面向输出流的 JSON 编码器。
type Encoder interface {
	Encode(value any) error
}

// Decoder 表示面向输入流的 JSON 解码器。
type Decoder interface {
	Decode(target any) error
}

// DefaultCodec 返回基于标准库 encoding/json 的默认实现。
func DefaultCodec() Codec {
	return NewStandardCodec()
}

// Marshal 使用指定 Codec 编码 JSON。
func Marshal(codec Codec, value any) ([]byte, error) {
	if codec == nil {
		codec = DefaultCodec()
	}
	return codec.Marshal(value)
}

// Unmarshal 使用指定 Codec 解码 JSON。
func Unmarshal(codec Codec, data []byte, target any) error {
	if codec == nil {
		codec = DefaultCodec()
	}
	return codec.Unmarshal(data, target)
}

// Encode 使用指定 Codec 向输出流编码 JSON。
func Encode(codec Codec, writer io.Writer, value any) error {
	if codec == nil {
		codec = DefaultCodec()
	}
	encoder, err := codec.NewEncoder(writer)
	if err != nil {
		return err
	}
	return encoder.Encode(value)
}

// Decode 使用指定 Codec 从输入流解码 JSON。
func Decode(codec Codec, reader io.Reader, target any) error {
	if codec == nil {
		codec = DefaultCodec()
	}
	decoder, err := codec.NewDecoder(reader)
	if err != nil {
		return err
	}
	return decoder.Decode(target)
}
