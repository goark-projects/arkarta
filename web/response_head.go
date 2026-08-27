package web

import (
	"io"

	"goark.dev/arkarta/servlet"
)

type noBodyResponse struct {
	servlet.Response
}

func (r noBodyResponse) Write(data []byte) (int, error) {
	return len(data), nil
}

func (r noBodyResponse) WriteString(value string) (int, error) {
	return len(value), nil
}

func (r noBodyResponse) BodyWriter() io.Writer {
	return r
}
