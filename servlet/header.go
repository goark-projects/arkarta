package servlet

import (
	"net/textproto"
	"sort"
)

// Header 表示与具体 HTTP 实现无关的消息头集合。
// Values 和 Visit 返回的值只在当前请求生命周期内有效，调用方不得修改或长期持有。
type Header interface {
	Get(name string) string
	Values(name string) []string
	Has(name string) bool
	Set(name, value string)
	Add(name, value string)
	Delete(name string)
	Visit(visitor func(name, value string) bool)
}

type mapHeader map[string][]string

// NewHeader 创建标准内存 Header 实现。
func NewHeader() Header {
	return make(mapHeader)
}

// CloneHeader 创建与源 Header 所有权独立的副本。
func CloneHeader(source Header) Header {
	cloned := make(mapHeader)
	if source == nil {
		return cloned
	}
	source.Visit(func(name, value string) bool {
		cloned.Add(name, value)
		return true
	})
	return cloned
}

// HeaderNames 返回去重并稳定排序后的规范 Header 名称。
func HeaderNames(header Header) []string {
	if header == nil {
		return nil
	}
	names := make(map[string]struct{})
	header.Visit(func(name, _ string) bool {
		names[canonicalHeaderName(name)] = struct{}{}
		return true
	})
	result := make([]string, 0, len(names))
	for name := range names {
		if name != "" {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}

func (h mapHeader) Get(name string) string {
	values := h[canonicalHeaderName(name)]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (h mapHeader) Values(name string) []string {
	return h[canonicalHeaderName(name)]
}

func (h mapHeader) Has(name string) bool {
	_, ok := h[canonicalHeaderName(name)]
	return ok
}

func (h mapHeader) Set(name, value string) {
	name = canonicalHeaderName(name)
	if name == "" {
		return
	}
	h[name] = []string{value}
}

func (h mapHeader) Add(name, value string) {
	name = canonicalHeaderName(name)
	if name == "" {
		return
	}
	h[name] = append(h[name], value)
}

func (h mapHeader) Delete(name string) {
	delete(h, canonicalHeaderName(name))
}

func (h mapHeader) Visit(visitor func(name, value string) bool) {
	if visitor == nil {
		return
	}
	for name, values := range h {
		for _, value := range values {
			if !visitor(name, value) {
				return
			}
		}
	}
}

func canonicalHeaderName(name string) string {
	return textproto.CanonicalMIMEHeaderKey(name)
}
