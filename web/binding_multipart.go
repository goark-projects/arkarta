package web

import (
	"fmt"
	"mime"
	"reflect"
	"strings"

	servletmultipart "goark.dev/arkarta/servlet/multipart"
)

var multipartPartType = reflect.TypeOf(servletmultipart.Part{})

// MultipartForm 解析并返回 multipart/form-data 表单。
func (c *Context) MultipartForm(options ...servletmultipart.Option) (*servletmultipart.Form, error) {
	if c == nil {
		return nil, ErrNilContext
	}
	if c.request == nil {
		return nil, newBindError(ErrNilContext)
	}
	if err := ensureMultipartContentType(c.request.Header().Get("Content-Type")); err != nil {
		return nil, err
	}
	parser := servletmultipart.NewParser(options...)
	form, err := servletmultipart.ParseRequest(c.request, parser)
	if err != nil {
		return nil, newBindError(err)
	}
	return form, nil
}

// BindMultipart 将 multipart 普通字段和文件字段绑定到结构体。
func (c *Context) BindMultipart(target any, options ...servletmultipart.Option) error {
	form, err := c.MultipartForm(options...)
	if err != nil {
		return err
	}
	binder, err := newStructBinder(target)
	if err != nil {
		return newBindError(err)
	}
	if err := binder.bindValues(form.Values()); err != nil {
		return newBindError(err)
	}
	if err := bindMultipartParts(binder, form); err != nil {
		return newBindError(err)
	}
	return nil
}

func ensureMultipartContentType(value string) error {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil {
		return ErrUnsupportedMediaType
	}
	if strings.EqualFold(mediaType, "multipart/form-data") {
		return nil
	}
	return ErrUnsupportedMediaType
}

func bindMultipartParts(binder structBinder, form *servletmultipart.Form) error {
	return walkStructFields(binder.value, func(field reflect.StructField, value reflect.Value) error {
		name, ok := bindingName(field, "multipart")
		if !ok || !value.CanSet() || !isMultipartTarget(value.Type()) {
			return nil
		}
		parts := partsByName(form, name)
		if len(parts) == 0 {
			return nil
		}
		if err := setMultipartField(value, parts); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	})
}

func isMultipartTarget(targetType reflect.Type) bool {
	if targetType == multipartPartType {
		return true
	}
	if targetType.Kind() == reflect.Slice && targetType.Elem() == multipartPartType {
		return true
	}
	if targetType.Kind() == reflect.Pointer {
		return isMultipartTarget(targetType.Elem())
	}
	return false
}

func partsByName(form *servletmultipart.Form, name string) []servletmultipart.Part {
	if form == nil {
		return nil
	}
	parts := form.Parts()
	result := make([]servletmultipart.Part, 0, len(parts))
	for _, part := range parts {
		if part.Name() == name {
			result = append(result, part)
		}
	}
	return result
}

func setMultipartField(field reflect.Value, parts []servletmultipart.Part) error {
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setMultipartField(field.Elem(), parts)
	}
	if field.Type() == multipartPartType {
		field.Set(reflect.ValueOf(parts[0]))
		return nil
	}
	if field.Kind() == reflect.Slice && field.Type().Elem() == multipartPartType {
		slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
		for index, part := range parts {
			slice.Index(index).Set(reflect.ValueOf(part))
		}
		field.Set(slice)
		return nil
	}
	return fmt.Errorf("unsupported multipart field type %s", field.Type())
}
