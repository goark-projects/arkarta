package servlet

import (
	"mime"
	"net/http"
	"net/textproto"
	"sort"
	"strconv"
	"time"
)

// Cookies 返回请求 Cookie 的防御性副本。
func (r *Request) Cookies() []*Cookie {
	r.cookiesOnce.Do(func() {
		r.cookies = parseRequestCookies(r.header)
	})
	cookies := r.cookies
	if len(cookies) == 0 {
		return nil
	}
	result := make([]*Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		cloned := *cookie
		result = append(result, &cloned)
	}
	return result
}

// Cookie 返回指定名称的 Cookie。
func (r *Request) Cookie(name string) (*Cookie, error) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == name {
			return cookie, nil
		}
	}
	return nil, ErrNoCookie
}

// ContentType 返回请求 Content-Type 的媒体类型部分。
func (r *Request) ContentType() string {
	mediaType, _, err := mime.ParseMediaType(r.Header().Get("Content-Type"))
	if err != nil {
		return ""
	}
	return mediaType
}

// CharacterEncoding 返回请求 Content-Type 中的 charset 参数。
func (r *Request) CharacterEncoding() string {
	_, params, err := mime.ParseMediaType(r.Header().Get("Content-Type"))
	if err != nil {
		return ""
	}
	return params["charset"]
}

// HeaderNames 返回请求头名称的稳定排序副本。
func (r *Request) HeaderNames() []string {
	return HeaderNames(r.header)
}

// HeaderValue 返回指定请求头的第一个值。
func (r *Request) HeaderValue(name string) (string, bool) {
	values := r.header.Values(name)
	if len(values) == 0 {
		return "", false
	}
	return values[0], true
}

// Headers 返回指定请求头的全部值副本。
func (r *Request) Headers(name string) []string {
	return append([]string(nil), r.header.Values(name)...)
}

// DateHeader 按 HTTP 日期格式解析请求头。
func (r *Request) DateHeader(name string) (time.Time, bool, error) {
	value, ok := r.HeaderValue(name)
	if !ok {
		return time.Time{}, false, nil
	}
	parsed, err := http.ParseTime(value)
	if err != nil {
		return time.Time{}, true, err
	}
	return parsed, true, nil
}

// IntHeader 按十进制整数解析请求头。
func (r *Request) IntHeader(name string) (int, bool, error) {
	value, ok := r.HeaderValue(name)
	if !ok {
		return 0, false, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, true, err
	}
	return parsed, true, nil
}

// Trailer 返回请求 Trailer 字段副本。
func (r *Request) Trailer() Header {
	return CloneHeader(r.trailer)
}

// TrailerFieldsReady 表示请求 Trailer 是否已由容器读取完成。
func (r *Request) TrailerFieldsReady() bool {
	return r.trailerReady == nil || r.trailerReady()
}

func requestPath(httpRequest *http.Request) string {
	if httpRequest.URL == nil {
		return ""
	}
	return httpRequest.URL.Path
}

func requestURI(httpRequest *http.Request) string {
	if httpRequest.URL == nil {
		return ""
	}
	if uri := httpRequest.URL.EscapedPath(); uri != "" {
		return uri
	}
	if httpRequest.URL.Path != "" {
		return httpRequest.URL.Path
	}
	return "/"
}

func requestQueryString(httpRequest *http.Request) string {
	if httpRequest.URL == nil {
		return ""
	}
	return httpRequest.URL.RawQuery
}

func requestInputFromHTTP(httpRequest *http.Request) RequestInput {
	scheme := "http"
	if httpRequest.URL != nil && httpRequest.URL.Scheme != "" {
		scheme = httpRequest.URL.Scheme
	} else if httpRequest.TLS != nil {
		scheme = "https"
	}
	localAddr := ""
	if addr, ok := httpRequest.Context().Value(http.LocalAddrContextKey).(interface{ String() string }); ok && addr != nil {
		localAddr = addr.String()
	}
	return RequestInput{
		Context:       httpRequest.Context(),
		Method:        httpRequest.Method,
		Protocol:      httpRequest.Proto,
		Scheme:        scheme,
		Host:          httpRequest.Host,
		RequestURI:    requestURI(httpRequest),
		Path:          requestPath(httpRequest),
		QueryString:   requestQueryString(httpRequest),
		Header:        mapHeader(httpRequest.Header),
		Body:          httpRequest.Body,
		ContentLength: httpRequest.ContentLength,
		RemoteAddr:    httpRequest.RemoteAddr,
		LocalAddr:     localAddr,
		Trailer:       mapHeader(httpRequest.Trailer),
		TrailerReady: func() bool {
			for _, values := range httpRequest.Trailer {
				if values == nil {
					return false
				}
			}
			return true
		},
	}
}

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
