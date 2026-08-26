package security

import (
	"context"
	"net/http"
	"sort"

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
	emptyRoleSemantic  EmptyRoleSemantic
	transportGuarantee TransportGuarantee
}

// ConstraintOption 定制安全约束。
type ConstraintOption func(*Constraint)

// NewConstraint 创建安全约束。
func NewConstraint(options ...ConstraintOption) Constraint {
	constraint := Constraint{
		roles:             make(map[string]struct{}),
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

// Clone 返回安全约束深拷贝。
func (c Constraint) Clone() Constraint {
	clone := Constraint{
		roles:              make(map[string]struct{}, len(c.roles)),
		emptyRoleSemantic:  c.emptyRoleSemantic,
		transportGuarantee: c.transportGuarantee,
	}
	for role := range c.roles {
		clone.roles[role] = struct{}{}
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
		if UserInRole(req, role) {
			return nil
		}
	}
	return servlet.NewHTTPError(http.StatusForbidden, "access denied", nil)
}
