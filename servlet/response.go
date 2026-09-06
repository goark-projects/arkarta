package servlet

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"
)

// ErrResponseCommitted 表示响应已经提交，不能再执行重置类操作。
var ErrResponseCommitted = errors.New("arkarta/servlet: response is committed")

// Response 表示容器提供的响应写出能力。
type Response interface {
	Header() Header
	SetStatus(code int)
	Status() int
	Write([]byte) (int, error)
	WriteString(value string) (int, error)
	Flush() error
	Committed() bool
	Reset() error
	BodyWriter() io.Writer
}

// ErrUnsupportedResponseControl 表示容器未实现可选响应控制接口。
var ErrUnsupportedResponseControl = errors.New("arkarta/servlet: unsupported response control")

// TrailerFieldsFunc 延迟提供响应 Trailer 字段。
type TrailerFieldsFunc func() Header

// TrailerFieldsControl 表示响应支持 Trailer 字段控制。
type TrailerFieldsControl interface {
	SetTrailerFields(fields TrailerFieldsFunc) error
	TrailerFields() Header
}

// BufferControl 表示响应支持缓冲区控制。
type BufferControl interface {
	SetBufferSize(size int) error
	BufferSize() int
	ResetBuffer() error
}

// SetTrailerFields 设置响应 Trailer 字段供应器。
func SetTrailerFields(res Response, fields TrailerFieldsFunc) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	control, ok := res.(TrailerFieldsControl)
	if !ok {
		return ErrUnsupportedResponseControl
	}
	return control.SetTrailerFields(fields)
}

// TrailerFields 返回响应 Trailer 字段副本。
func TrailerFields(res Response) Header {
	control, ok := res.(TrailerFieldsControl)
	if !ok {
		return NewHeader()
	}
	return control.TrailerFields()
}

// SetBufferSize 设置响应缓冲区大小。
func SetBufferSize(res Response, size int) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	control, ok := res.(BufferControl)
	if !ok {
		return ErrUnsupportedResponseControl
	}
	return control.SetBufferSize(size)
}

// BufferSize 返回响应缓冲区大小。
func BufferSize(res Response) int {
	control, ok := res.(BufferControl)
	if !ok {
		return 0
	}
	return control.BufferSize()
}

// ResetBuffer 在未提交前清空响应体缓冲。
func ResetBuffer(res Response) error {
	if err := ensureMutableResponse(res); err != nil {
		return err
	}
	control, ok := res.(BufferControl)
	if !ok {
		return ErrUnsupportedResponseControl
	}
	return control.ResetBuffer()
}

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

// ErrNilResponse 表示响应对象为空。
var ErrNilResponse = errors.New("arkarta/servlet: response is nil")

// ErrNilCookie 表示响应 Cookie 为空。
var ErrNilCookie = errors.New("arkarta/servlet: cookie is nil")

// ErrInvalidRedirectStatus 表示重定向状态码不属于 3xx。
var ErrInvalidRedirectStatus = errors.New("arkarta/servlet: invalid redirect status")

// AddCookie 向响应写入 Set-Cookie 头。
func AddCookie(res Response, cookie *Cookie) error {
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
		res.Header().Delete("Content-Type")
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
		res.Header().Delete("Content-Length")
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

// SetLocale 设置响应 Content-Language。
func SetLocale(res Response, locale Locale) error {
	if locale.Tag() == "" {
		return SetHeader(res, "Content-Language", "")
	}
	return SetHeader(res, "Content-Language", locale.Tag())
}

// ResponseLocale 返回响应 Content-Language。
func ResponseLocale(res Response) (Locale, bool) {
	value, ok := HeaderValue(res, "Content-Language")
	if !ok {
		return Locale{}, false
	}
	return NewLocale(value)
}
