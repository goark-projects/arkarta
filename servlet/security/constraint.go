package security

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"goark.dev/arkarta/servlet"
)

// EmptyRoleSemantic 表示无角色声明时的处理方式。
type EmptyRoleSemantic uint8

const (
	// EmptyRolePermit 表示无角色声明时允许访问。
	EmptyRolePermit EmptyRoleSemantic = iota
	// EmptyRoleDeny 表示无角色声明时拒绝访问。
	EmptyRoleDeny
)

// TransportGuarantee 表示传输安全要求。
type TransportGuarantee uint8

const (
	// TransportNone 表示不强制安全传输。
	TransportNone TransportGuarantee = iota
	// TransportConfidential 表示必须使用安全传输。
	TransportConfidential
)

// Constraint 描述 Servlet 安全约束。
type Constraint struct {
	roles              map[string]struct{}
	roleMappings       map[string]string
	methodConstraints  map[string]Constraint
	emptyRoleSemantic  EmptyRoleSemantic
	transportGuarantee TransportGuarantee
}

// ConstraintOption 定制安全约束。
type ConstraintOption func(*Constraint)

// NewConstraint 创建安全约束。
func NewConstraint(options ...ConstraintOption) Constraint {
	constraint := Constraint{
		roles:             make(map[string]struct{}),
		roleMappings:      make(map[string]string),
		methodConstraints: make(map[string]Constraint),
		emptyRoleSemantic: EmptyRolePermit,
	}
	for _, option := range options {
		if option != nil {
			option(&constraint)
		}
	}
	return constraint
}

// WithRoles 要求任一角色匹配。
func WithRoles(roles ...string) ConstraintOption {
	return func(constraint *Constraint) {
		for _, role := range roles {
			if role != "" {
				constraint.roles[role] = struct{}{}
			}
		}
	}
}

// WithRoleMapping 将组件内角色名映射到应用实际角色名。
func WithRoleMapping(declaredRole, actualRole string) ConstraintOption {
	return func(constraint *Constraint) {
		if declaredRole == "" || actualRole == "" {
			return
		}
		constraint.roleMappings[declaredRole] = actualRole
	}
}

// WithMethodConstraint 为指定 HTTP 方法设置专用约束。
func WithMethodConstraint(method string, child Constraint) ConstraintOption {
	return func(constraint *Constraint) {
		method = normalizeMethod(method)
		if method == "" {
			return
		}
		constraint.methodConstraints[method] = child.withoutMethodConstraints()
	}
}

// WithEmptyRoleSemantic 设置无角色声明时的处理方式。
func WithEmptyRoleSemantic(semantic EmptyRoleSemantic) ConstraintOption {
	return func(constraint *Constraint) {
		constraint.emptyRoleSemantic = semantic
	}
}

// WithTransportGuarantee 设置传输安全要求。
func WithTransportGuarantee(guarantee TransportGuarantee) ConstraintOption {
	return func(constraint *Constraint) {
		constraint.transportGuarantee = guarantee
	}
}

// Roles 返回角色集合副本。
func (c Constraint) Roles() []string {
	result := make([]string, 0, len(c.roles))
	for role := range c.roles {
		result = append(result, role)
	}
	sort.Strings(result)
	return result
}

// RoleMappings 返回角色映射副本。
func (c Constraint) RoleMappings() map[string]string {
	return cloneStringMap(c.roleMappings)
}

// MethodConstraint 返回指定 HTTP 方法的专用约束。
func (c Constraint) MethodConstraint(method string) (Constraint, bool) {
	child, ok := c.methodConstraints[normalizeMethod(method)]
	if !ok {
		return Constraint{}, false
	}
	return child.Clone(), true
}

// MethodConstraints 返回 HTTP 方法专用约束副本。
func (c Constraint) MethodConstraints() map[string]Constraint {
	if len(c.methodConstraints) == 0 {
		return map[string]Constraint{}
	}
	result := make(map[string]Constraint, len(c.methodConstraints))
	for method, child := range c.methodConstraints {
		result[method] = child.Clone()
	}
	return result
}

// Clone 返回安全约束深拷贝。
func (c Constraint) Clone() Constraint {
	clone := Constraint{
		roles:              make(map[string]struct{}, len(c.roles)),
		roleMappings:       cloneStringMap(c.roleMappings),
		methodConstraints:  make(map[string]Constraint, len(c.methodConstraints)),
		emptyRoleSemantic:  c.emptyRoleSemantic,
		transportGuarantee: c.transportGuarantee,
	}
	for role := range c.roles {
		clone.roles[role] = struct{}{}
	}
	for method, child := range c.methodConstraints {
		clone.methodConstraints[method] = child.Clone()
	}
	return clone
}

// EmptyRoleSemantic 返回无角色声明语义。
func (c Constraint) EmptyRoleSemantic() EmptyRoleSemantic {
	return c.emptyRoleSemantic
}

// TransportGuarantee 返回传输安全要求。
func (c Constraint) TransportGuarantee() TransportGuarantee {
	return c.transportGuarantee
}

// Authorize 校验请求是否满足安全约束。
func (c Constraint) Authorize(ctx context.Context, req *servlet.Request) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if req != nil {
		if child, ok := c.methodConstraint(req.Method()); ok {
			return child.authorizeBase(ctx, req)
		}
	}
	return c.authorizeBase(ctx, req)
}

func (c Constraint) authorizeBase(ctx context.Context, req *servlet.Request) error {
	if c.transportGuarantee == TransportConfidential && (req == nil || !req.IsSecure()) {
		return servlet.NewHTTPError(http.StatusForbidden, "secure transport required", nil)
	}
	if len(c.roles) == 0 {
		if c.emptyRoleSemantic == EmptyRoleDeny {
			return servlet.NewHTTPError(http.StatusForbidden, "access denied", nil)
		}
		return nil
	}
	if _, ok := CurrentPrincipal(req); !ok {
		return servlet.NewHTTPError(http.StatusUnauthorized, "authentication required", nil)
	}
	for role := range c.roles {
		if UserInRole(req, c.actualRole(role)) {
			return nil
		}
	}
	return servlet.NewHTTPError(http.StatusForbidden, "access denied", nil)
}

func (c Constraint) actualRole(role string) string {
	if actual, ok := c.roleMappings[role]; ok && actual != "" {
		return actual
	}
	return role
}

func (c Constraint) methodConstraint(method string) (Constraint, bool) {
	child, ok := c.methodConstraints[normalizeMethod(method)]
	if !ok {
		return Constraint{}, false
	}
	return child.withoutMethodConstraints(), true
}

func (c Constraint) withoutMethodConstraints() Constraint {
	clone := c.Clone()
	clone.methodConstraints = make(map[string]Constraint)
	return clone
}

func normalizeMethod(method string) string {
	return strings.ToUpper(strings.TrimSpace(method))
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
