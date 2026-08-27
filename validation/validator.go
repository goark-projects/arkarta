package validation

import (
	"context"
	"fmt"
	"reflect"
)

// Validator 表示 Arkarta 校验标准入口。
type Validator interface {
	Validate(ctx context.Context, value any) (Result, error)
}

// GroupValidator 表示支持显式校验分组的验证器。
type GroupValidator interface {
	Validator
	ValidateGroups(ctx context.Context, value any, groups ...string) (Result, error)
}

// DefaultValidator 基于结构体标签执行校验。
type DefaultValidator struct {
	tagName           string
	groupTagName      string
	constraints       map[string]Constraint
	objectConstraints map[reflect.Type][]objectConstraintRegistration
	messageResolver   MessageResolver
}

// NewValidator 创建默认结构体验证器。
func NewValidator(options ...Option) *DefaultValidator {
	validator := &DefaultValidator{
		tagName:           defaultTagName,
		groupTagName:      defaultGroupTagName,
		constraints:       make(map[string]Constraint),
		objectConstraints: make(map[reflect.Type][]objectConstraintRegistration),
	}
	for _, constraint := range builtinConstraints() {
		validator.register(constraint)
	}
	for _, option := range options {
		if option != nil {
			option(validator)
		}
	}
	return validator
}

// Validate 使用默认验证器校验目标。
func Validate(ctx context.Context, value any) (Result, error) {
	return NewValidator().Validate(ctx, value)
}

// ValidateGroups 使用默认验证器按指定分组校验目标。
func ValidateGroups(ctx context.Context, value any, groups ...string) (Result, error) {
	return NewValidator().ValidateGroups(ctx, value, groups...)
}

// Validate 校验结构体、结构体指针或结构体切片。
func (v *DefaultValidator) Validate(ctx context.Context, value any) (Result, error) {
	return v.ValidateGroups(ctx, value, DefaultGroup)
}

// ValidateGroups 校验结构体、结构体指针或结构体切片。
func (v *DefaultValidator) ValidateGroups(ctx context.Context, value any, groups ...string) (Result, error) {
	if v == nil {
		return Result{}, ErrNilValidator
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	root, ok := indirectValue(value)
	if !ok {
		return Result{}, ErrNilValue
	}
	activeGroups := newGroupSet(groups...)
	var violations []Violation
	if err := v.validateValue(ctx, "", root, activeGroups, &violations); err != nil {
		return Result{}, err
	}
	return NewResult(violations...), nil
}

func (v *DefaultValidator) register(constraint Constraint) {
	if constraint == nil || constraint.Name() == "" {
		return
	}
	v.constraints[constraint.Name()] = constraint
}

func (v *DefaultValidator) validateValue(ctx context.Context, path string, value reflect.Value, groups groupSet, violations *[]Violation) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	value = unwrapValue(value)
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.Struct:
		return v.validateStruct(ctx, path, value, groups, violations)
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			if path == "" {
				itemPath = fmt.Sprintf("[%d]", i)
			}
			if !isStructValue(value.Index(i)) {
				continue
			}
			if err := v.validateValue(ctx, itemPath, value.Index(i), groups, violations); err != nil {
				return err
			}
		}
		return nil
	default:
		return ErrUnsupportedValue
	}
}

func (v *DefaultValidator) validateStruct(ctx context.Context, path string, value reflect.Value, groups groupSet, violations *[]Violation) error {
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := valueType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := fieldName(field.Name, field.Tag.Get("json"))
		if name == "" {
			continue
		}
		fieldValue := value.Field(i)
		currentPath := fieldPath(path, name)
		rules, err := parseRules(field.Tag.Get(v.tagName))
		if err != nil {
			return fmt.Errorf("%w: %s", err, currentPath)
		}
		if parseGroupTag(field.Tag.Get(v.groupTagName)).activeIn(groups) {
			for _, rule := range rules {
				constraint, ok := v.constraints[rule.Name()]
				if !ok {
					return fmt.Errorf("%w: %s", ErrInvalidRule, rule.Name())
				}
				fieldCtx := FieldContext{
					path:  currentPath,
					field: field,
					value: fieldValue,
					rule:  rule,
				}
				violation, failed, err := constraint.Validate(ctx, fieldCtx)
				if err != nil {
					return fmt.Errorf("%w: %s", err, currentPath)
				}
				if failed {
					*violations = append(*violations, v.resolveMessage(ctx, violation, fieldCtx))
				}
			}
		}
		if isStructValue(fieldValue) {
			if err := v.validateValue(ctx, currentPath, fieldValue, groups, violations); err != nil {
				return err
			}
			continue
		}
		fieldValue = unwrapValue(fieldValue)
		if fieldValue.IsValid() && (fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Array) {
			if err := v.validateValue(ctx, currentPath, fieldValue, groups, violations); err != nil {
				return err
			}
		}
	}
	return v.validateObjectConstraints(ctx, path, value, groups, violations)
}

func (v *DefaultValidator) validateObjectConstraints(ctx context.Context, path string, value reflect.Value, groups groupSet, violations *[]Violation) error {
	registrations := v.objectConstraints[value.Type()]
	for _, registration := range registrations {
		if !newGroupSet(registration.groups...).activeIn(groups) {
			continue
		}
		result, err := registration.constraint.ValidateObject(ctx, ObjectContext{
			path:  path,
			value: value,
		})
		if err != nil {
			return fmt.Errorf("%w: %s", err, registration.constraint.Name())
		}
		*violations = append(*violations, result...)
	}
	return nil
}
