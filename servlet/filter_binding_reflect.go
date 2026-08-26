package servlet

import "reflect"

func isNilFilter(filter Filter) bool {
	if filter == nil {
		return true
	}
	reflected := reflect.ValueOf(filter)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
