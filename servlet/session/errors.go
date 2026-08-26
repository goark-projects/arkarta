package session

import "errors"

// ErrInvalidSession 表示会话已经失效。
var ErrInvalidSession = errors.New("arkarta/servlet/session: session is invalid")

// ErrSessionNotFound 表示指定会话不存在。
var ErrSessionNotFound = errors.New("arkarta/servlet/session: session not found")

// ErrNilManager 表示会话访问器缺少 Manager。
var ErrNilManager = errors.New("arkarta/servlet/session: manager is nil")

// ErrNilRequest 表示会话访问器缺少请求对象。
var ErrNilRequest = errors.New("arkarta/servlet/session: request is nil")

// ErrInvalidCookieConfig 表示会话 Cookie 配置非法。
var ErrInvalidCookieConfig = errors.New("arkarta/servlet/session: invalid cookie config")
