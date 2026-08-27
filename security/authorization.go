package security

import "context"

// AuthorizationRequest 表示一次授权请求。
type AuthorizationRequest struct {
	Authentication Authentication
	Resource       string
	Action         string
	Attributes     map[string]any
}

// Attribute 返回授权请求附加属性。
func (r AuthorizationRequest) Attribute(name string) (any, bool) {
	value, ok := r.Attributes[name]
	return value, ok
}

// AuthorizationDecision 表示授权决策结果。
type AuthorizationDecision struct {
	granted bool
	reason  string
}

// Grant 返回允许访问的决策。
func Grant() AuthorizationDecision {
	return AuthorizationDecision{granted: true}
}

// Deny 返回拒绝访问的决策。
func Deny(reason string) AuthorizationDecision {
	if reason == "" {
		reason = ErrAccessDenied.Error()
	}
	return AuthorizationDecision{reason: reason}
}

// Granted 表示是否允许访问。
func (d AuthorizationDecision) Granted() bool {
	return d.granted
}

// Reason 返回拒绝访问原因。
func (d AuthorizationDecision) Reason() string {
	return d.reason
}

// Authorizer 表示授权决策入口。
type Authorizer interface {
	Authorize(ctx context.Context, request AuthorizationRequest) (AuthorizationDecision, error)
}

// AuthorizerFunc 将普通函数适配为 Authorizer。
type AuthorizerFunc func(ctx context.Context, request AuthorizationRequest) (AuthorizationDecision, error)

// Authorize 执行底层授权函数。
func (f AuthorizerFunc) Authorize(ctx context.Context, request AuthorizationRequest) (AuthorizationDecision, error) {
	if f == nil {
		return Deny(ErrAccessDenied.Error()), nil
	}
	return f(ctx, request)
}

// RequireAuthority 创建要求指定权限的授权器。
func RequireAuthority(authority Authority) Authorizer {
	return AuthorizerFunc(func(ctx context.Context, request AuthorizationRequest) (AuthorizationDecision, error) {
		if err := ctx.Err(); err != nil {
			return AuthorizationDecision{}, err
		}
		if !request.Authentication.Authenticated() {
			return Deny(ErrUnauthenticated.Error()), nil
		}
		if request.Authentication.HasAuthority(authority) {
			return Grant(), nil
		}
		return Deny(ErrAccessDenied.Error()), nil
	})
}
