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
	if path == "" {
		path = "/"
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if handler, ok := r.exact[path]; ok {
		return handler, true
	}
	for _, route := range r.prefix {
		if matchPrefix(path, route.prefix) {
			return route.handler, true
		}
	}
	if ext := extensionOf(path); ext != "" {
		if handler, ok := r.extension[ext]; ok {
			return handler, true
		}
	}
	if r.defaultH != nil {
		return r.defaultH, true
	}
	return nil, false
}

// Serve 查找匹配处理器并执行。
func (r *Router) Serve(ctx context.Context, req *Request, res Response) error {
	handler, ok := r.Match(req.Path())
	if !ok {
		return NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound), nil)
	}
	return handler.Serve(ctx, req, res)
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

func extensionOf(path string) string {
	lastSlash := strings.LastIndexByte(path, '/')
	lastDot := strings.LastIndexByte(path, '.')
	if lastDot <= lastSlash {
		return ""
	}
	return path[lastDot:]
}
