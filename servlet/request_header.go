package servlet

import (
	"mime"
	"net/http"
	"sort"
	"strconv"
	"time"
)

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
	names := make([]string, 0, len(r.httpRequest.Header))
	for name := range r.httpRequest.Header {
		names = append(names, http.CanonicalHeaderKey(name))
	}
	sort.Strings(names)
	return names
}

// HeaderValue 返回指定请求头的第一个值。
func (r *Request) HeaderValue(name string) (string, bool) {
	values := r.httpRequest.Header.Values(name)
	if len(values) == 0 {
		return "", false
	}
	return values[0], true
}

// Headers 返回指定请求头的全部值副本。
func (r *Request) Headers(name string) []string {
	return append([]string(nil), r.httpRequest.Header.Values(name)...)
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
func (r *Request) Trailer() http.Header {
	if r.httpRequest.Trailer == nil {
		return http.Header{}
	}
	return r.httpRequest.Trailer.Clone()
}

// TrailerFieldsReady 表示请求 Trailer 是否已由容器读取完成。
func (r *Request) TrailerFieldsReady() bool {
	if r.httpRequest.Trailer == nil {
		return true
	}
	for _, values := range r.httpRequest.Trailer {
		if values == nil {
			return false
		}
	}
	return true
}
