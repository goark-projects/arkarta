package registration

import "reflect"

func isNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func typeName(value any) string {
	if value == nil {
		return ""
	}
	reflected := reflect.TypeOf(value)
	if reflected.Kind() == reflect.Pointer {
		reflected = reflected.Elem()
	}
	if reflected.PkgPath() == "" || reflected.Name() == "" {
		return reflected.String()
	}
	return reflected.PkgPath() + "." + reflected.Name()
}
