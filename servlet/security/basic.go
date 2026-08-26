package security

import (
	"context"
	"crypto/subtle"
	"fmt"
	"sort"
	"sync"

	"goark.dev/arkarta/servlet"
)

// BasicAuthenticator 实现 HTTP Basic 认证流程。
type BasicAuthenticator struct {
	realm     Realm
	realmName string
}

// BasicOption 定制 Basic 认证器。
type BasicOption func(*BasicAuthenticator)

// NewBasicAuthenticator 创建 Basic 认证器。
func NewBasicAuthenticator(realm Realm, options ...BasicOption) *BasicAuthenticator {
	authenticator := &BasicAuthenticator{
		realm:     realm,
		realmName: "arkarta",
	}
	for _, option := range options {
		if option != nil {
			option(authenticator)
		}
	}
	return authenticator
}

// WithBasicRealmName 设置 Basic challenge 的 realm 名称。
func WithBasicRealmName(name string) BasicOption {
	return func(authenticator *BasicAuthenticator) {
		if name != "" {
			authenticator.realmName = name
		}
	}
}

// Authenticate 从请求 Authorization 头执行 Basic 认证。
func (a *BasicAuthenticator) Authenticate(ctx context.Context, req *servlet.Request, res servlet.Response) (Identity, bool, error) {
	if a == nil || a.realm == nil {
		return Identity{}, false, ErrNilAuthenticator
	}
	if req == nil || req.HTTPRequest() == nil {
		return Identity{}, false, servlet.ErrNilHTTPRequest
	}
	username, password, ok := req.HTTPRequest().BasicAuth()
	if !ok {
		setUnauthorized(res, a.challenge())
		return Identity{}, false, nil
	}
	identity, verified, err := a.realm.Verify(ctx, username, password)
	if err != nil {
		return Identity{}, false, err
	}
	if !verified {
		setUnauthorized(res, a.challenge())
		return Identity{}, false, ErrAuthenticationFailed
	}
	return identity, true, nil
}

// Login 使用显式凭证执行 Basic 认证。
func (a *BasicAuthenticator) Login(ctx context.Context, _ *servlet.Request, username, password string) (Identity, error) {
	if a == nil || a.realm == nil {
		return Identity{}, ErrNilAuthenticator
	}
	identity, ok, err := a.realm.Verify(ctx, username, password)
	if err != nil {
		return Identity{}, err
	}
	if !ok {
		return Identity{}, ErrAuthenticationFailed
	}
	return identity, nil
}

// Logout 退出 Basic 认证；Basic 本身无服务端状态。
func (a *BasicAuthenticator) Logout(ctx context.Context, _ *servlet.Request, _ servlet.Response) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (a *BasicAuthenticator) challenge() string {
	return fmt.Sprintf(`Basic realm="%s"`, a.realmName)
}

// StaticRealm 是用于测试和最小容器的内存 Realm。
type StaticRealm struct {
	mu    sync.RWMutex
	users map[string]staticUser
}

type staticUser struct {
	password string
	roles    []string
}

// StaticRealmOption 定制 StaticRealm。
type StaticRealmOption func(*StaticRealm)

// NewStaticRealm 创建内存 Realm。
func NewStaticRealm(options ...StaticRealmOption) *StaticRealm {
	realm := &StaticRealm{
		users: make(map[string]staticUser),
	}
	for _, option := range options {
		if option != nil {
			option(realm)
		}
	}
	return realm
}

// WithStaticUser 添加一个测试用户。
func WithStaticUser(username, password string, roles ...string) StaticRealmOption {
	return func(realm *StaticRealm) {
		if realm == nil || username == "" {
			return
		}
		realm.users[username] = staticUser{
			password: password,
			roles:    normalizeRoles(roles),
		}
	}
}

// Verify 校验用户名密码并返回身份。
func (r *StaticRealm) Verify(ctx context.Context, username, password string) (Identity, bool, error) {
	if err := ctx.Err(); err != nil {
		return Identity{}, false, err
	}
	if r == nil {
		return Identity{}, false, ErrAuthenticationFailed
	}
	r.mu.RLock()
	user, ok := r.users[username]
	r.mu.RUnlock()
	if !ok || subtle.ConstantTimeCompare([]byte(user.password), []byte(password)) != 1 {
		return Identity{}, false, nil
	}
	roles := append([]string(nil), user.roles...)
	sort.Strings(roles)
	return NewIdentity(PrincipalFunc(func() string { return username }), AuthTypeBasic, roles...), true, nil
}
