package web

import (
	"errors"
	"mime"
	"strings"

	arkjson "goark.dev/arkarta/json"
	"goark.dev/arkarta/validation"
)

// BindJSON 将请求体按 JSON 解码到目标对象。
func (c *Context) BindJSON(target any) error {
	if c == nil {
		return ErrNilContext
	}
	if c.request == nil {
		return newBindError(ErrNilContext)
	}
	if err := ensureJSONContentType(c.request.Header().Get("Content-Type")); err != nil {
		return err
	}
	if err := arkjson.Decode(c.JSONCodec(), c.request.Body(), target); err != nil {
		if errors.Is(err, arkjson.ErrPayloadTooLarge) {
			return err
		}
		if errors.Is(err, arkjson.ErrNilTarget) || errors.Is(err, arkjson.ErrNilReader) {
			return err
		}
		return newBindError(err)
	}
	return nil
}

// Validate 使用当前校验器校验目标对象。
func (c *Context) Validate(target any) (validation.Result, error) {
	if c == nil {
		return validation.Result{}, ErrNilContext
	}
	return c.Validator().Validate(c.Context(), target)
}

// ValidateGroups 使用当前校验器按显式分组校验目标对象。
func (c *Context) ValidateGroups(target any, groups ...string) (validation.Result, error) {
	if c == nil {
		return validation.Result{}, ErrNilContext
	}
	if len(groups) == 0 {
		return c.Validate(target)
	}
	groupValidator, ok := c.Validator().(validation.GroupValidator)
	if !ok {
		return validation.Result{}, validation.ErrUnsupportedGroups
	}
	return groupValidator.ValidateGroups(c.Context(), target, groups...)
}

// BindAndValidateJSON 先绑定 JSON 请求体，再执行结构体验证。
func (c *Context) BindAndValidateJSON(target any) error {
	if err := c.BindJSON(target); err != nil {
		return err
	}
	result, err := c.Validate(target)
	if err != nil {
		return err
	}
	return result.Error()
}

// BindAndValidateJSONGroups 先绑定 JSON 请求体，再按显式分组执行结构体验证。
func (c *Context) BindAndValidateJSONGroups(target any, groups ...string) error {
	if err := c.BindJSON(target); err != nil {
		return err
	}
	result, err := c.ValidateGroups(target, groups...)
	if err != nil {
		return err
	}
	return result.Error()
}

func ensureJSONContentType(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return ErrUnsupportedMediaType
	}
	parts := strings.SplitN(strings.ToLower(mediaType), "/", 2)
	if len(parts) != 2 {
		return ErrUnsupportedMediaType
	}
	if parts[0] == "application" && parts[1] == "json" {
		return nil
	}
	if strings.HasSuffix(parts[1], "+json") {
		return nil
	}
	return ErrUnsupportedMediaType
}
