package async

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"goark.dev/arkarta/servlet"
)

func TestContextDispatchesAsyncRequest(t *testing.T) {
	t.Parallel()

	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/start", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	res := newAsyncResponse()
	var events []string
	async, err := NewContext(context.Background(), req, res, WithListener(ListenerFunc{
		Start: func(context.Context, Event) {
			events = append(events, "start")
		},
		Complete: func(context.Context, Event) {
			events = append(events, "complete")
		},
	}))
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}

	err = async.Dispatch(servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		if req.DispatchType() != servlet.DispatchAsync {
			t.Fatalf("dispatch = %v, want async", req.DispatchType())
		}
		_, writeErr := res.WriteString("async")
		return writeErr
	}))
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if err := async.Complete(nil); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if req.DispatchType() != servlet.DispatchRequest {
		t.Fatalf("dispatch restored = %v, want request", req.DispatchType())
	}
	want := []string{"start", "complete"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if res.body.String() != "async" {
		t.Fatalf("body = %q, want async", res.body.String())
	}
}

func TestContextTimeoutCompletes(t *testing.T) {
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/slow", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	var events []string
	async, err := NewContext(context.Background(), req, newAsyncResponse(),
		WithTimeout(time.Millisecond),
		WithListener(ListenerFunc{
			Timeout: func(context.Context, Event) {
				events = append(events, "timeout")
			},
			Complete: func(context.Context, Event) {
				events = append(events, "complete")
			},
		}),
	)
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}

	select {
	case <-async.Done():
	case <-time.After(time.Second):
		t.Fatal("async context did not complete after timeout")
	}
	if !errors.Is(async.Err(), ErrTimeout) {
		t.Fatalf("err = %v, want ErrTimeout", async.Err())
	}
	want := []string{"timeout", "complete"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestStreamWriteFlushAndClose(t *testing.T) {
	t.Parallel()

	res := newAsyncResponse()
	stream, err := NewStream(res)
	if err != nil {
		t.Fatalf("NewStream failed: %v", err)
	}
	if _, err := stream.Write(context.Background(), []byte("a")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := stream.Flush(context.Background()); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	if err := stream.Close(context.Background()); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if _, err := stream.Write(context.Background(), []byte("b")); !errors.Is(err, ErrCompleted) {
		t.Fatalf("write after close err = %v, want ErrCompleted", err)
	}
	if res.body.String() != "a" || res.flushes != 2 {
		t.Fatalf("stream result body=%q flushes=%d", res.body.String(), res.flushes)
	}
}

type asyncResponse struct {
	header    http.Header
	status    int
	committed bool
	body      bytes.Buffer
	flushes   int
}

func newAsyncResponse() *asyncResponse {
	return &asyncResponse{header: make(http.Header), status: http.StatusOK}
}

func (r *asyncResponse) Header() http.Header {
	return r.header
}

func (r *asyncResponse) SetStatus(code int) {
	r.status = code
}

func (r *asyncResponse) Status() int {
	return r.status
}

func (r *asyncResponse) Write(data []byte) (int, error) {
	r.committed = true
	return r.body.Write(data)
}

func (r *asyncResponse) WriteString(value string) (int, error) {
	return r.Write([]byte(value))
}

func (r *asyncResponse) Flush() error {
	r.committed = true
	r.flushes++
	return nil
}

func (r *asyncResponse) Committed() bool {
	return r.committed
}

func (r *asyncResponse) Reset() error {
	r.body.Reset()
	r.committed = false
	return nil
}

func (r *asyncResponse) BodyWriter() io.Writer {
	return r
}
