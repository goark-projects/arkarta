package security

import (
	"context"
	"errors"
	"net/http"

	"goark.dev/arkarta/servlet"
)

// ErrNilAuthenticator 表示认证入口为空。
var ErrNilAuthenticator = errors.New("arkarta/servlet/security: authenticator is nil")

// ErrAuthenticationFailed 表示凭证认证失败。
var ErrAuthenticationFailed = errors.New("arkarta/servlet/security: authentication failed")

// Authenticator 定义容器认证入口。
type Authenticator interface {
	Authenticate(ctx context.Context, req *servlet.Request, res servlet.Response) (Identity, bool, error)
	Login(ctx context.Context, req *servlet.Request, username, password string) (Identity, error)
	Logout(ctx context.Context, req *servlet.Request, res servlet.Response) error
}

// Realm 定义用户名密码到主体身份的校验边界。
type Realm interface {
	Verify(ctx context.Context, username, password string) (Identity, bool, error)
}

// RealmFunc 将函数适配为 Realm。
type RealmFunc func(ctx context.Context, username, password string) (Identity, bool, error)

// Verify 执行用户名密码校验。
func (f RealmFunc) Verify(ctx context.Context, username, password string) (Identity, bool, error) {
	return f(ctx, username, password)
}

// Authenticate 执行容器认证；认证成功时写入当前请求安全上下文。
func Authenticate(ctx context.Context, req *servlet.Request, res servlet.Response, authenticator Authenticator) (bool, error) {
	if authenticator == nil {
		return false, ErrNilAuthenticator
	}
	identity, ok, err := authenticator.Authenticate(ctx, req, res)
	if err != nil || !ok {
		return false, err
	}
	if !identity.Valid() {
		return false, ErrAuthenticationFailed
	}
	BindIdentity(req, identity)
	return true, nil
}

// Login 使用显式用户名密码登录；认证成功时写入当前请求安全上下文。
func Login(ctx context.Context, req *servlet.Request, username, password string, authenticator Authenticator) error {
	if authenticator == nil {
		return ErrNilAuthenticator
	}
	identity, err := authenticator.Login(ctx, req, username, password)
	if err != nil {
		return err
	}
	if !identity.Valid() {
		return ErrAuthenticationFailed
	}
	BindIdentity(req, identity)
	return nil
}

// Logout 清理当前请求安全上下文并调用认证器退出逻辑。
func Logout(ctx context.Context, req *servlet.Request, res servlet.Response, authenticator Authenticator) error {
	if authenticator == nil {
		return ErrNilAuthenticator
	}
	if err := authenticator.Logout(ctx, req, res); err != nil {
		return err
	}
	ClearPrincipal(req)
	return nil
}

func setUnauthorized(res servlet.Response, challenge string) {
	if res == nil {
		return
	}
	if challenge != "" {
		res.Header().Set("WWW-Authenticate", challenge)
	}
	res.SetStatus(http.StatusUnauthorized)
}
