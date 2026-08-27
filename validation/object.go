package validation

import (
	"context"
	"reflect"
)

// ObjectContext 表示对象级约束的校验上下文。
type ObjectContext struct {
	path  string
	value reflect.Value
}

// Path 返回当前对象路径。
func (c ObjectContext) Path() string {
	return c.path
}

// FieldPath 返回当前对象下的字段路径。
func (c ObjectContext) FieldPath(name string) string {
	return fieldPath(c.path, name)
}

// Value 返回当前对象反射值。
func (c ObjectContext) Value() reflect.Value {
	return c.value
}

// Interface 返回当前对象值。
func (c ObjectContext) Interface() any {
	return interfaceValue(c.value)
}

// ObjectConstraint 表示对象级约束。
type ObjectConstraint interface {
	Name() string
	ValidateObject(ctx context.Context, object ObjectContext) ([]Violation, error)
}

// ObjectConstraintFunc 将普通函数适配为对象级约束。
type ObjectConstraintFunc struct {
	ConstraintName string
	Fn             func(ctx context.Context, object ObjectContext) ([]Violation, error)
}

// Name 返回约束名称。
func (f ObjectConstraintFunc) Name() string {
	return f.ConstraintName
}

// ValidateObject 执行对象级约束函数。
func (f ObjectConstraintFunc) ValidateObject(ctx context.Context, object ObjectContext) ([]Violation, error) {
	if f.Fn == nil {
		return nil, nil
	}
	return f.Fn(ctx, object)
}

type objectConstraintRegistration struct {
	constraint ObjectConstraint
	groups     []string
}

func (v *DefaultValidator) registerObjectConstraint(sample any, constraint ObjectConstraint, groups []string) {
	if constraint == nil || constraint.Name() == "" {
		return
	}
	objectType, ok := objectSampleType(sample)
	if !ok {
		return
	}
	v.objectConstraints[objectType] = append(v.objectConstraints[objectType], objectConstraintRegistration{
		constraint: constraint,
		groups:     normalizeGroups(groups),
	})
}

func objectSampleType(sample any) (reflect.Type, bool) {
	if sample == nil {
		return nil, false
	}
	valueType := reflect.TypeOf(sample)
	for valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}
	if valueType.Kind() != reflect.Struct {
		return nil, false
	}
	return valueType, true
}
