package validation

import (
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

var timeType = reflect.TypeOf(time.Time{})

func indirectValue(value any) (reflect.Value, bool) {
	if value == nil {
		return reflect.Value{}, false
	}
	current := reflect.ValueOf(value)
	for current.Kind() == reflect.Pointer || current.Kind() == reflect.Interface {
		if current.IsNil() {
			return reflect.Value{}, false
		}
		current = current.Elem()
	}
	return current, true
}

func isEmptyValue(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return true
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Struct:
		return value.IsZero()
	default:
		return value.IsZero()
	}
}

func isBlankString(value reflect.Value) bool {
	value = unwrapValue(value)
	return value.Kind() == reflect.String && strings.TrimSpace(value.String()) == ""
}

func unwrapValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func numericValue(value reflect.Value) (float64, bool) {
	value = unwrapValue(value)
	if !value.IsValid() {
		return 0, false
	}
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(value.Uint()), true
	case reflect.Float32, reflect.Float64:
		return value.Float(), true
	default:
		return 0, false
	}
}

func lengthValue(value reflect.Value) (int, bool) {
	value = unwrapValue(value)
	if !value.IsValid() {
		return 0, false
	}
	switch value.Kind() {
	case reflect.String:
		return utf8.RuneCountInString(value.String()), true
	case reflect.Array, reflect.Map, reflect.Slice:
		return value.Len(), true
	default:
		return 0, false
	}
}

func isStructValue(value reflect.Value) bool {
	value = unwrapValue(value)
	return value.IsValid() && value.Kind() == reflect.Struct && value.Type() != timeType
}

func interfaceValue(value reflect.Value) any {
	if !value.IsValid() || !value.CanInterface() {
		return nil
	}
	return value.Interface()
}
