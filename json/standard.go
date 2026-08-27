package json

import (
	"bytes"
	stdjson "encoding/json"
	"io"
)

// StandardCodec 是基于标准库 encoding/json 的 JSON 实现。
type StandardCodec struct {
	escapeHTML            bool
	indentPrefix          string
	indent                string
	disallowUnknownFields bool
	useNumber             bool
	maxBytes              int64
}

// NewStandardCodec 创建标准库 JSON 编解码器。
func NewStandardCodec(options ...Option) *StandardCodec {
	codec := &StandardCodec{
		escapeHTML: true,
		maxBytes:   DefaultMaxBytes,
	}
	for _, option := range options {
		if option != nil {
			option(codec)
		}
	}
	return codec
}

// Name 返回实现名称。
func (c *StandardCodec) Name() string {
	return "encoding/json"
}

// Marshal 编码 JSON 字节。
func (c *StandardCodec) Marshal(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder, err := c.NewEncoder(&buffer)
	if err != nil {
		return nil, err
	}
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buffer.Bytes(), "\n"), nil
}

// Unmarshal 解码 JSON 字节。
func (c *StandardCodec) Unmarshal(data []byte, target any) error {
	if target == nil {
		return ErrNilTarget
	}
	if c.maxBytes >= 0 && int64(len(data)) > c.maxBytes {
		return ErrPayloadTooLarge
	}
	decoder, err := c.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return err
	}
	return decoder.Decode(target)
}

// NewEncoder 创建流式编码器。
func (c *StandardCodec) NewEncoder(writer io.Writer) (Encoder, error) {
	if writer == nil {
		return nil, ErrNilWriter
	}
	encoder := stdjson.NewEncoder(writer)
	encoder.SetEscapeHTML(c.escapeHTML)
	if c.indent != "" {
		encoder.SetIndent(c.indentPrefix, c.indent)
	}
	return &standardEncoder{encoder: encoder}, nil
}

// NewDecoder 创建流式解码器。
func (c *StandardCodec) NewDecoder(reader io.Reader) (Decoder, error) {
	if reader == nil {
		return nil, ErrNilReader
	}
	if c.maxBytes >= 0 {
		reader = &limitReader{reader: reader, limit: c.maxBytes}
	}
	decoder := stdjson.NewDecoder(reader)
	if c.disallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if c.useNumber {
		decoder.UseNumber()
	}
	return &standardDecoder{decoder: decoder}, nil
}

type standardEncoder struct {
	encoder *stdjson.Encoder
}

func (e *standardEncoder) Encode(value any) error {
	return e.encoder.Encode(value)
}

type standardDecoder struct {
	decoder *stdjson.Decoder
}

func (d *standardDecoder) Decode(target any) error {
	if target == nil {
		return ErrNilTarget
	}
	return d.decoder.Decode(target)
}

type limitReader struct {
	reader io.Reader
	limit  int64
	read   int64
}

func (r *limitReader) Read(data []byte) (int, error) {
	remaining := r.limit + 1 - r.read
	if remaining <= 0 {
		return 0, ErrPayloadTooLarge
	}
	if int64(len(data)) > remaining {
		data = data[:int(remaining)]
	}
	n, err := r.reader.Read(data)
	r.read += int64(n)
	if r.read > r.limit {
		return n, ErrPayloadTooLarge
	}
	return n, err
}
