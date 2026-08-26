package upgrade

import "errors"

// ErrUnsupported 表示容器不支持协议升级。
var ErrUnsupported = errors.New("arkarta/servlet/upgrade: unsupported")

// ErrNilHandler 表示协议升级处理器为空。
var ErrNilHandler = errors.New("arkarta/servlet/upgrade: handler is nil")

// ErrAlreadyCommitted 表示响应已经提交，不能升级。
var ErrAlreadyCommitted = errors.New("arkarta/servlet/upgrade: response already committed")
