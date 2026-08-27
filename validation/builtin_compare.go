package validation

import (
	"context"
	"fmt"
	"strconv"
)

func validateGT(_ context.Context, field FieldContext) (Violation, bool, error) {
	return validateComparison(field, "gt", "数值必须大于 "+field.Rule().Param(), func(value, limit float64) bool {
		return value > limit
	})
}

func validateGTE(_ context.Context, field FieldContext) (Violation, bool, error) {
	return validateComparison(field, "gte", "数值必须大于等于 "+field.Rule().Param(), func(value, limit float64) bool {
		return value >= limit
	})
}

func validateLT(_ context.Context, field FieldContext) (Violation, bool, error) {
	return validateComparison(field, "lt", "数值必须小于 "+field.Rule().Param(), func(value, limit float64) bool {
		return value < limit
	})
}

func validateLTE(_ context.Context, field FieldContext) (Violation, bool, error) {
	return validateComparison(field, "lte", "数值必须小于等于 "+field.Rule().Param(), func(value, limit float64) bool {
		return value <= limit
	})
}

func validateComparison(field FieldContext, code, message string, pass func(value, limit float64) bool) (Violation, bool, error) {
	if !unwrapValue(field.Value()).IsValid() {
		return Violation{}, false, nil
	}
	limit, err := strconv.ParseFloat(field.Rule().Param(), 64)
	if err != nil {
		return Violation{}, false, fmt.Errorf("%w: %s", ErrInvalidRule, code)
	}
	value, ok := numericValue(field.Value())
	if !ok {
		return Violation{}, false, fmt.Errorf("%w: %s requires number", ErrInvalidRule, code)
	}
	if pass(value, limit) {
		return Violation{}, false, nil
	}
	return NewViolation(field.Path(), code, message, interfaceValue(field.Value())), true, nil
}
