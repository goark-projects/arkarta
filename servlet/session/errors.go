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

// ErrCookieConfigLocked 表示应用已经启动，不能再修改 Session Cookie 配置。
var ErrCookieConfigLocked = errors.New("arkarta/servlet/session: cookie config is locked")

// ErrInvalidURLRewriteConfig 表示 URL 重写配置非法。
var ErrInvalidURLRewriteConfig = errors.New("arkarta/servlet/session: invalid url rewrite config")

// ErrNilStore 表示会话持久化存储为空。
var ErrNilStore = errors.New("arkarta/servlet/session: store is nil")

// ErrDuplicateSessionID 表示指定会话 ID 已存在。
var ErrDuplicateSessionID = errors.New("arkarta/servlet/session: duplicate session id")
