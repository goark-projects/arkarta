package web

// Interceptor 表示围绕 Web Handler 执行的拦截器。
type Interceptor interface {
	Intercept(ctx *Context, next Handler) (Result, error)
}

// InterceptorFunc 将普通函数适配为拦截器。
type InterceptorFunc func(ctx *Context, next Handler) (Result, error)

// Intercept 执行底层拦截函数。
func (f InterceptorFunc) Intercept(ctx *Context, next Handler) (Result, error) {
	if f == nil {
		return next.Handle(ctx)
	}
	return f(ctx, next)
}

func chainHandler(handler Handler, interceptors []Interceptor) Handler {
	next := handler
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		current := next
		next = HandlerFunc(func(ctx *Context) (Result, error) {
			return interceptor.Intercept(ctx, current)
		})
	}
	return next
}
