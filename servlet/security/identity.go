package security

import "sort"

const (
	// AuthTypeBasic 表示 HTTP Basic 认证。
	AuthTypeBasic = "BASIC"
)

// Identity 表示一次认证成功后的主体与角色集合。
type Identity struct {
	principal Principal
	authType  string
	roles     []string
}

// NewIdentity 创建认证身份快照。
func NewIdentity(principal Principal, authType string, roles ...string) Identity {
	identity := Identity{
		principal: principal,
		authType:  authType,
		roles:     normalizeRoles(roles),
	}
	return identity
}

// Principal 返回认证主体。
func (i Identity) Principal() Principal {
	return i.principal
}

// AuthType 返回认证方式。
func (i Identity) AuthType() string {
	return i.authType
}

// Roles 返回角色集合副本。
func (i Identity) Roles() []string {
	return append([]string(nil), i.roles...)
}

// Valid 判断身份是否包含有效主体。
func (i Identity) Valid() bool {
	return i.principal != nil
}

func normalizeRoles(roles []string) []string {
	seen := make(map[string]struct{}, len(roles))
	result := make([]string, 0, len(roles))
	for _, role := range roles {
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		result = append(result, role)
	}
	sort.Strings(result)
	return result
}
