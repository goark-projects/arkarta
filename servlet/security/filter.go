package security

import (
	"context"

	"goark.dev/arkarta/servlet"
)

// Filter 用安全约束保护后续处理链。
type Filter struct {
	constraint Constraint
}

// NewFilter 创建安全过滤器。
func NewFilter(constraint Constraint) *Filter {
	return &Filter{constraint: constraint}
}

// Filter 执行安全校验。
func (f *Filter) Filter(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
	if err := f.constraint.Authorize(ctx, req); err != nil {
		return err
	}
	return chain.Next(ctx, req, res)
}
