package web

import (
	"net/http"

	arkjson "goark.dev/arkarta/json"
	"goark.dev/arkarta/servlet"
)

const textContentType = "text/plain; charset=utf-8"

// Result 表示可以写入 Servlet 响应的 Web 结果。
type Result interface {
	Write(ctx *Context) error
}

type resultFunc func(ctx *Context) error

func (f resultFunc) Write(ctx *Context) error {
	return f(ctx)
}

// JSON 创建 JSON 响应结果。
func JSON(statusCode int, value any) Result {
	return jsonResult(statusCode, value, true)
}

// Text 创建纯文本响应结果。
func Text(statusCode int, value string) Result {
	return resultFunc(func(ctx *Context) error {
		if err := ensureWritableContext(ctx); err != nil {
			return err
		}
		if ok := accepts(ctx, "text/plain"); !ok {
			return servlet.NewHTTPError(http.StatusNotAcceptable, http.StatusText(http.StatusNotAcceptable), nil)
		}
		if err := servlet.SetContentType(ctx.response, textContentType); err != nil {
			return err
		}
		ctx.response.SetStatus(normalizeStatus(statusCode, http.StatusOK))
		_, err := ctx.response.WriteString(value)
		return err
	})
}

// NoContent 创建无响应体结果。
func NoContent() Result {
	return resultFunc(func(ctx *Context) error {
		if err := ensureWritableContext(ctx); err != nil {
			return err
		}
		ctx.response.SetStatus(http.StatusNoContent)
		return nil
	})
}

func errorJSON(statusCode int, value any) Result {
	return jsonResult(statusCode, value, false)
}

func jsonResult(statusCode int, value any, negotiate bool) Result {
	return resultFunc(func(ctx *Context) error {
		if err := ensureWritableContext(ctx); err != nil {
			return err
		}
		if negotiate {
			if ok := accepts(ctx, arkjson.ContentType); !ok {
				return servlet.NewHTTPError(http.StatusNotAcceptable, http.StatusText(http.StatusNotAcceptable), nil)
			}
		}
		if err := servlet.SetContentType(ctx.response, arkjson.ContentType); err != nil {
			return err
		}
		ctx.response.SetStatus(normalizeStatus(statusCode, http.StatusOK))
		return arkjson.Encode(ctx.JSONCodec(), ctx.response.BodyWriter(), value)
	})
}

func accepts(ctx *Context, contentType string) bool {
	if ctx == nil || ctx.request == nil {
		return true
	}
	_, ok := ctx.request.NegotiateContentType(contentType)
	return ok
}

func ensureWritableContext(ctx *Context) error {
	if ctx == nil || ctx.response == nil {
		return ErrNilContext
	}
	return nil
}

func normalizeStatus(statusCode, fallback int) int {
	if statusCode == 0 {
		return fallback
	}
	if statusCode < 100 || statusCode > 999 {
		return http.StatusInternalServerError
	}
	return statusCode
}
