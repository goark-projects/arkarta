package tck

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/async"
)

// RunAsyncLifecycle 执行 Servlet Async Profile 的兼容性测试。
func RunAsyncLifecycle(t *testing.T) {
	t.Helper()
	t.Run("await_and_completion_state", runAsyncAwaitAndCompletionState)
	t.Run("dispatch_count_and_idempotency", runAsyncDispatchCountAndIdempotency)
	t.Run("timeout_event_order", runAsyncTimeoutEventOrder)
	t.Run("stream_rejects_after_close", runAsyncStreamRejectsAfterClose)
}

func runAsyncAwaitAndCompletionState(t *testing.T) {
	t.Helper()
	req := newTCKRequest(t, http.MethodGet, "/async")
	cause := errors.New("tck async failure")
	var events []string
	ctx, err := async.NewContext(context.Background(), req, newResponseStub(), async.WithListener(async.ListenerFunc{
		Error: func(_ context.Context, event async.Event) {
			events = append(events, "error:"+event.Err.Error())
		},
		Complete: func(_ context.Context, event async.Event) {
			events = append(events, "complete:"+event.Err.Error())
		},
	}))
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}

	if err := ctx.Complete(cause); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if err := ctx.Await(context.Background()); !errors.Is(err, cause) {
		t.Fatalf("Await err = %v, want cause", err)
	}
	if !ctx.Completed() {
		t.Fatal("Completed should be true after Complete")
	}
	if !reflect.DeepEqual(events, []string{"error:tck async failure", "complete:tck async failure"}) {
		t.Fatalf("events = %#v, want error then complete", events)
	}
}

func runAsyncDispatchCountAndIdempotency(t *testing.T) {
	t.Helper()
	req := newTCKRequest(t, http.MethodGet, "/dispatch")
	ctx, err := async.NewContext(context.Background(), req, newResponseStub())
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}
	handler := servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, _ servlet.Response) error {
		if req.DispatchType() != servlet.DispatchAsync {
			t.Fatalf("dispatch type = %v, want async", req.DispatchType())
		}
		return nil
	})

	if err := ctx.Dispatch(handler); err != nil {
		t.Fatalf("first Dispatch failed: %v", err)
	}
	if err := ctx.Dispatch(handler); err != nil {
		t.Fatalf("second Dispatch failed: %v", err)
	}
	if ctx.DispatchCount() != 2 {
		t.Fatalf("dispatch count = %d, want 2", ctx.DispatchCount())
	}
	if err := ctx.Complete(nil); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if err := ctx.Complete(nil); !errors.Is(err, async.ErrCompleted) {
		t.Fatalf("second Complete err = %v, want ErrCompleted", err)
	}
	if err := ctx.Dispatch(handler); !errors.Is(err, async.ErrCompleted) {
		t.Fatalf("Dispatch after complete err = %v, want ErrCompleted", err)
	}
	if req.DispatchType() != servlet.DispatchRequest {
		t.Fatalf("dispatch type restored = %v, want request", req.DispatchType())
	}
}

func runAsyncTimeoutEventOrder(t *testing.T) {
	t.Helper()
	req := newTCKRequest(t, http.MethodGet, "/timeout")
	var events []string
	ctx, err := async.NewContext(context.Background(), req, newResponseStub(),
		async.WithTimeout(time.Millisecond),
		async.WithListener(async.ListenerFunc{
			Timeout: func(context.Context, async.Event) {
				events = append(events, "timeout")
			},
			Complete: func(context.Context, async.Event) {
				events = append(events, "complete")
			},
		}),
	)
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}

	if err := ctx.Await(context.Background()); !errors.Is(err, async.ErrTimeout) {
		t.Fatalf("Await err = %v, want ErrTimeout", err)
	}
	if !reflect.DeepEqual(events, []string{"timeout", "complete"}) {
		t.Fatalf("events = %#v, want timeout then complete", events)
	}
}

func runAsyncStreamRejectsAfterClose(t *testing.T) {
	t.Helper()
	res := newResponseStub()
	stream, err := async.NewStream(res)
	if err != nil {
		t.Fatalf("NewStream failed: %v", err)
	}
	if _, err := stream.Write(context.Background(), []byte("data")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := stream.Close(context.Background()); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if _, err := stream.Write(context.Background(), []byte("x")); !errors.Is(err, async.ErrCompleted) {
		t.Fatalf("Write after close err = %v, want ErrCompleted", err)
	}
	if res.body.String() != "data" || res.flushes != 1 {
		t.Fatalf("stream body/flushes = %q/%d, want data/1", res.body.String(), res.flushes)
	}
}

func newTCKRequest(t *testing.T, method, target string, options ...servlet.RequestOption) *servlet.Request {
	t.Helper()
	req, err := servlet.NewRequest(httptest.NewRequest(method, target, nil), options...)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	return req
}
