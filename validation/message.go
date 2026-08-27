package validation

import "context"

// MessageContext 表示一次校验失败消息解析上下文。
type MessageContext struct {
	path           string
	code           string
	defaultMessage string
	param          string
	field          FieldContext
}

// Path 返回字段路径。
func (c MessageContext) Path() string {
	return c.path
}

// Code 返回机器可读错误码。
func (c MessageContext) Code() string {
	return c.code
}

// DefaultMessage 返回约束产生的默认消息。
func (c MessageContext) DefaultMessage() string {
	return c.defaultMessage
}

// Param 返回规则参数。
func (c MessageContext) Param() string {
	return c.param
}

// Field 返回字段校验上下文。
func (c MessageContext) Field() FieldContext {
	return c.field
}

// MessageResolver 按上下文解析校验失败消息。
type MessageResolver interface {
	ResolveMessage(ctx context.Context, message MessageContext) (string, bool)
}

// MessageResolverFunc 将普通函数适配为 MessageResolver。
type MessageResolverFunc func(ctx context.Context, message MessageContext) (string, bool)

// ResolveMessage 执行底层消息解析函数。
func (f MessageResolverFunc) ResolveMessage(ctx context.Context, message MessageContext) (string, bool) {
	if f == nil {
		return "", false
	}
	return f(ctx, message)
}

func (v *DefaultValidator) resolveMessage(ctx context.Context, violation Violation, field FieldContext) Violation {
	if v.messageResolver == nil {
		return violation
	}
	message, ok := v.messageResolver.ResolveMessage(ctx, MessageContext{
		path:           violation.Path(),
		code:           violation.Code(),
		defaultMessage: violation.Message(),
		param:          field.Rule().Param(),
		field:          field,
	})
	if !ok {
		return violation
	}
	return violation.WithMessage(message)
}
