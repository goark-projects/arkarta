package servlet

import (
	"mime"
	"sort"
	"strconv"
	"strings"
)

// MediaType 表示 HTTP 媒体类型或 Accept 头中的媒体范围。
type MediaType struct {
	value   string
	typ     string
	subtype string
	params  map[string]string
	quality float64
	index   int
}

// NewMediaType 解析单个媒体类型。
func NewMediaType(value string) (MediaType, bool) {
	return parseMediaType(value, 0)
}

// Type 返回主类型。
func (m MediaType) Type() string {
	return m.typ
}

// Subtype 返回子类型。
func (m MediaType) Subtype() string {
	return m.subtype
}

// Quality 返回 Accept 头声明的质量因子。
func (m MediaType) Quality() float64 {
	return m.quality
}

// Parameter 返回媒体类型参数。
func (m MediaType) Parameter(name string) (string, bool) {
	value, ok := m.params[strings.ToLower(strings.TrimSpace(name))]
	return value, ok
}

// Parameters 返回媒体类型参数副本。
func (m MediaType) Parameters() map[string]string {
	return cloneMediaParams(m.params)
}

// String 返回规范化媒体类型字符串。
func (m MediaType) String() string {
	if m.value == "" {
		return ""
	}
	return m.value
}

// Matches 判断当前媒体范围是否接受目标媒体类型。
func (m MediaType) Matches(candidate string) bool {
	other, ok := NewMediaType(candidate)
	if !ok {
		return false
	}
	if m.typ != "*" && m.typ != other.typ {
		return false
	}
	if m.subtype == "*" || m.subtype == other.subtype {
		return true
	}
	if strings.HasPrefix(m.subtype, "*+") {
		return strings.HasSuffix(other.subtype, strings.TrimPrefix(m.subtype, "*"))
	}
	return false
}

// Specificity 返回媒体范围匹配的精确程度。
func (m MediaType) Specificity() int {
	switch {
	case m.typ == "*" && m.subtype == "*":
		return 0
	case m.subtype == "*" || strings.HasPrefix(m.subtype, "*+"):
		return 1
	default:
		return 2
	}
}

// ParseAccept 解析一个或多个 Accept 头，返回按客户端偏好排序的媒体范围。
func ParseAccept(values ...string) []MediaType {
	var result []MediaType
	index := 0
	for _, value := range values {
		for _, part := range splitCommaHeader(value) {
			mediaType, ok := parseMediaType(part, index)
			index++
			if !ok {
				continue
			}
			result = append(result, mediaType)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Quality() != right.Quality() {
			return left.Quality() > right.Quality()
		}
		if left.Specificity() != right.Specificity() {
			return left.Specificity() > right.Specificity()
		}
		if len(left.params) != len(right.params) {
			return len(left.params) > len(right.params)
		}
		return left.index < right.index
	})
	return result
}

// AcceptedMediaTypes 返回请求 Accept 头中客户端声明的媒体范围。
func (r *Request) AcceptedMediaTypes() []MediaType {
	if r == nil || r.httpRequest == nil {
		return nil
	}
	return ParseAccept(r.httpRequest.Header.Values("Accept")...)
}

// NegotiateContentType 从候选响应类型中选择最符合 Accept 头的类型。
func (r *Request) NegotiateContentType(candidates ...string) (string, bool) {
	return NegotiateContentType(r.AcceptedMediaTypes(), candidates...)
}

// NegotiateContentType 根据已解析的 Accept 媒体范围选择响应类型。
func NegotiateContentType(accepted []MediaType, candidates ...string) (string, bool) {
	if len(candidates) == 0 {
		return "", false
	}
	parsedCandidates := make([]MediaType, 0, len(candidates))
	candidateTexts := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		mediaType, ok := NewMediaType(candidate)
		if !ok {
			continue
		}
		parsedCandidates = append(parsedCandidates, mediaType)
		candidateTexts = append(candidateTexts, candidate)
	}
	if len(parsedCandidates) == 0 {
		return "", false
	}
	if len(accepted) == 0 {
		return candidateTexts[0], true
	}

	bestIndex := -1
	bestQuality := -1.0
	bestSpecificity := -1
	for i, candidate := range parsedCandidates {
		quality, specificity, ok := mediaQuality(candidate, accepted)
		if !ok || quality <= 0 {
			continue
		}
		if quality > bestQuality || quality == bestQuality && specificity > bestSpecificity {
			bestIndex = i
			bestQuality = quality
			bestSpecificity = specificity
		}
	}
	if bestIndex < 0 {
		return "", false
	}
	return candidateTexts[bestIndex], true
}

func mediaQuality(candidate MediaType, accepted []MediaType) (float64, int, bool) {
	bestQuality := 0.0
	bestSpecificity := -1
	found := false
	for _, mediaRange := range accepted {
		if !mediaRange.Matches(candidate.String()) {
			continue
		}
		specificity := mediaRange.Specificity()
		if !found || specificity > bestSpecificity || specificity == bestSpecificity && mediaRange.Quality() > bestQuality {
			found = true
			bestQuality = mediaRange.Quality()
			bestSpecificity = specificity
		}
	}
	return bestQuality, bestSpecificity, found
}

func parseMediaType(value string, index int) (MediaType, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return MediaType{}, false
	}
	rawType, params, err := mime.ParseMediaType(value)
	if err != nil {
		return MediaType{}, false
	}
	typ, subtype, ok := strings.Cut(strings.ToLower(rawType), "/")
	if !ok || typ == "" || subtype == "" {
		return MediaType{}, false
	}
	quality := 1.0
	normalizedParams := make(map[string]string, len(params))
	for name, param := range params {
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "q" {
			if parsed, err := strconv.ParseFloat(param, 64); err == nil && parsed >= 0 && parsed <= 1 {
				quality = parsed
			}
			continue
		}
		if name != "" {
			normalizedParams[name] = param
		}
	}
	return MediaType{
		value:   mime.FormatMediaType(typ+"/"+subtype, normalizedParams),
		typ:     typ,
		subtype: subtype,
		params:  normalizedParams,
		quality: quality,
		index:   index,
	}, true
}

func splitCommaHeader(value string) []string {
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
		case r == ',' && !inQuote:
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

func cloneMediaParams(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
