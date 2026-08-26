package security

import "goark.dev/arkarta/servlet"

const (
	// AttributePrincipal 保存当前认证主体。
	AttributePrincipal = "arkarta.servlet.security.principal"
	// AttributeAuthType 保存当前认证方式。
	AttributeAuthType = "arkarta.servlet.security.auth_type"
	// AttributeRoles 保存当前主体角色集合。
	AttributeRoles = "arkarta.servlet.security.roles"
	// AttributeRunAsRole 保存当前 Servlet 执行身份角色。
	AttributeRunAsRole = "arkarta.servlet.security.run_as_role"
)

// Principal 表示认证主体。
type Principal interface {
	Name() string
}

// PrincipalFunc 将函数适配为 Principal。
type PrincipalFunc func() string

// Name 返回主体名称。
func (f PrincipalFunc) Name() string {
	return f()
}

// SetPrincipal 绑定当前请求的认证主体。
func SetPrincipal(req *servlet.Request, principal Principal, authType string, roles ...string) {
	if req == nil {
		return
	}
	req.SetAttribute(AttributePrincipal, principal)
	req.SetAttribute(AttributeAuthType, authType)
	roleSet := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		if role != "" {
			roleSet[role] = struct{}{}
		}
	}
	req.SetAttribute(AttributeRoles, roleSet)
}

// BindIdentity 将认证身份写入当前请求。
func BindIdentity(req *servlet.Request, identity Identity) {
	if !identity.Valid() {
		return
	}
	SetPrincipal(req, identity.Principal(), identity.AuthType(), identity.Roles()...)
}

// ClearPrincipal 清理当前请求认证上下文。
func ClearPrincipal(req *servlet.Request) {
	if req == nil {
		return
	}
	req.SetAttribute(AttributePrincipal, nil)
	req.SetAttribute(AttributeAuthType, nil)
	req.SetAttribute(AttributeRoles, nil)
}

// CurrentPrincipal 返回当前请求认证主体。
func CurrentPrincipal(req *servlet.Request) (Principal, bool) {
	if req == nil {
		return nil, false
	}
	value, ok := req.Attribute(AttributePrincipal)
	if !ok {
		return nil, false
	}
	principal, ok := value.(Principal)
	return principal, ok && principal != nil
}

// AuthType 返回当前认证方式。
func AuthType(req *servlet.Request) string {
	if req == nil {
		return ""
	}
	value, _ := req.Attribute(AttributeAuthType)
	authType, _ := value.(string)
	return authType
}

// RemoteUser 返回当前主体名称。
func RemoteUser(req *servlet.Request) string {
	principal, ok := CurrentPrincipal(req)
	if !ok {
		return ""
	}
	return principal.Name()
}

// UserInRole 判断当前主体是否拥有指定角色。
func UserInRole(req *servlet.Request, role string) bool {
	if req == nil || role == "" {
		return false
	}
	if RunAsRole(req) == role {
		return true
	}
	value, ok := req.Attribute(AttributeRoles)
	if !ok {
		return false
	}
	roles, ok := value.(map[string]struct{})
	if !ok {
		return false
	}
	_, ok = roles[role]
	return ok
}

// RunAsRole 返回当前 Servlet 执行身份角色。
func RunAsRole(req *servlet.Request) string {
	if req == nil {
		return ""
	}
	value, _ := req.Attribute(AttributeRunAsRole)
	role, _ := value.(string)
	return role
}

// RunAs 在指定执行身份下运行函数，并在返回后恢复旧身份。
func RunAs(req *servlet.Request, role string, fn func() error) error {
	if req == nil || role == "" {
		if fn == nil {
			return nil
		}
		return fn()
	}
	previous, hadPrevious := req.Attribute(AttributeRunAsRole)
	req.SetAttribute(AttributeRunAsRole, role)
	defer func() {
		if hadPrevious {
			req.SetAttribute(AttributeRunAsRole, previous)
			return
		}
		req.SetAttribute(AttributeRunAsRole, nil)
	}()
	if fn == nil {
		return nil
	}
	return fn()
}
