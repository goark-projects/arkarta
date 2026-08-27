package web

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"sync"

	arkjson "goark.dev/arkarta/json"
	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/validation"
)

// Router 是 Arkarta Web 的方法路由器。
type Router struct {
	mu           sync.RWMutex
	routes       []route
	interceptors []Interceptor
	advice       []ResponseAdvice
	codec        arkjson.Codec
	validator    validation.Validator
	errorMapper  ErrorMapper
}

type route struct {
	method       string
	pattern      routePattern
	handler      Handler
	order        int
	interceptors []Interceptor
}

// NewRouter 创建 Web 路由器。
func NewRouter(options ...Option) *Router {
	router := &Router{
		codec:       arkjson.DefaultCodec(),
		validator:   validation.NewValidator(),
		errorMapper: DefaultErrorMapper{},
	}
	for _, option := range options {
		if option != nil {
			option(router)
		}
	}
	return router
}

// Handle 注册 HTTP 方法和路径模式。
func (r *Router) Handle(method, pattern string, handler Handler) error {
	return r.handle(method, pattern, handler, nil)
}

func (r *Router) handle(method, pattern string, handler Handler, interceptors []Interceptor) error {
	if r == nil {
		return ErrNilContext
	}
	if handler == nil {
		return ErrNilHandler
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		return ErrInvalidRoutePattern
	}
	parsed, err := parseRoutePattern(pattern)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.routes {
		if existing.method == method && existing.pattern.raw == parsed.raw {
			return ErrDuplicateRoute
		}
	}
	r.routes = append(r.routes, route{
		method:       method,
		pattern:      parsed,
		handler:      handler,
		order:        len(r.routes),
		interceptors: append([]Interceptor(nil), interceptors...),
	})
	return nil
}

// Use 注册全局拦截器。
func (r *Router) Use(interceptor Interceptor) {
	if r == nil || interceptor == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.interceptors = append(r.interceptors, interceptor)
}

// UseResponseAdvice 注册全局响应增强器。
func (r *Router) UseResponseAdvice(advice ResponseAdvice) {
	if r == nil || advice == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.advice = append(r.advice, advice)
}

// Serve 执行 Web 路由匹配、拦截器链和结果写出。
func (r *Router) Serve(ctx context.Context, req *servlet.Request, res servlet.Response) error {
	if r == nil {
		return ErrNilContext
	}
	webCtx := newContext(ctx, req, res, nil, r.codec, r.validator)
	if req == nil || res == nil {
		return r.writeError(webCtx, servlet.NewHTTPError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), ErrNilContext))
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method()))
	match, allowed, ok := r.lookup(method, req.Path())
	if !ok {
		if len(allowed) > 0 {
			res.Header().Set("Allow", strings.Join(allowed, ", "))
			if method == http.MethodOptions {
				return NoContent().Write(webCtx)
			}
			return r.writeError(webCtx, servlet.NewHTTPError(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), nil))
		}
		return r.writeError(webCtx, servlet.NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound), nil))
	}

	if method == http.MethodHead {
		res = noBodyResponse{Response: res}
	}
	webCtx = newContext(ctx, req, res, match.pathValues, r.codec, r.validator)
	handler := chainHandler(match.handler, match.interceptors)
	result, err := handler.Handle(webCtx)
	if err != nil {
		return r.writeError(webCtx, err)
	}
	if result == nil {
		return nil
	}
	result, err = applyResponseAdvice(webCtx, result, match.advice)
	if err != nil {
		return r.writeError(webCtx, err)
	}
	if result == nil {
		return nil
	}
	if err := result.Write(webCtx); err != nil {
		return r.writeError(webCtx, err)
	}
	return nil
}

func (r *Router) lookup(method, requestPath string) (routeMatch, []string, bool) {
	method = strings.ToUpper(strings.TrimSpace(method))
	requestPath = normalizeRoutePath(requestPath)

	r.mu.RLock()
	defer r.mu.RUnlock()

	var best routeMatch
	var found bool
	allowed := make(map[string]struct{})
	var fallback routeMatch
	var fallbackFound bool
	for _, candidate := range r.routes {
		pathValues, pathMatched := candidate.pattern.match(requestPath)
		if !pathMatched {
			continue
		}
		recordAllowedMethod(allowed, candidate.method)
		if candidate.method != method {
			if method == http.MethodHead && candidate.method == http.MethodGet {
				if !fallbackFound || betterRoute(candidate, fallback.route) {
					fallbackFound = true
					fallback = newRouteMatch(candidate, pathValues, r.interceptors, r.advice)
				}
			}
			continue
		}
		if !found || betterRoute(candidate, best.route) {
			found = true
			best = newRouteMatch(candidate, pathValues, r.interceptors, r.advice)
		}
	}
	if found {
		return best, sortedMethods(allowed), true
	}
	if method == http.MethodHead && fallbackFound {
		return fallback, sortedMethods(allowed), true
	}
	return best, sortedMethods(allowed), found
}

func (r *Router) writeError(ctx *Context, err error) error {
	if ctx == nil || ctx.response == nil {
		return err
	}
	if ctx.response.Committed() {
		return err
	}
	mapper := r.errorMapper
	if mapper == nil {
		mapper = DefaultErrorMapper{}
	}
	result := mapper.MapError(ctx, err)
	if result == nil {
		return err
	}
	return result.Write(ctx)
}

type routeMatch struct {
	route        route
	handler      Handler
	pathValues   map[string]string
	interceptors []Interceptor
	advice       []ResponseAdvice
}

func newRouteMatch(candidate route, pathValues map[string]string, globalInterceptors []Interceptor, advice []ResponseAdvice) routeMatch {
	interceptors := make([]Interceptor, 0, len(globalInterceptors)+len(candidate.interceptors))
	interceptors = append(interceptors, globalInterceptors...)
	interceptors = append(interceptors, candidate.interceptors...)
	return routeMatch{
		route:        candidate,
		handler:      candidate.handler,
		pathValues:   pathValues,
		interceptors: interceptors,
		advice:       append([]ResponseAdvice(nil), advice...),
	}
}

func betterRoute(candidate, current route) bool {
	if candidate.pattern.score != current.pattern.score {
		return candidate.pattern.score > current.pattern.score
	}
	if len(candidate.pattern.segments) != len(current.pattern.segments) {
		return len(candidate.pattern.segments) > len(current.pattern.segments)
	}
	return candidate.order < current.order
}

func sortedMethods(methods map[string]struct{}) []string {
	if len(methods) == 0 {
		return nil
	}
	result := make([]string, 0, len(methods))
	for method := range methods {
		result = append(result, method)
	}
	sort.Strings(result)
	return result
}

func recordAllowedMethod(methods map[string]struct{}, method string) {
	methods[method] = struct{}{}
	if method == http.MethodGet {
		methods[http.MethodHead] = struct{}{}
	}
	methods[http.MethodOptions] = struct{}{}
}
