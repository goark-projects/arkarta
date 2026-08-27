package validation

import (
	"context"
	"fmt"
	"net/mail"
	"reflect"
	"strconv"
	"strings"
)

func builtinConstraints() []Constraint {
	return []Constraint{
		ConstraintFunc{RuleName: "required", Fn: validateRequired},
		ConstraintFunc{RuleName: "notblank", Fn: validateNotBlank},
		ConstraintFunc{RuleName: "min", Fn: validateMin},
		ConstraintFunc{RuleName: "max", Fn: validateMax},
		ConstraintFunc{RuleName: "len", Fn: validateLen},
		ConstraintFunc{RuleName: "email", Fn: validateEmail},
		ConstraintFunc{RuleName: "oneof", Fn: validateOneOf},
	}
}

func validateRequired(_ context.Context, field FieldContext) (Violation, bool, error) {
	if isEmptyValue(field.Value()) {
		return NewViolation(field.Path(), "required", "不能为空", interfaceValue(field.Value())), true, nil
	}
	return Violation{}, false, nil
}

func validateNotBlank(_ context.Context, field FieldContext) (Violation, bool, error) {
	if isBlankString(field.Value()) {
		return NewViolation(field.Path(), "notblank", "不能为空白字符串", interfaceValue(field.Value())), true, nil
	}
	return Violation{}, false, nil
}

func validateMin(_ context.Context, field FieldContext) (Violation, bool, error) {
	if !unwrapValue(field.Value()).IsValid() {
		return Violation{}, false, nil
	}
	limit, err := strconv.ParseFloat(field.Rule().Param(), 64)
	if err != nil {
		return Violation{}, false, fmt.Errorf("%w: min", ErrInvalidRule)
	}
	if length, ok := lengthValue(field.Value()); ok {
		if float64(length) < limit {
			return NewViolation(field.Path(), "min", "长度不能小于 "+field.Rule().Param(), interfaceValue(field.Value())), true, nil
		}
		return Violation{}, false, nil
	}
	value, ok := numericValue(field.Value())
	if !ok {
		return Violation{}, false, fmt.Errorf("%w: min requires number or sized value", ErrInvalidRule)
	}
	if value < limit {
		return NewViolation(field.Path(), "min", "数值不能小于 "+field.Rule().Param(), interfaceValue(field.Value())), true, nil
	}
	return Violation{}, false, nil
}

func validateMax(_ context.Context, field FieldContext) (Violation, bool, error) {
	if !unwrapValue(field.Value()).IsValid() {
		return Violation{}, false, nil
	}
	limit, err := strconv.ParseFloat(field.Rule().Param(), 64)
	if err != nil {
		return Violation{}, false, fmt.Errorf("%w: max", ErrInvalidRule)
	}
	if length, ok := lengthValue(field.Value()); ok {
		if float64(length) > limit {
			return NewViolation(field.Path(), "max", "长度不能大于 "+field.Rule().Param(), interfaceValue(field.Value())), true, nil
		}
		return Violation{}, false, nil
	}
	value, ok := numericValue(field.Value())
	if !ok {
		return Violation{}, false, fmt.Errorf("%w: max requires number or sized value", ErrInvalidRule)
	}
	if value > limit {
		return NewViolation(field.Path(), "max", "数值不能大于 "+field.Rule().Param(), interfaceValue(field.Value())), true, nil
	}
	return Violation{}, false, nil
}

func validateLen(_ context.Context, field FieldContext) (Violation, bool, error) {
	if !unwrapValue(field.Value()).IsValid() {
		return Violation{}, false, nil
	}
	limit, err := strconv.Atoi(field.Rule().Param())
	if err != nil {
		return Violation{}, false, fmt.Errorf("%w: len", ErrInvalidRule)
	}
	length, ok := lengthValue(field.Value())
	if !ok {
		return Violation{}, false, fmt.Errorf("%w: len requires sized value", ErrInvalidRule)
	}
	if length != limit {
		return NewViolation(field.Path(), "len", "长度必须等于 "+field.Rule().Param(), interfaceValue(field.Value())), true, nil
	}
	return Violation{}, false, nil
}

func validateEmail(_ context.Context, field FieldContext) (Violation, bool, error) {
	value := unwrapValue(field.Value())
	if !value.IsValid() || value.Kind() != reflect.String || value.String() == "" {
		return Violation{}, false, nil
	}
	if _, err := mail.ParseAddress(value.String()); err != nil {
		return NewViolation(field.Path(), "email", "必须是合法邮箱地址", interfaceValue(field.Value())), true, nil
	}
	return Violation{}, false, nil
}

func validateOneOf(_ context.Context, field FieldContext) (Violation, bool, error) {
	if !unwrapValue(field.Value()).IsValid() {
		return Violation{}, false, nil
	}
	candidates := strings.Split(field.Rule().Param(), "|")
	value := fmt.Sprint(interfaceValue(field.Value()))
	for _, candidate := range candidates {
		if value == strings.TrimSpace(candidate) {
			return Violation{}, false, nil
		}
	}
	return NewViolation(field.Path(), "oneof", "必须属于允许值集合", interfaceValue(field.Value())), true, nil
}
