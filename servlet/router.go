package servlet

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// ErrInvalidMappingPattern 表示路径映射模式非法。
var ErrInvalidMappingPattern = errors.New("arkarta/servlet: invalid mapping pattern")

// ErrDuplicateMapping 表示路径映射模式重复注册。
var ErrDuplicateMapping = errors.New("arkarta/servlet: duplicate mapping pattern")

// Router 实现 Servlet 标准路径映射。
type Router struct {
	mu sync.RWMutex

	exact     map[string]Handler
	prefix    []prefixRoute
	extension map[string]Handler
	defaultH  Handler
}

type prefixRoute struct {
	prefix  string
	pattern string
	handler Handler
}

// RouteMatch 表示一次路由匹配结果。
type RouteMatch struct {
	handler Handler
	mapping RequestMapping
}

// Handler 返回命中的处理器。
func (m RouteMatch) Handler() Handler {
	return m.handler
}

// Mapping 返回命中的 Servlet 映射信息。
func (m RouteMatch) Mapping() RequestMapping {
	return m.mapping
}

// NewRouter 创建空路由器。
func NewRouter() *Router {
	return &Router{
		exact:     make(map[string]Handler),
		extension: make(map[string]Handler),
	}
}

// Handle 注册 Servlet 路径映射。
func (r *Router) Handle(pattern string, handler Handler) error {
	if handler == nil {
		return ErrNilHandler
	}
	kind, value, err := parseMappingPattern(pattern)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	switch kind {
	case mappingDefault:
		if r.defaultH != nil {
			return ErrDuplicateMapping
		}
		r.defaultH = handler
	case mappingExact:
		if _, exists := r.exact[value]; exists {
			return ErrDuplicateMapping
		}
		r.exact[value] = handler
	case mappingPrefix:
		for _, route := range r.prefix {
			if route.pattern == pattern {
				return ErrDuplicateMapping
			}
		}
		r.prefix = append(r.prefix, prefixRoute{prefix: value, pattern: pattern, handler: handler})
		sort.SliceStable(r.prefix, func(i, j int) bool {
			return len(r.prefix[i].prefix) > len(r.prefix[j].prefix)
		})
	case mappingExtension:
		if _, exists := r.extension[value]; exists {
			return ErrDuplicateMapping
		}
		r.extension[value] = handler
	default:
		return ErrInvalidMappingPattern
	}
	return nil
}

// Match 按 Servlet 映射优先级查找处理器。
func (r *Router) Match(path string) (Handler, bool) {
	match, ok := r.MatchRoute(path)
	if !ok {
		return nil, false
	}
	return match.Handler(), true
}

// MatchRoute 按 Servlet 映射优先级查找处理器和映射信息。
func (r *Router) MatchRoute(path string) (RouteMatch, bool) {
	if path == "" {
		path = "/"
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if handler, ok := r.exact[path]; ok {
		return RouteMatch{
			handler: handler,
			mapping: newRequestMapping(
				path,
				MappingExact,
				path,
				"",
			),
		}, true
	}
	for _, route := range r.prefix {
		if matchPrefix(path, route.prefix) {
			return RouteMatch{
				handler: route.handler,
				mapping: newRequestMapping(
					route.pattern,
					MappingPrefix,
					route.prefix,
					prefixPathInfo(path, route.prefix),
				),
			}, true
		}
	}
	if ext := extensionOf(path); ext != "" {
		if handler, ok := r.extension[ext]; ok {
			return RouteMatch{
				handler: handler,
				mapping: newRequestMapping(
					"*"+ext,
					MappingExtension,
					path,
					"",
				),
			}, true
		}
	}
	if r.defaultH != nil {
		return RouteMatch{
			handler: r.defaultH,
			mapping: newRequestMapping(
				"/",
				MappingDefault,
				path,
				"",
			),
		}, true
	}
	return RouteMatch{}, false
}

// Serve 查找匹配处理器并执行。
func (r *Router) Serve(ctx context.Context, req *Request, res Response) error {
	match, ok := r.MatchRoute(req.Path())
	if !ok {
		return NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound), nil)
	}
	req.applyMapping(match.Mapping())
	return match.Handler().Serve(ctx, req, res)
}

type mappingKind uint8

const (
	mappingDefault mappingKind = iota
	mappingExact
	mappingPrefix
	mappingExtension
)

func parseMappingPattern(pattern string) (mappingKind, string, error) {
	if pattern == "" {
		return 0, "", ErrInvalidMappingPattern
	}
	if pattern == "/" {
		return mappingDefault, "/", nil
	}
	if strings.HasPrefix(pattern, "*.") && len(pattern) > 2 && !strings.Contains(pattern[2:], "/") {
		return mappingExtension, pattern[1:], nil
	}
	if strings.HasPrefix(pattern, "/") {
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if prefix == "" {
				prefix = "/"
			}
			return mappingPrefix, prefix, nil
		}
		if strings.Contains(pattern, "*") {
			return 0, "", ErrInvalidMappingPattern
		}
		return mappingExact, pattern, nil
	}
	return 0, "", ErrInvalidMappingPattern
}

func matchPrefix(path, prefix string) bool {
	if prefix == "/" {
		return true
	}
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func prefixPathInfo(path, prefix string) string {
	if prefix == "/" {
		if path == "/" {
			return ""
		}
		return path
	}
	if path == prefix {
		return ""
	}
	return strings.TrimPrefix(path, prefix)
}

func extensionOf(path string) string {
	lastSlash := strings.LastIndexByte(path, '/')
	lastDot := strings.LastIndexByte(path, '.')
	if lastDot <= lastSlash {
		return ""
	}
	return path[lastDot:]
}
