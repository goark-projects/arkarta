package registration

import (
	"context"
	"reflect"
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
)

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

func isSessionListener(value any) bool {
	if isNil(value) {
		return false
	}
	target := reflect.TypeOf(value)
	for _, methodName := range []string{"SessionCreated", "SessionDestroyed", "SessionIDChanged"} {
		method, ok := target.MethodByName(methodName)
		if !ok || !isListenerMethod(method.Type) {
			return false
		}
	}
	return true
}

func isListenerMethod(method reflect.Type) bool {
	return method.NumIn() == 3 &&
		method.In(1).Implements(contextType) &&
		method.NumOut() == 1 &&
		method.Out(0).Implements(errorType)
}
