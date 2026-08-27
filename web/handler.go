package web

// Handler 表示 Web 层处理器。
type Handler interface {
	Handle(ctx *Context) (Result, error)
}

// HandlerFunc 将普通函数适配为 Handler。
type HandlerFunc func(ctx *Context) (Result, error)

// Handle 执行底层处理函数。
func (f HandlerFunc) Handle(ctx *Context) (Result, error) {
	if f == nil {
		return nil, ErrNilHandler
	}
	return f(ctx)
}
