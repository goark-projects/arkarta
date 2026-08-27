package websocket

import (
	"sort"
	"strconv"
	"strings"
)

// Extension 表示 WebSocket 扩展及其参数。
type Extension struct {
	name   string
	params map[string]string
}

// NewExtension 创建扩展声明。
func NewExtension(name string, params map[string]string) (Extension, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if !validToken(name) {
		return Extension{}, false
	}
	return Extension{name: name, params: normalizeParameters(params)}, true
}

// Name 返回扩展名称。
func (e Extension) Name() string {
	return e.name
}

// Parameter 返回扩展参数。
func (e Extension) Parameter(name string) (string, bool) {
	value, ok := e.params[strings.ToLower(strings.TrimSpace(name))]
	return value, ok
}

// Parameters 返回参数副本。
func (e Extension) Parameters() map[string]string {
	return cloneStringMap(e.params)
}

func (e Extension) String() string {
	var builder strings.Builder
	builder.WriteString(e.name)
	keys := make([]string, 0, len(e.params))
	for name := range e.params {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		builder.WriteString("; ")
		builder.WriteString(name)
		value := e.params[name]
		if value == "" {
			continue
		}
		builder.WriteByte('=')
		if validToken(value) {
			builder.WriteString(value)
			continue
		}
		builder.WriteString(strconv.Quote(value))
	}
	return builder.String()
}

// ParseExtensions 解析 Sec-WebSocket-Extensions 头。
func ParseExtensions(values ...string) []Extension {
	result := make([]Extension, 0)
	for _, value := range values {
		for _, item := range splitHeaderValues(value, ',') {
			if extension, ok := parseExtension(item); ok {
				result = append(result, extension)
			}
		}
	}
	return result
}

// FormatExtensions 格式化 Sec-WebSocket-Extensions 响应头。
func FormatExtensions(extensions []Extension) string {
	parts := make([]string, 0, len(extensions))
	for _, extension := range extensions {
		if extension.name != "" {
			parts = append(parts, extension.String())
		}
	}
	return strings.Join(parts, ", ")
}

func parseExtension(value string) (Extension, bool) {
	parts := splitHeaderValues(value, ';')
	if len(parts) == 0 {
		return Extension{}, false
	}
	name := strings.ToLower(strings.TrimSpace(parts[0]))
	if !validToken(name) {
		return Extension{}, false
	}
	params := make(map[string]string)
	for _, part := range parts[1:] {
		key, rawValue, hasValue := strings.Cut(strings.TrimSpace(part), "=")
		key = strings.ToLower(strings.TrimSpace(key))
		if !validToken(key) {
			continue
		}
		if !hasValue {
			params[key] = ""
			continue
		}
		rawValue = strings.TrimSpace(rawValue)
		if unquoted, err := strconv.Unquote(rawValue); err == nil {
			rawValue = unquoted
		}
		params[key] = rawValue
	}
	return Extension{name: name, params: params}, true
}

// ExtensionNegotiator 定义扩展协商策略。
type ExtensionNegotiator interface {
	Negotiate(offers []Extension) (Extension, bool)
}

// ExtensionNegotiatorFunc 将函数适配为扩展协商器。
type ExtensionNegotiatorFunc func(offers []Extension) (Extension, bool)

// Negotiate 执行扩展协商。
func (f ExtensionNegotiatorFunc) Negotiate(offers []Extension) (Extension, bool) {
	return f(offers)
}

func appendTokens(dst []string, values ...string) []string {
	seen := make(map[string]struct{}, len(dst)+len(values))
	result := make([]string, 0, len(dst)+len(values))
	for _, value := range append(append([]string(nil), dst...), values...) {
		for _, token := range headerTokens([]string{value}) {
			normalized := strings.ToLower(token)
			if _, ok := seen[normalized]; ok || !validToken(token) {
				continue
			}
			seen[normalized] = struct{}{}
			result = append(result, token)
		}
	}
	return result
}

func headerTokens(values []string) []string {
	result := make([]string, 0)
	for _, value := range values {
		for _, part := range splitHeaderValues(value, ',') {
			token := strings.TrimSpace(part)
			if validToken(token) {
				result = append(result, token)
			}
		}
	}
	return result
}

func splitHeaderValues(value string, separator rune) []string {
	var result []string
	var builder strings.Builder
	inQuote := false
	escaped := false
	for _, r := range value {
		switch {
		case escaped:
			builder.WriteRune(r)
			escaped = false
		case r == '\\' && inQuote:
			builder.WriteRune(r)
			escaped = true
		case r == '"':
			builder.WriteRune(r)
			inQuote = !inQuote
		case r == separator && !inQuote:
			if item := strings.TrimSpace(builder.String()); item != "" {
				result = append(result, item)
			}
			builder.Reset()
		default:
			builder.WriteRune(r)
		}
	}
	if item := strings.TrimSpace(builder.String()); item != "" {
		result = append(result, item)
	}
	return result
}

func validToken(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r <= 0x20 || r >= 0x7f {
			return false
		}
		switch r {
		case '(', ')', '<', '>', '@', ',', ';', ':', '\\', '"', '/', '[', ']', '?', '=', '{', '}':
			return false
		}
	}
	return true
}

func cloneExtensions(src []Extension) []Extension {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Extension, len(src))
	for i, extension := range src {
		dst[i] = Extension{name: extension.name, params: cloneStringMap(extension.params)}
	}
	return dst
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func normalizeParameters(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		key = strings.ToLower(strings.TrimSpace(key))
		if validToken(key) {
			dst[key] = value
		}
	}
	return dst
}
