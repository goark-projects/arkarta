package servlet

import (
	"net/http"
	"strconv"
	"time"
)

// SetHeader 设置响应头。
func SetHeader(res Response, name, value string) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	res.Header().Set(name, value)
	return nil
}

// AddHeader 追加响应头。
func AddHeader(res Response, name, value string) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	res.Header().Add(name, value)
	return nil
}

// HeaderValue 返回响应头的第一个值。
func HeaderValue(res Response, name string) (string, bool) {
	if res == nil {
		return "", false
	}
	values := res.Header().Values(name)
	if len(values) == 0 {
		return "", false
	}
	return values[0], true
}

// HeaderValues 返回响应头的全部值副本。
func HeaderValues(res Response, name string) []string {
	if res == nil {
		return nil
	}
	return append([]string(nil), res.Header().Values(name)...)
}

// ContainsHeader 判断响应头是否已经存在。
func ContainsHeader(res Response, name string) bool {
	if res == nil {
		return false
	}
	return res.Header().Has(name)
}

// SetDateHeader 按 HTTP 日期格式设置响应头。
func SetDateHeader(res Response, name string, value time.Time) error {
	return SetHeader(res, name, value.UTC().Format(http.TimeFormat))
}

// AddDateHeader 按 HTTP 日期格式追加响应头。
func AddDateHeader(res Response, name string, value time.Time) error {
	return AddHeader(res, name, value.UTC().Format(http.TimeFormat))
}

// SetIntHeader 按十进制整数设置响应头。
func SetIntHeader(res Response, name string, value int) error {
	return SetHeader(res, name, strconv.Itoa(value))
}

// AddIntHeader 按十进制整数追加响应头。
func AddIntHeader(res Response, name string, value int) error {
	return AddHeader(res, name, strconv.Itoa(value))
}
