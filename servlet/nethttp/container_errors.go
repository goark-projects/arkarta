package nethttp

import "errors"

// ErrUnsupportedProfile 表示 net/http 参考容器不支持部署声明的 Profile。
var ErrUnsupportedProfile = errors.New("arkarta/servlet/nethttp: unsupported profile")

// ErrNilContainer 表示 Server 缺少参考容器。
var ErrNilContainer = errors.New("arkarta/servlet/nethttp: container is nil")

// ErrNilListener 表示 Server 缺少网络监听器。
var ErrNilListener = errors.New("arkarta/servlet/nethttp: listener is nil")
