package tck

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
)

// RunDispatcher 执行 RequestDispatcher Core Profile 兼容性测试。
func RunDispatcher(t *testing.T) {
	t.Helper()
	t.Run("forward", func(t *testing.T) {
		router := servlet.NewRouter()
		mustHandle(t, router, "/target", servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
			if req.DispatchType() != servlet.DispatchForward {
				t.Fatalf("dispatch = %v, want forward", req.DispatchType())
			}
			_, err := res.WriteString("forward:" + req.Path())
			return err
		}))
		req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/source", nil))
		if err != nil {
			t.Fatalf("NewRequest failed: %v", err)
		}
		response := newMemoryResponse()
		dispatcher, err := servlet.NewRequestDispatcher(router, "/target")
		if err != nil {
			t.Fatalf("NewRequestDispatcher failed: %v", err)
		}
		if err := dispatcher.Forward(context.Background(), req, response); err != nil {
			t.Fatalf("Forward failed: %v", err)
		}
		if response.body.String() != "forward:/target" {
			t.Fatalf("body = %q, want forward:/target", response.body.String())
		}
	})
	t.Run("include", func(t *testing.T) {
		router := servlet.NewRouter()
		mustHandle(t, router, "/fragment", servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			res.SetStatus(http.StatusCreated)
			res.Header().Set("X-Leak", "true")
			_, err := res.WriteString("fragment")
			return err
		}))
		req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/page", nil))
		if err != nil {
			t.Fatalf("NewRequest failed: %v", err)
		}
		response := newMemoryResponse()
		response.SetStatus(http.StatusAccepted)
		dispatcher, err := servlet.NewRequestDispatcher(router, "/fragment")
		if err != nil {
			t.Fatalf("NewRequestDispatcher failed: %v", err)
		}
		if err := dispatcher.Include(context.Background(), req, response); err != nil {
			t.Fatalf("Include failed: %v", err)
		}
		if response.Status() != http.StatusAccepted || response.Header().Get("X-Leak") != "" {
			t.Fatalf("include leaked response metadata")
		}
	})
}

type memoryResponse struct {
	header    http.Header
	status    int
	committed bool
	body      bytes.Buffer
}

func newMemoryResponse() *memoryResponse {
	return &memoryResponse{
		header: make(http.Header),
		status: http.StatusOK,
	}
}

func (r *memoryResponse) Header() http.Header {
	return r.header
}

func (r *memoryResponse) SetStatus(code int) {
	if !r.committed {
		r.status = code
	}
}

func (r *memoryResponse) Status() int {
	return r.status
}

func (r *memoryResponse) Write(data []byte) (int, error) {
	r.committed = true
	return r.body.Write(data)
}

func (r *memoryResponse) WriteString(value string) (int, error) {
	return r.Write([]byte(value))
}

func (r *memoryResponse) Flush() error {
	r.committed = true
	return nil
}

func (r *memoryResponse) Committed() bool {
	return r.committed
}

func (r *memoryResponse) Reset() error {
	if r.committed {
		return servlet.ErrResponseCommitted
	}
	r.header = make(http.Header)
	r.status = http.StatusOK
	r.body.Reset()
	return nil
}

func (r *memoryResponse) BodyWriter() io.Writer {
	return r
}
