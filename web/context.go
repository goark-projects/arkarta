package web

import (
	"context"

	arkjson "goark.dev/arkarta/json"
	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/validation"
)

// Context 表示一次 Web 请求处理上下文。
type Context struct {
	ctx        context.Context
	request    *servlet.Request
	response   servlet.Response
	pathValues map[string]string
	codec      arkjson.Codec
	validator  validation.Validator
}

func newContext(ctx context.Context, request *servlet.Request, response servlet.Response, pathValues map[string]string, codec arkjson.Codec, validator validation.Validator) *Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if codec == nil {
		codec = arkjson.DefaultCodec()
	}
	if validator == nil {
		validator = validation.NewValidator()
	}
	return &Context{
		ctx:        ctx,
		request:    request,
		response:   response,
		pathValues: clonePathValues(pathValues),
		codec:      codec,
		validator:  validator,
	}
}

// Context 返回底层请求上下文。
func (c *Context) Context() context.Context {
	if c == nil || c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

// Request 返回 Servlet 请求。
func (c *Context) Request() *servlet.Request {
	if c == nil {
		return nil
	}
	return c.request
}

// Response 返回 Servlet 响应。
func (c *Context) Response() servlet.Response {
	if c == nil {
		return nil
	}
	return c.response
}

// PathValue 返回路径变量值；不存在时返回空字符串。
func (c *Context) PathValue(name string) string {
	value, _ := c.Param(name)
	return value
}

// Param 返回路径变量值和存在标记。
func (c *Context) Param(name string) (string, bool) {
	if c == nil {
		return "", false
	}
	value, ok := c.pathValues[name]
	return value, ok
}

// PathValues 返回路径变量副本。
func (c *Context) PathValues() map[string]string {
	if c == nil {
		return map[string]string{}
	}
	return clonePathValues(c.pathValues)
}

// QueryValue 返回查询参数第一个值。
func (c *Context) QueryValue(name string) string {
	if c == nil || c.request == nil {
		return ""
	}
	return c.request.Query().Get(name)
}

// HeaderValue 返回请求头第一个值。
func (c *Context) HeaderValue(name string) string {
	if c == nil || c.request == nil {
		return ""
	}
	return c.request.Header().Get(name)
}

// Cookie 返回请求 Cookie。
func (c *Context) Cookie(name string) (*servlet.Cookie, error) {
	if c == nil || c.request == nil {
		return nil, servlet.ErrNoCookie
	}
	return c.request.Cookie(name)
}

// JSONCodec 返回当前请求使用的 JSON 编解码器。
func (c *Context) JSONCodec() arkjson.Codec {
	if c == nil || c.codec == nil {
		return arkjson.DefaultCodec()
	}
	return c.codec
}

// Validator 返回当前请求使用的校验器。
func (c *Context) Validator() validation.Validator {
	if c == nil || c.validator == nil {
		return validation.NewValidator()
	}
	return c.validator
}

func clonePathValues(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
