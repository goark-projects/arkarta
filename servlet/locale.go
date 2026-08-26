package servlet

import (
	"sort"
	"strconv"
	"strings"
)

// Locale 表示经过标准化的语言标签。
type Locale struct {
	tag      string
	language string
	region   string
}

// NewLocale 从 BCP 47 风格标签构造 Locale。
func NewLocale(tag string) (Locale, bool) {
	tag = strings.TrimSpace(tag)
	if tag == "" || strings.ContainsAny(tag, " \t\r\n") {
		return Locale{}, false
	}
	parts := strings.Split(strings.ReplaceAll(tag, "_", "-"), "-")
	if parts[0] == "" {
		return Locale{}, false
	}
	language := strings.ToLower(parts[0])
	region := ""
	if len(parts) > 1 && parts[1] != "" {
		region = strings.ToUpper(parts[1])
	}
	normalized := language
	if region != "" {
		normalized += "-" + region
	}
	return Locale{tag: normalized, language: language, region: region}, true
}

// String 返回标准化语言标签。
func (l Locale) String() string {
	return l.tag
}

// Tag 返回标准化语言标签。
func (l Locale) Tag() string {
	return l.tag
}

// Language 返回语言子标签。
func (l Locale) Language() string {
	return l.language
}

// Region 返回地区子标签。
func (l Locale) Region() string {
	return l.region
}

// Locale 返回 Accept-Language 中优先级最高的语言。
func (r *Request) Locale() (Locale, bool) {
	locales := r.Locales()
	if len(locales) == 0 {
		return Locale{}, false
	}
	return locales[0], true
}

// Locales 按客户端声明优先级返回语言列表。
func (r *Request) Locales() []Locale {
	return ParseAcceptLanguage(r.Header().Get("Accept-Language"))
}

// ParseAcceptLanguage 解析 HTTP Accept-Language 头。
func ParseAcceptLanguage(header string) []Locale {
	type candidate struct {
		locale Locale
		q      float64
		index  int
	}
	candidates := make([]candidate, 0)
	for index, item := range strings.Split(header, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		value, paramsText, _ := strings.Cut(item, ";")
		locale, ok := NewLocale(value)
		if !ok {
			continue
		}
		q := 1.0
		if paramsText != "" {
			for _, param := range strings.Split(paramsText, ";") {
				name, raw, ok := strings.Cut(strings.TrimSpace(param), "=")
				if !ok || !strings.EqualFold(name, "q") {
					continue
				}
				parsed, err := strconv.ParseFloat(raw, 64)
				if err == nil && parsed >= 0 && parsed <= 1 {
					q = parsed
				}
			}
		}
		if q == 0 {
			continue
		}
		candidates = append(candidates, candidate{locale: locale, q: q, index: index})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].q == candidates[j].q {
			return candidates[i].index < candidates[j].index
		}
		return candidates[i].q > candidates[j].q
	})
	result := make([]Locale, len(candidates))
	for i, item := range candidates {
		result[i] = item.locale
	}
	return result
}
