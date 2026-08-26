package nethttp

import (
	"errors"
	"io"
	"net/http"

	"goark.dev/arkarta/servlet"
)

// ErrFlushUnsupported 表示底层响应不支持 Flush。
var ErrFlushUnsupported = errors.New("arkarta/servlet/nethttp: flush unsupported")

// ErrInvalidBufferSize 表示响应缓冲区大小非法。
var ErrInvalidBufferSize = errors.New("arkarta/servlet/nethttp: invalid buffer size")

// Response 将标准库 http.ResponseWriter 包装为 servlet.Response。
type Response struct {
	writer          http.ResponseWriter
	status          int
	committed       bool
	bufferSize      int
	trailerSupplier servlet.TrailerFieldsFunc
	trailerFields   http.Header
}

// NewResponse 创建标准库响应适配器。
func NewResponse(writer http.ResponseWriter) *Response {
	return &Response{
		writer: writer,
		status: http.StatusOK,
	}
}

// Header 返回响应头。
func (r *Response) Header() http.Header {
	return r.writer.Header()
}

// SetStatus 设置 HTTP 状态码；响应提交后调用不会改变已发送状态。
func (r *Response) SetStatus(code int) {
	if r.committed {
		return
	}
	if code < 100 || code > 999 {
		code = http.StatusInternalServerError
	}
	r.status = code
}

// Status 返回当前 HTTP 状态码。
func (r *Response) Status() int {
	return r.status
}

// Write 写出响应体，首次写入会提交状态码。
func (r *Response) Write(data []byte) (int, error) {
	r.commit()
	return r.writer.Write(data)
}

// WriteString 写出字符串响应体。
func (r *Response) WriteString(value string) (int, error) {
	return r.Write([]byte(value))
}

// Flush 刷新响应缓冲区。
func (r *Response) Flush() error {
	r.commit()
	flusher, ok := r.writer.(http.Flusher)
	if !ok {
		return ErrFlushUnsupported
	}
	flusher.Flush()
	return nil
}

// Committed 表示响应是否已经提交。
func (r *Response) Committed() bool {
	return r.committed
}

// Reset 在响应提交前清空响应头并恢复默认状态码。
func (r *Response) Reset() error {
	if r.committed {
		return servlet.ErrResponseCommitted
	}
	header := r.writer.Header()
	for key := range header {
		delete(header, key)
	}
	r.status = http.StatusOK
	return nil
}

// SetBufferSize 设置响应缓冲区大小。
func (r *Response) SetBufferSize(size int) error {
	if r.committed {
		return servlet.ErrResponseCommitted
	}
	if size < 0 {
		return ErrInvalidBufferSize
	}
	r.bufferSize = size
	return nil
}

// BufferSize 返回响应缓冲区大小。
func (r *Response) BufferSize() int {
	return r.bufferSize
}

// ResetBuffer 在响应提交前清空响应体缓冲。
func (r *Response) ResetBuffer() error {
	if r.committed {
		return servlet.ErrResponseCommitted
	}
	return nil
}

// SetTrailerFields 设置响应 Trailer 字段供应器。
func (r *Response) SetTrailerFields(fields servlet.TrailerFieldsFunc) error {
	if r.committed {
		return servlet.ErrResponseCommitted
	}
	r.trailerSupplier = fields
	r.trailerFields = nil
	r.Header().Del("Trailer")
	if fields == nil {
		return nil
	}
	for name := range fields() {
		canonical := http.CanonicalHeaderKey(name)
		if canonical != "" {
			r.Header().Add("Trailer", canonical)
		}
	}
	return nil
}

// TrailerFields 返回响应 Trailer 字段副本。
func (r *Response) TrailerFields() http.Header {
	if r.trailerFields != nil {
		return r.trailerFields.Clone()
	}
	if r.trailerSupplier == nil {
		return http.Header{}
	}
	return r.trailerSupplier().Clone()
}

// BodyWriter 返回响应体写出器。
func (r *Response) BodyWriter() io.Writer {
	return r
}

func (r *Response) finish() {
	r.writeTrailers()
	if !r.committed {
		r.commit()
	}
}

func (r *Response) commit() {
	if r.committed {
		return
	}
	r.committed = true
	r.writer.WriteHeader(r.status)
}

func (r *Response) writeTrailers() {
	if r.trailerSupplier == nil {
		return
	}
	fields := r.trailerSupplier().Clone()
	r.trailerFields = fields
	for name, values := range fields {
		canonical := http.CanonicalHeaderKey(name)
		if canonical == "" {
			continue
		}
		r.Header().Del(canonical)
		for _, value := range values {
			r.Header().Add(canonical, value)
		}
	}
}
