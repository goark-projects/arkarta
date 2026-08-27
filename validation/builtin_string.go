package validation

import (
	"context"
	"reflect"
	"strings"
)

func validateContains(_ context.Context, field FieldContext) (Violation, bool, error) {
	return validateStringPredicate(field, "contains", "必须包含指定片段", func(value, param string) bool {
		return strings.Contains(value, param)
	})
}

func validateStartsWith(_ context.Context, field FieldContext) (Violation, bool, error) {
	return validateStringPredicate(field, "startswith", "必须以指定片段开头", func(value, param string) bool {
		return strings.HasPrefix(value, param)
	})
}

func validateEndsWith(_ context.Context, field FieldContext) (Violation, bool, error) {
	return validateStringPredicate(field, "endswith", "必须以指定片段结尾", func(value, param string) bool {
		return strings.HasSuffix(value, param)
	})
}

func validateStringPredicate(field FieldContext, code, message string, pass func(value, param string) bool) (Violation, bool, error) {
	value := unwrapValue(field.Value())
	if !value.IsValid() || value.Kind() != reflect.String || value.String() == "" {
		return Violation{}, false, nil
	}
	if pass(value.String(), field.Rule().Param()) {
		return Violation{}, false, nil
	}
	return NewViolation(field.Path(), code, message, interfaceValue(field.Value())), true, nil
}
