package web

import (
	arkjson "goark.dev/arkarta/json"
	"goark.dev/arkarta/validation"
)

// Option 定制 Web Router。
type Option func(*Router)

// WithJSONCodec 设置请求和响应使用的 JSON 编解码器。
func WithJSONCodec(codec arkjson.Codec) Option {
	return func(router *Router) {
		if codec != nil {
			router.codec = codec
		}
	}
}

// WithValidator 设置 Web 层使用的校验器。
func WithValidator(validator validation.Validator) Option {
	return func(router *Router) {
		if validator != nil {
			router.validator = validator
		}
	}
}

// WithErrorMapper 设置统一错误响应映射器。
func WithErrorMapper(mapper ErrorMapper) Option {
	return func(router *Router) {
		if mapper != nil {
			router.errorMapper = mapper
		}
	}
}

// WithInterceptor 注册初始拦截器。
func WithInterceptor(interceptor Interceptor) Option {
	return func(router *Router) {
		router.Use(interceptor)
	}
}

// WithResponseAdvice 注册初始响应增强器。
func WithResponseAdvice(advice ResponseAdvice) Option {
	return func(router *Router) {
		router.UseResponseAdvice(advice)
	}
}
