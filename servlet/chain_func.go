package servlet

import "context"

// ChainFunc 将函数适配为 Chain。
type ChainFunc func(ctx context.Context, req *Request, res Response) error

// Next 执行后续处理链函数。
func (f ChainFunc) Next(ctx context.Context, req *Request, res Response) error {
	if f == nil {
		return nil
	}
	return f(ctx, req, res)
}
