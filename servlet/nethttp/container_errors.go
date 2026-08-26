package nethttp

import "errors"

// ErrUnsupportedProfile 表示 net/http 参考容器不支持部署声明的 Profile。
var ErrUnsupportedProfile = errors.New("arkarta/servlet/nethttp: unsupported profile")
