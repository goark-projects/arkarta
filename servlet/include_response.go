package servlet

import (
	"io"
	"net/http"
)

type includeResponse struct {
	target Response
	header http.Header
}

func newIncludeResponse(target Response) Response {
	if target == nil {
		return nil
	}
	return &includeResponse{
		target: target,
		header: make(http.Header),
	}
}

func (r *includeResponse) Header() http.Header {
	return r.header
}

func (r *includeResponse) SetStatus(int) {
}

func (r *includeResponse) Status() int {
	return r.target.Status()
}

func (r *includeResponse) Write(data []byte) (int, error) {
	return r.target.Write(data)
}

func (r *includeResponse) WriteString(value string) (int, error) {
	return r.target.WriteString(value)
}

func (r *includeResponse) Flush() error {
	return r.target.Flush()
}

func (r *includeResponse) Committed() bool {
	return r.target.Committed()
}

func (r *includeResponse) Reset() error {
	return ErrResponseCommitted
}

func (r *includeResponse) BodyWriter() io.Writer {
	return r
}
