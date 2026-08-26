package servlet

import (
	"context"
	"errors"
	"sync"
)

// ErrChainAlreadyAdvanced 表示同一个过滤器链节点已经向后推进过。
var ErrChainAlreadyAdvanced = errors.New("arkarta/servlet: filter chain already advanced")

// Filter 是请求进入目标处理器前后的横切处理器。
type Filter interface {
	Filter(ctx context.Context, req *Request, res Response, chain Chain) error
}

// FilterFunc 将普通函数适配为 Filter。
type FilterFunc func(ctx context.Context, req *Request, res Response, chain Chain) error

// Filter 执行函数式过滤器。
func (f FilterFunc) Filter(ctx context.Context, req *Request, res Response, chain Chain) error {
	return f(ctx, req, res, chain)
}

// Chain 表示当前过滤器之后的剩余链路。
type Chain interface {
	Next(ctx context.Context, req *Request, res Response) error
}

// ChainFilters 将过滤器和目标处理器组合为一个 Handler。
func ChainFilters(target Handler, filters ...Filter) Handler {
	chainFilters := make([]Filter, 0, len(filters))
	for _, filter := range filters {
		if filter != nil {
			chainFilters = append(chainFilters, filter)
		}
	}
	return HandlerFunc(func(ctx context.Context, req *Request, res Response) error {
		chain := &filterChain{
			target:  target,
			filters: chainFilters,
		}
		return chain.Next(ctx, req, res)
	})
}

type filterChain struct {
	target  Handler
	filters []Filter
	index   int

	mu       sync.Mutex
	advanced bool
}

func (c *filterChain) Next(ctx context.Context, req *Request, res Response) error {
	c.mu.Lock()
	if c.advanced {
		c.mu.Unlock()
		return ErrChainAlreadyAdvanced
	}
	c.advanced = true
	c.mu.Unlock()

	if c.index >= len(c.filters) {
		return c.target.Serve(ctx, req, res)
	}
	next := &filterChain{
		target:  c.target,
		filters: c.filters,
		index:   c.index + 1,
	}
	return c.filters[c.index].Filter(ctx, req, res, next)
}
