package web

// ResponseAdvice 表示响应结果写出前的显式增强点。
type ResponseAdvice interface {
	BeforeWrite(ctx *Context, result Result) (Result, error)
}

// ResponseAdviceFunc 将普通函数适配为 ResponseAdvice。
type ResponseAdviceFunc func(ctx *Context, result Result) (Result, error)

// BeforeWrite 执行底层响应增强函数。
func (f ResponseAdviceFunc) BeforeWrite(ctx *Context, result Result) (Result, error) {
	if f == nil {
		return result, nil
	}
	return f(ctx, result)
}

func applyResponseAdvice(ctx *Context, result Result, advice []ResponseAdvice) (Result, error) {
	var err error
	for _, item := range advice {
		if item == nil {
			continue
		}
		result, err = item.BeforeWrite(ctx, result)
		if err != nil || result == nil {
			return result, err
		}
	}
	return result, nil
}
