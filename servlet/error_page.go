package servlet

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

// ErrNilErrorPageRegistry 表示错误页注册表为空。
var ErrNilErrorPageRegistry = errors.New("arkarta/servlet: error page registry is nil")

// ErrorPageRegistry 保存状态码和错误类型到错误页处理器的映射。
type ErrorPageRegistry struct {
	mu         sync.RWMutex
	status     map[int]Handler
	errorTypes []errorTypeMapping
}

type errorTypeMapping struct {
	match   func(error) bool
	handler Handler
}

// NewErrorPageRegistry 创建错误页注册表。
func NewErrorPageRegistry() *ErrorPageRegistry {
	return &ErrorPageRegistry{
		status: make(map[int]Handler),
	}
}

// RegisterStatus 注册 HTTP 状态码错误页。
func (r *ErrorPageRegistry) RegisterStatus(statusCode int, handler Handler) error {
	if r == nil {
		return ErrNilErrorPageRegistry
	}
	if handler == nil {
		return ErrNilHandler
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status[normalizeStatus(statusCode)] = handler
	return nil
}

// Handle 尝试用已注册错误页处理错误。
func (r *ErrorPageRegistry) Handle(ctx context.Context, req *Request, res Response, statusCode int, cause error) (bool, error) {
	if r == nil {
		return false, nil
	}
	if res != nil && res.Committed() {
		return false, ErrResponseCommitted
	}
	handler := r.match(statusCode, cause)
	if handler == nil {
		return false, nil
	}

	statusCode = normalizeStatus(statusCode)
	if res != nil {
		res.SetStatus(statusCode)
	}
	req.SetAttribute(AttributeErrorStatusCode, statusCode)
	req.SetAttribute(AttributeErrorException, cause)
	req.SetAttribute(AttributeErrorRequestURI, req.Path())

	snapshot := req.dispatchSnapshot()
	req.applyDispatch(req.Path(), DispatchError)
	defer req.restoreDispatch(snapshot)
	return true, handler.Serve(ctx, req, res)
}

func (r *ErrorPageRegistry) registerErrorType(match func(error) bool, handler Handler) error {
	if r == nil {
		return ErrNilErrorPageRegistry
	}
	if handler == nil {
		return ErrNilHandler
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errorTypes = append(r.errorTypes, errorTypeMapping{
		match:   match,
		handler: handler,
	})
	return nil
}

func (r *ErrorPageRegistry) match(statusCode int, cause error) Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if cause != nil {
		for _, mapping := range r.errorTypes {
			if mapping.match(cause) {
				return mapping.handler
			}
		}
	}
	return r.status[normalizeStatus(statusCode)]
}

func normalizeStatus(statusCode int) int {
	if statusCode < 100 || statusCode > 999 {
		return http.StatusInternalServerError
	}
	return statusCode
}
