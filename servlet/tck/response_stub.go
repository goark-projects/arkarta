package tck

import (
	"bytes"
	"io"
	"net/http"

	"goark.dev/arkarta/servlet"
)

// responseStub 是 TCK 内部使用的最小 Servlet Response 实现。
type responseStub struct {
	header    servlet.Header
	status    int
	body      bytes.Buffer
	committed bool
	flushes   int
}

func newResponseStub() *responseStub {
	return &responseStub{
		header: servlet.NewHeader(),
		status: http.StatusOK,
	}
}

func (r *responseStub) Header() servlet.Header {
	return r.header
}

func (r *responseStub) SetStatus(code int) {
	r.status = code
}

func (r *responseStub) Status() int {
	return r.status
}

func (r *responseStub) Write(data []byte) (int, error) {
	r.committed = true
	return r.body.Write(data)
}

func (r *responseStub) WriteString(value string) (int, error) {
	return r.Write([]byte(value))
}

func (r *responseStub) Flush() error {
	r.committed = true
	r.flushes++
	return nil
}

func (r *responseStub) Committed() bool {
	return r.committed
}

func (r *responseStub) Reset() error {
	r.body.Reset()
	r.committed = false
	return nil
}

func (r *responseStub) BodyWriter() io.Writer {
	return r
}
