package json

import (
	"bytes"
	"io"

	bytedance "github.com/bytedance/sonic"
)

// SonicCodec 是基于 bytedance sonic 的高性能 JSON 实现。
type SonicCodec struct {
	api                   bytedance.API
	escapeHTML            bool
	indentPrefix          string
	indent                string
	disallowUnknownFields bool
	useNumber             bool
	sortMapKeys           bool
	maxBytes              int64
}

// SonicOption 定制 sonic JSON 编解码行为。
type SonicOption func(*SonicCodec)

// NewSonicCodec 创建 sonic JSON 编解码器。
func NewSonicCodec(options ...SonicOption) *SonicCodec {
	codec := &SonicCodec{
		escapeHTML: true,
		maxBytes:   DefaultMaxBytes,
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

// WithSonicEscapeHTML 设置字符串编码时是否转义 HTML 字符。
func WithSonicEscapeHTML(enabled bool) SonicOption {
	return func(codec *SonicCodec) {
		codec.escapeHTML = enabled
	}
}

// WithSonicIndent 设置 JSON 输出缩进；空 indent 表示紧凑输出。
func WithSonicIndent(prefix, indent string) SonicOption {
	return func(codec *SonicCodec) {
		codec.indentPrefix = prefix
		codec.indent = indent
	}
}

// WithSonicDisallowUnknownFields 设置结构体解码时是否拒绝未知字段。
func WithSonicDisallowUnknownFields(enabled bool) SonicOption {
	return func(codec *SonicCodec) {
		codec.disallowUnknownFields = enabled
	}
}

// WithSonicUseNumber 设置数字解码时是否保留 json.Number 精度。
func WithSonicUseNumber(enabled bool) SonicOption {
	return func(codec *SonicCodec) {
		codec.useNumber = enabled
	}
}

// WithSonicMaxBytes 设置输入流最大读取字节数；负数表示不限制。
func WithSonicMaxBytes(maxBytes int64) SonicOption {
	return func(codec *SonicCodec) {
		codec.maxBytes = maxBytes
	}
}

// WithSonicSortMapKeys 设置编码 map 时是否稳定排序键。
func WithSonicSortMapKeys(enabled bool) SonicOption {
	return func(codec *SonicCodec) {
		codec.sortMapKeys = enabled
	}
}

// Name 返回实现名称。
func (c *SonicCodec) Name() string {
	return "sonic"
}

// Marshal 编码 JSON 字节。
func (c *SonicCodec) Marshal(value any) ([]byte, error) {
	if c.indent != "" {
		return c.api.MarshalIndent(value, c.indentPrefix, c.indent)
	}
	return c.api.Marshal(value)
}

// Unmarshal 解码 JSON 字节。
func (c *SonicCodec) Unmarshal(data []byte, target any) error {
	if target == nil {
		return ErrNilTarget
	}
	if c.maxBytes >= 0 && int64(len(data)) > c.maxBytes {
		return ErrPayloadTooLarge
	}
	return c.api.Unmarshal(data, target)
}

// NewEncoder 创建流式编码器。
func (c *SonicCodec) NewEncoder(writer io.Writer) (Encoder, error) {
	if writer == nil {
		return nil, ErrNilWriter
	}
	return &sonicEncoder{codec: c, writer: writer}, nil
}

// NewDecoder 创建流式解码器。
func (c *SonicCodec) NewDecoder(reader io.Reader) (Decoder, error) {
	if reader == nil {
		return nil, ErrNilReader
	}
	if c.maxBytes >= 0 {
		reader = &sonicLimitReader{reader: reader, limit: c.maxBytes}
	}
	return &sonicDecoder{codec: c, reader: reader}, nil
}

type sonicEncoder struct {
	codec  *SonicCodec
	writer io.Writer
}

func (e *sonicEncoder) Encode(value any) error {
	data, err := e.codec.Marshal(value)
	if err != nil {
		return err
	}
	_, err = e.writer.Write(data)
	return err
}

type sonicDecoder struct {
	codec  *SonicCodec
	reader io.Reader
}

func (d *sonicDecoder) Decode(target any) error {
	if target == nil {
		return ErrNilTarget
	}
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(d.reader); err != nil {
		return err
	}
	return d.codec.Unmarshal(buffer.Bytes(), target)
}

type sonicLimitReader struct {
	reader io.Reader
	limit  int64
	read   int64
}

func (r *sonicLimitReader) Read(data []byte) (int, error) {
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
