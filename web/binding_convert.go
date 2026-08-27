package web

import (
	"fmt"
	"reflect"
	"strconv"
)

func setFieldValue(field reflect.Value, values []string) error {
	if len(values) == 0 {
		return nil
	}
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setFieldValue(field.Elem(), values)
	}
	if field.Kind() == reflect.Slice {
		return setSliceValue(field, values)
	}
	return setScalarValue(field, values[0])
}

func setSliceValue(field reflect.Value, values []string) error {
	slice := reflect.MakeSlice(field.Type(), 0, len(values))
	for _, raw := range values {
		item := reflect.New(field.Type().Elem()).Elem()
		if err := setFieldValue(item, []string{raw}); err != nil {
			return err
		}
		slice = reflect.Append(slice, item)
	}
	field.Set(slice)
	return nil
}

func setScalarValue(field reflect.Value, raw string) error {
	if field.CanAddr() {
		addr := field.Addr()
		if addr.Type().Implements(textUnmarshalerType) {
			return addr.Interface().(interface{ UnmarshalText([]byte) error }).UnmarshalText([]byte(raw))
		}
	}
	if field.Type().Implements(textUnmarshalerType) {
		target := field.Interface().(interface{ UnmarshalText([]byte) error })
		return target.UnmarshalText([]byte(raw))
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
		return nil
	case reflect.Bool:
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		field.SetBool(value)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(raw, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(value)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		value, err := strconv.ParseUint(raw, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(value)
		return nil
	case reflect.Float32, reflect.Float64:
		value, err := strconv.ParseFloat(raw, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(value)
		return nil
	default:
		return fmt.Errorf("unsupported field type %s", field.Type())
	}
}
