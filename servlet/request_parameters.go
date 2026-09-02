package servlet

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// ErrFormBodyTooLarge 表示 URL 编码表单体超过标准解析上限。
var ErrFormBodyTooLarge = errors.New("arkarta/servlet: form body too large")

// ParseParameters 解析并缓存请求参数。
func (r *Request) ParseParameters() error {
	r.parametersOnce.Do(func() {
		r.parameters, r.parametersErr = r.readParameters()
	})
	return r.parametersErr
}

// Parameters 返回查询串和表单体合并后的请求参数副本。
func (r *Request) Parameters() (url.Values, error) {
	if err := r.ParseParameters(); err != nil {
		return nil, err
	}
	return cloneURLValues(r.parameters), nil
}

// Parameter 返回指定请求参数的第一个值。
func (r *Request) Parameter(name string) (string, bool, error) {
	values, err := r.Parameters()
	if err != nil {
		return "", false, err
	}
	list, ok := values[name]
	if !ok || len(list) == 0 {
		return "", false, nil
	}
	return list[0], true, nil
}

// ParameterValues 返回指定请求参数的全部值副本。
func (r *Request) ParameterValues(name string) ([]string, bool, error) {
	values, err := r.Parameters()
	if err != nil {
		return nil, false, err
	}
	list, ok := values[name]
	if !ok {
		return nil, false, nil
	}
	return append([]string(nil), list...), true, nil
}

// ParameterNames 返回请求参数名称的稳定排序副本。
func (r *Request) ParameterNames() ([]string, error) {
	values, err := r.Parameters()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func (r *Request) readParameters() (url.Values, error) {
	values, err := url.ParseQuery(r.QueryString())
	if err != nil {
		return nil, err
	}
	if shouldParseFormParameters(r) {
		body, err := io.ReadAll(io.LimitReader(r.Body(), r.maxFormBodySize+1))
		if err != nil {
			return nil, err
		}
		if int64(len(body)) > r.maxFormBodySize {
			return nil, ErrFormBodyTooLarge
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		for name, list := range form {
			values[name] = append(values[name], list...)
		}
	}
	return values, nil
}

func shouldParseFormParameters(request *Request) bool {
	if request == nil || request.Body() == nil {
		return false
	}
	switch request.Method() {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
	default:
		return false
	}
	mediaType, _, err := mime.ParseMediaType(request.Header().Get("Content-Type"))
	if err != nil {
		return false
	}
	return strings.EqualFold(mediaType, "application/x-www-form-urlencoded")
}

func cloneURLValues(src url.Values) url.Values {
	if len(src) == 0 {
		return url.Values{}
	}
	dst := make(url.Values, len(src))
	for key, values := range src {
		dst[key] = append([]string(nil), values...)
	}
	return dst
}
