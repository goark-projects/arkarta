package servlet

import "context"

// ChainFilterBindings 将带分发约束的 Filter 与目标处理器组合为 Handler。
func ChainFilterBindings(target Handler, bindings ...FilterBinding) Handler {
	return HandlerFunc(func(ctx context.Context, req *Request, res Response) error {
		dispatchType := DispatchRequest
		if req != nil {
			dispatchType = req.DispatchType()
		}
		filters := make([]Filter, 0, len(bindings))
		for _, binding := range bindings {
			if binding.Filter() != nil && binding.Matches(dispatchType) {
				filters = append(filters, binding.Filter())
			}
		}
		return ChainFilters(target, filters...).Serve(ctx, req, res)
	})
}
