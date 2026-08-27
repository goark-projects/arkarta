package validation

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
)

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func validateRegexp(_ context.Context, field FieldContext) (Violation, bool, error) {
	value := unwrapValue(field.Value())
	if !value.IsValid() || value.Kind() != reflect.String || value.String() == "" {
		return Violation{}, false, nil
	}
	pattern, err := regexp.Compile(field.Rule().Param())
	if err != nil {
		return Violation{}, false, fmt.Errorf("%w: regexp", ErrInvalidRule)
	}
	if !pattern.MatchString(value.String()) {
		return NewViolation(field.Path(), "regexp", "格式不符合正则表达式", interfaceValue(field.Value())), true, nil
	}
	return Violation{}, false, nil
}

func validateURL(_ context.Context, field FieldContext) (Violation, bool, error) {
	value := unwrapValue(field.Value())
	if !value.IsValid() || value.Kind() != reflect.String || value.String() == "" {
		return Violation{}, false, nil
	}
	parsed, err := url.ParseRequestURI(value.String())
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return NewViolation(field.Path(), "url", "必须是合法 URL", interfaceValue(field.Value())), true, nil
	}
	return Violation{}, false, nil
}

func validateUUID(_ context.Context, field FieldContext) (Violation, bool, error) {
	value := unwrapValue(field.Value())
	if !value.IsValid() || value.Kind() != reflect.String || value.String() == "" {
		return Violation{}, false, nil
	}
	if !uuidPattern.MatchString(value.String()) {
		return NewViolation(field.Path(), "uuid", "必须是合法 UUID", interfaceValue(field.Value())), true, nil
	}
	return Violation{}, false, nil
}
