package security

import "context"

// AuthenticationManager 表示认证管理入口。
type AuthenticationManager interface {
	Authenticate(ctx context.Context, credential Credential) (Authentication, error)
}

// AuthenticationManagerFunc 将普通函数适配为 AuthenticationManager。
type AuthenticationManagerFunc func(ctx context.Context, credential Credential) (Authentication, error)

// Authenticate 执行底层认证函数。
func (f AuthenticationManagerFunc) Authenticate(ctx context.Context, credential Credential) (Authentication, error) {
	if f == nil {
		return Authentication{}, ErrBadCredentials
	}
	return f(ctx, credential)
}
