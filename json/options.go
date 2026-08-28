package json

// Option 定制默认 sonic JSON 编解码行为。
type Option = SonicOption

// NewCodec 创建 Arkarta 默认 JSON 编解码器。
func NewCodec(options ...Option) *SonicCodec {
	return NewSonicCodec(options...)
}

// WithEscapeHTML 设置字符串编码时是否转义 HTML 字符。
func WithEscapeHTML(enabled bool) Option {
	return WithSonicEscapeHTML(enabled)
}

// WithIndent 设置 JSON 输出缩进；空 indent 表示紧凑输出。
func WithIndent(prefix, indent string) Option {
	return WithSonicIndent(prefix, indent)
}

// WithDisallowUnknownFields 设置结构体解码时是否拒绝未知字段。
func WithDisallowUnknownFields(enabled bool) Option {
	return WithSonicDisallowUnknownFields(enabled)
}

// WithUseNumber 设置数字解码时是否保留文本精度。
func WithUseNumber(enabled bool) Option {
	return WithSonicUseNumber(enabled)
}

// WithMaxBytes 设置输入流最大读取字节数；负数表示不限制。
func WithMaxBytes(maxBytes int64) Option {
	return WithSonicMaxBytes(maxBytes)
}

// WithSortMapKeys 设置编码 map 时是否稳定排序键。
func WithSortMapKeys(enabled bool) Option {
	return WithSonicSortMapKeys(enabled)
}
