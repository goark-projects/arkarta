package web

import (
	"path"
	"strings"
	"unicode"
)

type routePattern struct {
	raw      string
	segments []routeSegment
	score    int
}

type routeSegment struct {
	literal  string
	name     string
	variable bool
}

func parseRoutePattern(pattern string) (routePattern, error) {
	if pattern == "" || !strings.HasPrefix(pattern, "/") {
		return routePattern{}, ErrInvalidRoutePattern
	}
	normalized := normalizeRoutePath(pattern)
	segments := splitRoutePath(normalized)
	seen := make(map[string]struct{}, len(segments))
	result := routePattern{
		raw:      normalized,
		segments: make([]routeSegment, 0, len(segments)),
	}
	for _, segment := range segments {
		routeSegment, err := parseRouteSegment(segment)
		if err != nil {
			return routePattern{}, err
		}
		if routeSegment.variable {
			if _, exists := seen[routeSegment.name]; exists {
				return routePattern{}, ErrInvalidRoutePattern
			}
			seen[routeSegment.name] = struct{}{}
		} else {
			result.score++
		}
		result.segments = append(result.segments, routeSegment)
	}
	return result, nil
}

func parseRouteSegment(segment string) (routeSegment, error) {
	if segment == "" {
		return routeSegment{}, ErrInvalidRoutePattern
	}
	if strings.HasPrefix(segment, "{") || strings.HasSuffix(segment, "}") {
		if !strings.HasPrefix(segment, "{") || !strings.HasSuffix(segment, "}") {
			return routeSegment{}, ErrInvalidRoutePattern
		}
		name := strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")
		if !validPathValueName(name) {
			return routeSegment{}, ErrInvalidRoutePattern
		}
		return routeSegment{name: name, variable: true}, nil
	}
	if strings.ContainsAny(segment, "{}") {
		return routeSegment{}, ErrInvalidRoutePattern
	}
	return routeSegment{literal: segment}, nil
}

func (p routePattern) match(requestPath string) (map[string]string, bool) {
	segments := splitRoutePath(normalizeRoutePath(requestPath))
	if len(segments) != len(p.segments) {
		return nil, false
	}
	values := make(map[string]string)
	for i, routeSegment := range p.segments {
		value := segments[i]
		if routeSegment.variable {
			values[routeSegment.name] = value
			continue
		}
		if routeSegment.literal != value {
			return nil, false
		}
	}
	return values, true
}

func normalizeRoutePath(value string) string {
	if value == "" {
		return "/"
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	clean := path.Clean(value)
	if clean == "." {
		return "/"
	}
	return clean
}

func splitRoutePath(value string) []string {
	value = normalizeRoutePath(value)
	if value == "/" {
		return nil
	}
	return strings.Split(strings.Trim(value, "/"), "/")
}

func validPathValueName(name string) bool {
	if name == "" {
		return false
	}
	for index, r := range name {
		if index == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
