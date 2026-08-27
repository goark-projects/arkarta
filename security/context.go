package security

import "context"

type contextKey struct{}

// SecurityContext 表示当前执行链的安全上下文。
type SecurityContext struct {
	authentication Authentication
}

// NewContext 创建安全上下文。
func NewContext(authentication Authentication) SecurityContext {
	return SecurityContext{authentication: authentication}
}

// Authentication 返回当前认证结果。
func (c SecurityContext) Authentication() Authentication {
	return c.authentication
}

// Authenticated 表示当前上下文是否已认证。
func (c SecurityContext) Authenticated() bool {
	return c.authentication.Authenticated()
}

// ContextWithSecurity 将安全上下文写入标准库 context。
func ContextWithSecurity(parent context.Context, securityContext SecurityContext) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, contextKey{}, securityContext)
}

// FromContext 返回当前安全上下文。
func FromContext(ctx context.Context) (SecurityContext, bool) {
	if ctx == nil {
		return SecurityContext{}, false
	}
	value, ok := ctx.Value(contextKey{}).(SecurityContext)
	return value, ok
}

// AuthenticationFromContext 返回当前认证结果。
func AuthenticationFromContext(ctx context.Context) (Authentication, bool) {
	securityContext, ok := FromContext(ctx)
	if !ok || !securityContext.Authenticated() {
		return Authentication{}, false
	}
	return securityContext.Authentication(), true
}
