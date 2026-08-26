package session

import "errors"

// ErrInvalidSession 表示会话已经失效。
var ErrInvalidSession = errors.New("arkarta/servlet/session: session is invalid")

// ErrSessionNotFound 表示指定会话不存在。
var ErrSessionNotFound = errors.New("arkarta/servlet/session: session not found")
