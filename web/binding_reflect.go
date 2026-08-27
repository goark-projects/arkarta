package web

import (
	"encoding"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

type structBinder struct {
	value reflect.Value
}

func newStructBinder(target any) (structBinder, error) {
	if target == nil {
		return structBinder{}, ErrInvalidBindTarget
	}
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return structBinder{}, ErrInvalidBindTarget
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return structBinder{}, ErrInvalidBindTarget
	}
	return structBinder{value: value}, nil
}

func (b structBinder) bindValues(values url.Values) error {
	return walkStructFields(b.value, func(field reflect.StructField, value reflect.Value) error {
		name, ok := bindingName(field, "form")
		if !ok {
			return nil
		}
		list, exists := values[name]
		if !exists || len(list) == 0 || !value.CanSet() {
			return nil
		}
		if err := setFieldValue(value, list); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	})
}

func walkStructFields(value reflect.Value, visit func(reflect.StructField, reflect.Value) error) error {
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := valueType.Field(i)
		fieldValue := value.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if shouldRecurseField(field, fieldValue) {
			if err := walkStructFields(indirectStructValue(fieldValue), visit); err != nil {
				return err
			}
			continue
		}
		if err := visit(field, fieldValue); err != nil {
			return err
		}
	}
	return nil
}

func shouldRecurseField(field reflect.StructField, value reflect.Value) bool {
	if !field.Anonymous || field.Tag.Get("form") != "" || field.Tag.Get("multipart") != "" {
		return false
	}
	value = indirectStructValue(value)
	return value.IsValid() && value.Kind() == reflect.Struct
}

func indirectStructValue(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() && value.CanSet() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func bindingName(field reflect.StructField, primaryTag string) (string, bool) {
	if field.Tag.Get(primaryTag) == "-" {
		return "", false
	}
	if name, ok := tagName(field.Tag.Get(primaryTag)); ok {
		return name, true
	}
	if field.Tag.Get("json") == "-" {
		return "", false
	}
	if name, ok := tagName(field.Tag.Get("json")); ok {
		return name, true
	}
	return lowerFirst(field.Name), true
}

func tagName(tag string) (string, bool) {
	if tag == "-" {
		return "", false
	}
	name := strings.TrimSpace(strings.Split(tag, ",")[0])
	if name == "" {
		return "", false
	}
	return name, true
}

func lowerFirst(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToLower(value[:1]) + value[1:]
}
