package web

import (
	"net/http"
	"path"
	"strings"
)

// Group 表示共享路径前缀和局部拦截器的路由分组。
type Group struct {
	router       *Router
	prefix       string
	interceptors []Interceptor
}

// Group 创建共享路径前缀的路由分组。
func (r *Router) Group(prefix string) *Group {
	return &Group{
		router: r,
		prefix: normalizeGroupPrefix(prefix),
	}
}

// Group 创建子路由分组，并继承当前分组的局部拦截器。
func (g *Group) Group(prefix string) *Group {
	if g == nil {
		return &Group{prefix: normalizeGroupPrefix(prefix)}
	}
	return &Group{
		router:       g.router,
		prefix:       joinRoutePath(g.prefix, prefix),
		interceptors: append([]Interceptor(nil), g.interceptors...),
	}
}

// Use 注册当前分组内生效的拦截器。
func (g *Group) Use(interceptor Interceptor) {
	if g == nil || interceptor == nil {
		return
	}
	g.interceptors = append(g.interceptors, interceptor)
}

// Handle 在当前分组下注册 HTTP 方法和路径模式。
func (g *Group) Handle(method, pattern string, handler Handler) error {
	if g == nil || g.router == nil {
		return ErrNilContext
	}
	return g.router.handle(method, joinRoutePath(g.prefix, pattern), handler, g.interceptors)
}

// GET 注册 GET 路由。
func (g *Group) GET(pattern string, handler Handler) error {
	return g.Handle(http.MethodGet, pattern, handler)
}

// POST 注册 POST 路由。
func (g *Group) POST(pattern string, handler Handler) error {
	return g.Handle(http.MethodPost, pattern, handler)
}

// PUT 注册 PUT 路由。
func (g *Group) PUT(pattern string, handler Handler) error {
	return g.Handle(http.MethodPut, pattern, handler)
}

// PATCH 注册 PATCH 路由。
func (g *Group) PATCH(pattern string, handler Handler) error {
	return g.Handle(http.MethodPatch, pattern, handler)
}

// DELETE 注册 DELETE 路由。
func (g *Group) DELETE(pattern string, handler Handler) error {
	return g.Handle(http.MethodDelete, pattern, handler)
}

// HEAD 注册 HEAD 路由。
func (g *Group) HEAD(pattern string, handler Handler) error {
	return g.Handle(http.MethodHead, pattern, handler)
}

// OPTIONS 注册 OPTIONS 路由。
func (g *Group) OPTIONS(pattern string, handler Handler) error {
	return g.Handle(http.MethodOptions, pattern, handler)
}

func normalizeGroupPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return ""
	}
	return normalizeRoutePath(prefix)
}

func joinRoutePath(prefix, pattern string) string {
	pattern = strings.TrimSpace(pattern)
	if prefix == "" {
		return normalizeRoutePath(pattern)
	}
	if pattern == "" || pattern == "/" {
		return prefix
	}
	return normalizeRoutePath(path.Join(prefix, pattern))
}
