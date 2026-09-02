package async

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
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

func TestContextAwaitQuiescenceWaitsForTimedOutWorker(t *testing.T) {
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/slow", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	async, err := NewContext(context.Background(), req, newAsyncResponse(), WithTimeout(time.Millisecond))
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}

	started := make(chan struct{})
	release := make(chan struct{})
	async.Go(func(context.Context) error {
		close(started)
		<-release
		return nil
	})
	<-started

	if err := async.Await(context.Background()); !errors.Is(err, ErrTimeout) {
		t.Fatalf("Await err = %v, want ErrTimeout", err)
	}
	waitCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := async.AwaitQuiescence(waitCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("AwaitQuiescence err = %v, want context deadline exceeded", err)
	}

	close(release)
	if err := async.AwaitQuiescence(context.Background()); err != nil {
		t.Fatalf("AwaitQuiescence failed: %v", err)
	}
}

func TestContextCompleteIsIdempotentAndAwaitReturnsError(t *testing.T) {
	t.Parallel()

	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/slow", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	cause := errors.New("boom")
	var events []string
	async, err := NewContext(context.Background(), req, newAsyncResponse(), WithListener(ListenerFunc{
		Error: func(_ context.Context, event Event) {
			events = append(events, "error:"+event.Err.Error())
		},
		Complete: func(_ context.Context, event Event) {
			events = append(events, "complete:"+event.Err.Error())
		},
	}))
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}

	if err := async.Complete(cause); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if err := async.Await(context.Background()); !errors.Is(err, cause) {
		t.Fatalf("Await err = %v, want cause", err)
	}
	if err := async.Complete(nil); !errors.Is(err, ErrCompleted) {
		t.Fatalf("second Complete err = %v, want ErrCompleted", err)
	}
	if !async.Completed() {
		t.Fatal("context should report completed")
	}
	want := []string{"error:boom", "complete:boom"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestContextDispatchCountsAndRejectsAfterComplete(t *testing.T) {
	t.Parallel()

	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/start", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	async, err := NewContext(context.Background(), req, newAsyncResponse())
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}
	handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return nil
	})

	if err := async.Dispatch(handler); err != nil {
		t.Fatalf("first Dispatch failed: %v", err)
	}
	if err := async.Dispatch(handler); err != nil {
		t.Fatalf("second Dispatch failed: %v", err)
	}
	if async.DispatchCount() != 2 {
		t.Fatalf("dispatch count = %d, want 2", async.DispatchCount())
	}
	if err := async.Complete(nil); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if err := async.Dispatch(handler); !errors.Is(err, ErrCompleted) {
		t.Fatalf("Dispatch after complete err = %v, want ErrCompleted", err)
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

func TestStreamCloseWaitsForInFlightWrite(t *testing.T) {
	res := newBlockingAsyncResponse()
	stream, err := NewStream(res)
	if err != nil {
		t.Fatalf("NewStream failed: %v", err)
	}
	writeDone := make(chan error, 1)
	go func() {
		_, err := stream.Write(context.Background(), []byte("payload"))
		writeDone <- err
	}()
	<-res.writeStarted

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- stream.Close(context.Background())
	}()

	select {
	case <-res.flushCalled:
		t.Fatal("Close flushed before in-flight Write finished")
	case <-time.After(20 * time.Millisecond):
	}
	close(res.releaseWrite)
	if err := <-writeDone; err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := <-closeDone; err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if res.body.String() != "payload" {
		t.Fatalf("body = %q, want payload", res.body.String())
	}
}

type asyncResponse struct {
	header    servlet.Header
	status    int
	committed bool
	body      bytes.Buffer
	flushes   int
}

func newAsyncResponse() *asyncResponse {
	return &asyncResponse{header: servlet.NewHeader(), status: http.StatusOK}
}

func (r *asyncResponse) Header() servlet.Header {
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

type blockingAsyncResponse struct {
	asyncResponse
	writeStarted chan struct{}
	releaseWrite chan struct{}
	flushCalled  chan struct{}
	flushOnce    sync.Once
}

func newBlockingAsyncResponse() *blockingAsyncResponse {
	return &blockingAsyncResponse{
		asyncResponse: asyncResponse{header: servlet.NewHeader(), status: http.StatusOK},
		writeStarted:  make(chan struct{}),
		releaseWrite:  make(chan struct{}),
		flushCalled:   make(chan struct{}),
	}
}

func (r *blockingAsyncResponse) Write(data []byte) (int, error) {
	close(r.writeStarted)
	<-r.releaseWrite
	return r.asyncResponse.Write(data)
}

func (r *blockingAsyncResponse) Flush() error {
	r.flushOnce.Do(func() {
		close(r.flushCalled)
	})
	return r.asyncResponse.Flush()
}
