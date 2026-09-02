package servlet

import "io"

var _ Response = (*contractResponse)(nil)

type contractResponse struct {
	header Header
}

func (r *contractResponse) Header() Header               { return r.header }
func (*contractResponse) SetStatus(int)                  {}
func (*contractResponse) Status() int                    { return 200 }
func (*contractResponse) Write(data []byte) (int, error) { return len(data), nil }
func (*contractResponse) WriteString(value string) (int, error) {
	return len(value), nil
}
func (*contractResponse) Flush() error          { return nil }
func (*contractResponse) Committed() bool       { return false }
func (*contractResponse) Reset() error          { return nil }
func (*contractResponse) BodyWriter() io.Writer { return io.Discard }
