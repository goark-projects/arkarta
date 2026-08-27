package security

import "sort"

// Authentication 表示一次认证结果快照。
type Authentication struct {
	principal     Principal
	credentials   Credential
	authorities   []Authority
	authenticated bool
	details       map[string]any
}

// AuthenticationOption 定制认证结果。
type AuthenticationOption func(*Authentication)

// NewAuthentication 创建认证结果。
func NewAuthentication(principal Principal, options ...AuthenticationOption) Authentication {
	auth := Authentication{
		principal: principal,
		details:   make(map[string]any),
	}
	for _, option := range options {
		if option != nil {
			option(&auth)
		}
	}
	auth.authorities = normalizeAuthorities(auth.authorities)
	return auth
}

// WithCredentials 设置认证使用的凭证快照。
func WithCredentials(credentials Credential) AuthenticationOption {
	return func(auth *Authentication) {
		auth.credentials = credentials
	}
}

// WithAuthorities 设置主体权限集合。
func WithAuthorities(authorities ...Authority) AuthenticationOption {
	return func(auth *Authentication) {
		auth.authorities = append(auth.authorities, authorities...)
	}
}

// WithAuthenticated 设置认证是否已通过。
func WithAuthenticated(authenticated bool) AuthenticationOption {
	return func(auth *Authentication) {
		auth.authenticated = authenticated
	}
}

// WithDetail 设置认证附加信息。
func WithDetail(key string, value any) AuthenticationOption {
	return func(auth *Authentication) {
		if key == "" {
			return
		}
		if value == nil {
			delete(auth.details, key)
			return
		}
		auth.details[key] = value
	}
}

// Principal 返回认证主体。
func (a Authentication) Principal() Principal {
	return a.principal
}

// Credentials 返回认证凭证。
func (a Authentication) Credentials() Credential {
	return a.credentials
}

// Authorities 返回权限集合副本。
func (a Authentication) Authorities() []Authority {
	return append([]Authority(nil), a.authorities...)
}

// Authenticated 表示认证是否已通过。
func (a Authentication) Authenticated() bool {
	return a.authenticated && a.principal != nil
}

// HasAuthority 判断主体是否拥有指定权限。
func (a Authentication) HasAuthority(authority Authority) bool {
	for _, candidate := range a.authorities {
		if candidate == authority {
			return true
		}
	}
	return false
}

// Detail 返回指定认证附加信息。
func (a Authentication) Detail(key string) (any, bool) {
	value, ok := a.details[key]
	return value, ok
}

// Details 返回认证附加信息副本。
func (a Authentication) Details() map[string]any {
	result := make(map[string]any, len(a.details))
	for key, value := range a.details {
		result[key] = value
	}
	return result
}

func normalizeAuthorities(authorities []Authority) []Authority {
	seen := make(map[Authority]struct{}, len(authorities))
	result := make([]Authority, 0, len(authorities))
	for _, authority := range authorities {
		if authority == "" {
			continue
		}
		if _, exists := seen[authority]; exists {
			continue
		}
		seen[authority] = struct{}{}
		result = append(result, authority)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}
