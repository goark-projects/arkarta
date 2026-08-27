package json

import "errors"

// ErrNilReader 表示 JSON 解码缺少输入流。
var ErrNilReader = errors.New("arkarta/json: reader is nil")

// ErrNilWriter 表示 JSON 编码缺少输出流。
var ErrNilWriter = errors.New("arkarta/json: writer is nil")

// ErrNilTarget 表示 JSON 解码目标为空。
var ErrNilTarget = errors.New("arkarta/json: target is nil")

// ErrPayloadTooLarge 表示 JSON 输入超过安全限制。
var ErrPayloadTooLarge = errors.New("arkarta/json: payload too large")
