package servlet

import (
	"errors"
	"mime"
	"net/http"
	"strconv"
)

// ErrNilResponse 表示响应对象为空。
var ErrNilResponse = errors.New("arkarta/servlet: response is nil")

// ErrNilCookie 表示响应 Cookie 为空。
var ErrNilCookie = errors.New("arkarta/servlet: cookie is nil")

// ErrInvalidRedirectStatus 表示重定向状态码不属于 3xx。
var ErrInvalidRedirectStatus = errors.New("arkarta/servlet: invalid redirect status")

// AddCookie 向响应写入 Set-Cookie 头。
func AddCookie(res Response, cookie *http.Cookie) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	if cookie == nil {
		return ErrNilCookie
	}
	res.Header().Add("Set-Cookie", cookie.String())
	return nil
}

// SetContentType 设置响应 Content-Type。
func SetContentType(res Response, contentType string) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	if contentType == "" {
		res.Header().Del("Content-Type")
		return nil
	}
	res.Header().Set("Content-Type", contentType)
	return nil
}

// ContentType 返回响应 Content-Type。
func ContentType(res Response) string {
	if res == nil {
		return ""
	}
	return res.Header().Get("Content-Type")
}

// SetCharacterEncoding 设置 Content-Type 中的 charset 参数。
func SetCharacterEncoding(res Response, charset string) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	contentType := res.Header().Get("Content-Type")
	if contentType == "" && charset == "" {
		return nil
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType == "" {
		mediaType = "text/plain"
		params = map[string]string{}
	}
	if charset == "" {
		delete(params, "charset")
	} else {
		params["charset"] = charset
	}
	res.Header().Set("Content-Type", mime.FormatMediaType(mediaType, params))
	return nil
}

// CharacterEncoding 返回 Content-Type 中声明的 charset。
func CharacterEncoding(res Response) string {
	if res == nil {
		return ""
	}
	_, params, err := mime.ParseMediaType(res.Header().Get("Content-Type"))
	if err != nil {
		return ""
	}
	return params["charset"]
}

// SetContentLength 设置响应 Content-Length；负数会删除该响应头。
func SetContentLength(res Response, length int64) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	if length < 0 {
		res.Header().Del("Content-Length")
		return nil
	}
	res.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	return nil
}

// Redirect 发送重定向响应。
func Redirect(res Response, location string, statusCode int) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	if statusCode == 0 {
		statusCode = http.StatusFound
	}
	if statusCode < 300 || statusCode > 399 {
		return ErrInvalidRedirectStatus
	}
	if err := res.Reset(); err != nil {
		return err
	}
	res.Header().Set("Location", location)
	res.SetStatus(statusCode)
	return nil
}

// SendError 发送标准错误响应。
func SendError(res Response, statusCode int, message string) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	if err := res.Reset(); err != nil {
		return err
	}
	if statusCode < 100 || statusCode > 999 {
		statusCode = http.StatusInternalServerError
	}
	if message == "" {
		message = http.StatusText(statusCode)
	}
	if message == "" {
		message = "HTTP error"
	}
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.SetStatus(statusCode)
	_, err := res.WriteString(message + "\n")
	return err
}

func ensureMutableResponse(res Response) error {
	if res == nil {
		return ErrNilResponse
	}
	if res.Committed() {
		return ErrResponseCommitted
	}
	return nil
}
