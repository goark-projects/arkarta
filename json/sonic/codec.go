package sonic

import arkjson "goark.dev/arkarta/json"

// Codec 是根包 SonicCodec 的兼容别名。
type Codec = arkjson.SonicCodec

// NewCodec 创建 sonic JSON 编解码器。
func NewCodec(options ...Option) *Codec {
	return arkjson.NewSonicCodec(options...)
}
