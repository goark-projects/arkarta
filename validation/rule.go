package validation

import (
	"context"
	"reflect"
)

// Rule 表示一个字段级校验规则。
type Rule struct {
	name  string
	param string
}

// Name 返回规则名称。
func (r Rule) Name() string {
	return r.name
}

// Param 返回规则参数。
func (r Rule) Param() string {
	return r.param
}

// FieldContext 表示字段校验上下文。
type FieldContext struct {
	path  string
	field reflect.StructField
	value reflect.Value
	rule  Rule
}

// Path 返回字段路径。
func (c FieldContext) Path() string {
	return c.path
}

// Field 返回结构体字段元信息。
func (c FieldContext) Field() reflect.StructField {
	return c.field
}

// Value 返回字段值。
func (c FieldContext) Value() reflect.Value {
	return c.value
}

// Rule 返回当前执行的校验规则。
func (c FieldContext) Rule() Rule {
	return c.rule
}

// Constraint 表示一个可注册的字段级约束。
type Constraint interface {
	Name() string
	Validate(ctx context.Context, field FieldContext) (Violation, bool, error)
}

// ConstraintFunc 将函数适配为校验约束。
type ConstraintFunc struct {
	RuleName string
	Fn       func(ctx context.Context, field FieldContext) (Violation, bool, error)
}

// Name 返回约束名称。
func (f ConstraintFunc) Name() string {
	return f.RuleName
}

// Validate 执行约束函数。
func (f ConstraintFunc) Validate(ctx context.Context, field FieldContext) (Violation, bool, error) {
	if f.Fn == nil {
		return Violation{}, false, nil
	}
	return f.Fn(ctx, field)
}
