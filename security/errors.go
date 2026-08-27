package security

import "errors"

// ErrBadCredentials 表示凭证无效或不能被当前认证管理器接受。
var ErrBadCredentials = errors.New("arkarta/security: bad credentials")

// ErrUnauthenticated 表示当前请求没有有效认证信息。
var ErrUnauthenticated = errors.New("arkarta/security: unauthenticated")

// ErrAccessDenied 表示当前主体没有访问权限。
var ErrAccessDenied = errors.New("arkarta/security: access denied")
