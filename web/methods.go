package web

import "net/http"

// GET 注册 GET 路由。
func (r *Router) GET(pattern string, handler Handler) error {
	return r.Handle(http.MethodGet, pattern, handler)
}

// POST 注册 POST 路由。
func (r *Router) POST(pattern string, handler Handler) error {
	return r.Handle(http.MethodPost, pattern, handler)
}

// PUT 注册 PUT 路由。
func (r *Router) PUT(pattern string, handler Handler) error {
	return r.Handle(http.MethodPut, pattern, handler)
}

// PATCH 注册 PATCH 路由。
func (r *Router) PATCH(pattern string, handler Handler) error {
	return r.Handle(http.MethodPatch, pattern, handler)
}

// DELETE 注册 DELETE 路由。
func (r *Router) DELETE(pattern string, handler Handler) error {
	return r.Handle(http.MethodDelete, pattern, handler)
}

// HEAD 注册 HEAD 路由。
func (r *Router) HEAD(pattern string, handler Handler) error {
	return r.Handle(http.MethodHead, pattern, handler)
}

// OPTIONS 注册 OPTIONS 路由。
func (r *Router) OPTIONS(pattern string, handler Handler) error {
	return r.Handle(http.MethodOptions, pattern, handler)
}
