package servlet

import (
	"context"
	"errors"
	"net/http"
)

// ErrNilRouter 表示 RequestDispatcher 缺少路由器。
var ErrNilRouter = errors.New("arkarta/servlet: router is nil")

// ErrDispatcherTargetNotFound 表示分发目标不存在。
var ErrDispatcherTargetNotFound = errors.New("arkarta/servlet: dispatcher target not found")

// RequestDispatcher 执行服务端请求分发。
type RequestDispatcher interface {
	Forward(ctx context.Context, req *Request, res Response) error
	Include(ctx context.Context, req *Request, res Response) error
	Error(ctx context.Context, req *Request, res Response, statusCode int, cause error) error
}

// Dispatcher 是基于 Router 的 RequestDispatcher 实现。
type Dispatcher struct {
	router *Router
	path   string
}

// NewRequestDispatcher 创建指定目标路径的请求分发器。
func NewRequestDispatcher(router *Router, path string) (*Dispatcher, error) {
	if router == nil {
		return nil, ErrNilRouter
	}
	if path == "" || path[0] != '/' {
		return nil, ErrInvalidMappingPattern
	}
	return &Dispatcher{router: router, path: path}, nil
}

// Forward 在响应提交前转发请求。
func (d *Dispatcher) Forward(ctx context.Context, req *Request, res Response) error {
	if res != nil && res.Committed() {
		return ErrResponseCommitted
	}
	if res != nil {
		if err := res.Reset(); err != nil {
			return err
		}
	}
	req.SetAttribute(AttributeForwardRequestURI, req.Path())
	return d.dispatch(ctx, req, res, DispatchForward)
}

// Include 包含目标处理器输出，但隔离目标状态码和 Header。
func (d *Dispatcher) Include(ctx context.Context, req *Request, res Response) error {
	req.SetAttribute(AttributeIncludeRequestURI, req.Path())
	return d.dispatch(ctx, req, newIncludeResponse(res), DispatchInclude)
}

// Error 执行错误分发。
func (d *Dispatcher) Error(ctx context.Context, req *Request, res Response, statusCode int, cause error) error {
	if statusCode < 100 || statusCode > 999 {
		statusCode = http.StatusInternalServerError
	}
	if res != nil && !res.Committed() {
		res.SetStatus(statusCode)
	}
	req.SetAttribute(AttributeErrorStatusCode, statusCode)
	req.SetAttribute(AttributeErrorException, cause)
	req.SetAttribute(AttributeErrorRequestURI, req.Path())
	return d.dispatch(ctx, req, res, DispatchError)
}

func (d *Dispatcher) dispatch(ctx context.Context, req *Request, res Response, dispatchType DispatchType) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	handler, ok := d.router.Match(d.path)
	if !ok {
		return ErrDispatcherTargetNotFound
	}

	snapshot := req.dispatchSnapshot()
	req.applyDispatch(d.path, dispatchType)
	defer req.restoreDispatch(snapshot)
	return handler.Serve(ctx, req, res)
}
