package servlet

import (
	"errors"
	"net/http"
)

// ErrUnsupportedResponseControl 表示容器未实现可选响应控制接口。
var ErrUnsupportedResponseControl = errors.New("arkarta/servlet: unsupported response control")

// TrailerFieldsFunc 延迟提供响应 Trailer 字段。
type TrailerFieldsFunc func() http.Header

// TrailerFieldsControl 表示响应支持 Trailer 字段控制。
type TrailerFieldsControl interface {
	SetTrailerFields(fields TrailerFieldsFunc) error
	TrailerFields() http.Header
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
func TrailerFields(res Response) http.Header {
	control, ok := res.(TrailerFieldsControl)
	if !ok {
		return http.Header{}
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
