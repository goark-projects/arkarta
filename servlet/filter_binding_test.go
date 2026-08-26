package servlet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestChainFilterBindingsHonorsDispatchType(t *testing.T) {
	t.Parallel()

	var calls []string
	target := HandlerFunc(func(context.Context, *Request, Response) error {
		calls = append(calls, "handler")
		return nil
	})
	requestBinding, err := BindFilter(recordFilter("request", &calls), DispatchRequest)
	if err != nil {
		t.Fatalf("BindFilter request failed: %v", err)
	}
	forwardBinding, err := BindFilter(recordFilter("forward", &calls), DispatchForward)
	if err != nil {
		t.Fatalf("BindFilter forward failed: %v", err)
	}
	handler := ChainFilterBindings(target, requestBinding, forwardBinding)

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/", nil), WithDispatchType(DispatchForward))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if err := handler.Serve(context.Background(), req, nil); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}

	want := []string{"forward", "handler"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestFilterBindingDefaultsToRequestDispatcher(t *testing.T) {
	t.Parallel()

	binding, err := NewFilterBinding("audit", FilterFunc(noopBindingFilter), WithFilterInitParam("level", "full"))
	if err != nil {
		t.Fatalf("NewFilterBinding failed: %v", err)
	}
	if binding.Name() != "audit" {
		t.Fatalf("name = %q, want audit", binding.Name())
	}
	if !binding.Matches(DispatchRequest) || binding.Matches(DispatchForward) {
		t.Fatal("default binding should match only request dispatcher")
	}
	params := binding.InitParams()
	params["level"] = "changed"
	if binding.InitParams()["level"] != "full" {
		t.Fatal("init params should be isolated")
	}
}

func TestFilterBindingHonorsURLPattern(t *testing.T) {
	t.Parallel()

	var calls []string
	target := HandlerFunc(func(context.Context, *Request, Response) error {
		calls = append(calls, "handler")
		return nil
	})
	binding, err := NewFilterBinding("secure", recordFilter("secure", &calls), WithFilterURLPattern("/secure/*"))
	if err != nil {
		t.Fatalf("NewFilterBinding failed: %v", err)
	}
	handler := ChainFilterBindings(target, binding)

	publicReq, err := NewRequest(httptest.NewRequest(http.MethodGet, "/public/index.html", nil))
	if err != nil {
		t.Fatalf("NewRequest public failed: %v", err)
	}
	if err := handler.Serve(context.Background(), publicReq, nil); err != nil {
		t.Fatalf("Serve public failed: %v", err)
	}
	secureReq, err := NewRequest(httptest.NewRequest(http.MethodGet, "/secure/orders", nil))
	if err != nil {
		t.Fatalf("NewRequest secure failed: %v", err)
	}
	if err := handler.Serve(context.Background(), secureReq, nil); err != nil {
		t.Fatalf("Serve secure failed: %v", err)
	}

	want := []string{"handler", "secure", "handler"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
	if binding.URLPattern() != "/secure/*" {
		t.Fatalf("url pattern = %q, want /secure/*", binding.URLPattern())
	}
}

func TestFilterBindingRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	var nilFilter FilterFunc
	if _, err := NewFilterBinding("nil", nilFilter); !errors.Is(err, ErrNilFilter) {
		t.Fatalf("nil filter err = %v, want ErrNilFilter", err)
	}
	if _, err := NewFilterBinding("bad", FilterFunc(noopBindingFilter), WithFilterDispatchTypes(DispatchTypes(1<<7))); !errors.Is(err, ErrInvalidDispatchTypes) {
		t.Fatalf("invalid dispatch err = %v, want ErrInvalidDispatchTypes", err)
	}
	if _, err := NewFilterBinding("bad", FilterFunc(noopBindingFilter), WithFilterInitParam("", "bad")); !errors.Is(err, ErrInvalidFilterConfig) {
		t.Fatalf("invalid init param err = %v, want ErrInvalidFilterConfig", err)
	}
	if _, err := NewFilterBinding("bad", FilterFunc(noopBindingFilter), WithFilterURLPattern("bad")); !errors.Is(err, ErrInvalidMappingPattern) {
		t.Fatalf("invalid URL pattern err = %v, want ErrInvalidMappingPattern", err)
	}
}

func TestDispatchTypesList(t *testing.T) {
	t.Parallel()

	dispatchers, err := NewDispatchTypes(DispatchError, DispatchRequest)
	if err != nil {
		t.Fatalf("NewDispatchTypes failed: %v", err)
	}
	want := []DispatchType{DispatchRequest, DispatchError}
	if !reflect.DeepEqual(dispatchers.List(), want) {
		t.Fatalf("list = %#v, want %#v", dispatchers.List(), want)
	}
	if _, err := NewDispatchTypes(DispatchType(99)); !errors.Is(err, ErrInvalidDispatchTypes) {
		t.Fatalf("invalid dispatcher err = %v, want ErrInvalidDispatchTypes", err)
	}
}

func recordFilter(name string, calls *[]string) Filter {
	return FilterFunc(func(ctx context.Context, req *Request, res Response, chain Chain) error {
		*calls = append(*calls, name)
		return chain.Next(ctx, req, res)
	})
}

func noopBindingFilter(ctx context.Context, req *Request, res Response, chain Chain) error {
	return chain.Next(ctx, req, res)
}
