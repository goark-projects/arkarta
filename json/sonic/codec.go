package sonic

import (
	"bytes"
	"io"

	bytedance "github.com/bytedance/sonic"

	arkjson "goark.dev/arkarta/json"
)

// Codec 是基于 bytedance sonic 的 Arkarta JSON 实现。
type Codec struct {
	api                   bytedance.API
	escapeHTML            bool
	indentPrefix          string
	indent                string
	disallowUnknownFields bool
	useNumber             bool
	sortMapKeys           bool
	maxBytes              int64
}

// NewCodec 创建 sonic JSON 编解码器。
func NewCodec(options ...Option) *Codec {
	codec := &Codec{
		escapeHTML: true,
		maxBytes:   arkjson.DefaultMaxBytes,
	}
	for _, option := range options {
		if option != nil {
			option(codec)
		}
	}
	codec.api = bytedance.Config{
		EscapeHTML:            codec.escapeHTML,
		SortMapKeys:           codec.sortMapKeys,
		DisallowUnknownFields: codec.disallowUnknownFields,
		UseNumber:             codec.useNumber,
		NoEncoderNewline:      true,
	}.Froze()
	return codec
}

// Name 返回实现名称。
func (c *Codec) Name() string {
	return "sonic"
}

// Marshal 编码 JSON 字节。
func (c *Codec) Marshal(value any) ([]byte, error) {
	if c.indent != "" {
		return c.api.MarshalIndent(value, c.indentPrefix, c.indent)
	}
	return c.api.Marshal(value)
}

// Unmarshal 解码 JSON 字节。
func (c *Codec) Unmarshal(data []byte, target any) error {
	if target == nil {
		return arkjson.ErrNilTarget
	}
	if c.maxBytes >= 0 && int64(len(data)) > c.maxBytes {
		return arkjson.ErrPayloadTooLarge
	}
	return c.api.Unmarshal(data, target)
}

// NewEncoder 创建流式编码器。
func (c *Codec) NewEncoder(writer io.Writer) (arkjson.Encoder, error) {
	if writer == nil {
		return nil, arkjson.ErrNilWriter
	}
	return &encoder{codec: c, writer: writer}, nil
}

// NewDecoder 创建流式解码器。
func (c *Codec) NewDecoder(reader io.Reader) (arkjson.Decoder, error) {
	if reader == nil {
		return nil, arkjson.ErrNilReader
	}
	if c.maxBytes >= 0 {
		reader = &limitReader{reader: reader, limit: c.maxBytes}
	}
	return &decoder{codec: c, reader: reader}, nil
}

type encoder struct {
	codec  *Codec
	writer io.Writer
}

func (e *encoder) Encode(value any) error {
	data, err := e.codec.Marshal(value)
	if err != nil {
		return err
	}
	_, err = e.writer.Write(data)
	return err
}

type decoder struct {
	codec  *Codec
	reader io.Reader
}

func (d *decoder) Decode(target any) error {
	if target == nil {
		return arkjson.ErrNilTarget
	}
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(d.reader); err != nil {
		return err
	}
	return d.codec.Unmarshal(buffer.Bytes(), target)
}

type limitReader struct {
	reader io.Reader
	limit  int64
	read   int64
}

func (r *limitReader) Read(data []byte) (int, error) {
	remaining := r.limit + 1 - r.read
	if remaining <= 0 {
		return 0, arkjson.ErrPayloadTooLarge
	}
	if int64(len(data)) > remaining {
		data = data[:int(remaining)]
	}
	n, err := r.reader.Read(data)
	r.read += int64(n)
	if r.read > r.limit {
		return n, arkjson.ErrPayloadTooLarge
	}
	return n, err
}
