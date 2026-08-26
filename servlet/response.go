package servlet

import (
	"errors"
	"io"
	"net/http"
)

// ErrResponseCommitted 表示响应已经提交，不能再执行重置类操作。
var ErrResponseCommitted = errors.New("arkarta/servlet: response is committed")

// Response 表示容器提供的响应写出能力。
type Response interface {
	Header() http.Header
	SetStatus(code int)
	Status() int
	Write([]byte) (int, error)
	WriteString(value string) (int, error)
	Flush() error
	Committed() bool
	Reset() error
	BodyWriter() io.Writer
}
