package web

import (
	"mime"
	"net/http"
	"net/url"
	"strings"
)

const formContentType = "application/x-www-form-urlencoded"

// BindForm 将查询串和 application/x-www-form-urlencoded 表单绑定到结构体。
func (c *Context) BindForm(target any) error {
	if c == nil {
		return ErrNilContext
	}
	if c.request == nil {
		return newBindError(ErrNilContext)
	}
	if err := ensureFormContentType(c.request.Method(), c.request.Header().Get("Content-Type")); err != nil {
		return err
	}
	values, err := c.request.Parameters()
	if err != nil {
		return newBindError(err)
	}
	if err := bindURLValues(target, values); err != nil {
		return newBindError(err)
	}
	return nil
}

func ensureFormContentType(method, value string) error {
	value = strings.TrimSpace(value)
	if value == "" || !methodAllowsBody(method) {
		return nil
	}
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return ErrUnsupportedMediaType
	}
	if strings.EqualFold(mediaType, formContentType) {
		return nil
	}
	return ErrUnsupportedMediaType
}

func methodAllowsBody(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

func bindURLValues(target any, values url.Values) error {
	binder, err := newStructBinder(target)
	if err != nil {
		return err
	}
	return binder.bindValues(values)
}
