package servlet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestFilterChainExecutesInOrder(t *testing.T) {
	t.Parallel()

	var calls []string
	handler := ChainFilters(HandlerFunc(func(context.Context, *Request, Response) error {
		calls = append(calls, "handler")
		return nil
	}), FilterFunc(func(ctx context.Context, req *Request, res Response, chain Chain) error {
		calls = append(calls, "first-before")
		if err := chain.Next(ctx, req, res); err != nil {
			return err
		}
		calls = append(calls, "first-after")
		return nil
	}), FilterFunc(func(ctx context.Context, req *Request, res Response, chain Chain) error {
		calls = append(calls, "second-before")
		if err := chain.Next(ctx, req, res); err != nil {
			return err
		}
		calls = append(calls, "second-after")
		return nil
	}))

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if err := handler.Serve(context.Background(), req, nil); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	want := []string{"first-before", "second-before", "handler", "second-after", "first-after"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestFilterChainCanShortCircuit(t *testing.T) {
	t.Parallel()

	called := false
	handler := ChainFilters(HandlerFunc(func(context.Context, *Request, Response) error {
		called = true
		return nil
	}), FilterFunc(func(context.Context, *Request, Response, Chain) error {
		return NewHTTPError(http.StatusUnauthorized, "unauthorized", nil)
	}))

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	err = handler.Serve(context.Background(), req, nil)
	if err == nil {
		t.Fatal("Serve should return short-circuit error")
	}
	if called {
		t.Fatal("短路过滤器后不应调用目标处理器")
	}
}

func TestFilterChainRejectsDoubleNext(t *testing.T) {
	t.Parallel()

	handler := ChainFilters(HandlerFunc(func(context.Context, *Request, Response) error {
		return nil
	}), FilterFunc(func(ctx context.Context, req *Request, res Response, chain Chain) error {
		if err := chain.Next(ctx, req, res); err != nil {
			return err
		}
		return chain.Next(ctx, req, res)
	}))

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	err = handler.Serve(context.Background(), req, nil)
	if !errors.Is(err, ErrChainAlreadyAdvanced) {
		t.Fatalf("err = %v, want ErrChainAlreadyAdvanced", err)
	}
}
