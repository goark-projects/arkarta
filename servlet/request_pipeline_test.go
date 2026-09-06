package servlet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRequestAttributeListenerEvents(t *testing.T) {
	t.Parallel()
	var events []string
	listener := RequestAttributeListenerFunc{
		Added: func(_ context.Context, event RequestAttributeEvent) {
			events = append(events, "add:"+event.Name)
		},
		Replaced: func(_ context.Context, event RequestAttributeEvent) {
			events = append(events, "replace:"+event.Name)
		},
		Removed: func(_ context.Context, event RequestAttributeEvent) {
			events = append(events, "remove:"+event.Name)
		},
	}
	req, err := NewRequest(
		httptest.NewRequest(http.MethodGet, "/", nil),
		WithRequestAttributeListener(listener),
	)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	req.SetAttribute("trace", "a")
	req.SetAttribute("trace", "b")
	req.SetAttribute("trace", nil)

	want := []string{"add:trace", "replace:trace", "remove:trace"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestContextAttributeListenerEvents(t *testing.T) {
	t.Parallel()

	var events []string
	app, err := NewWebApp("orders", WithContextAttributeListener(ContextAttributeListenerFunc{
		Added: func(_ context.Context, event ContextAttributeEvent) {
			events = append(events, "add:"+event.Name)
		},
		Replaced: func(_ context.Context, event ContextAttributeEvent) {
			events = append(events, "replace:"+event.Name)
		},
		Removed: func(_ context.Context, event ContextAttributeEvent) {
			events = append(events, "remove:"+event.Name)
		},
	}))
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	app.SetAttribute("mode", "blue")
	app.SetAttribute("mode", "green")
	app.SetAttribute("mode", nil)

	want := []string{"add:mode", "replace:mode", "remove:mode"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestHandlerFuncServe(t *testing.T) {
	t.Parallel()

	called := false
	handler := HandlerFunc(func(ctx context.Context, req *Request, res Response) error {
		called = true
		if ctx == nil {
			t.Fatal("context 不能为空")
		}
		if req.Method() != http.MethodPost {
			t.Fatalf("method = %s, want %s", req.Method(), http.MethodPost)
		}
		return nil
	})

	req, err := NewRequest(httptest.NewRequest(http.MethodPost, "/orders", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if err := handler.Serve(context.Background(), req, nil); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}
	if !called {
		t.Fatal("handler 未被调用")
	}
}

func TestRequestAttributesAreIsolatedPerRequest(t *testing.T) {
	t.Parallel()

	first, err := NewRequest(httptest.NewRequest(http.MethodGet, "/first?q=1", nil))
	if err != nil {
		t.Fatalf("NewRequest first failed: %v", err)
	}
	second, err := NewRequest(httptest.NewRequest(http.MethodGet, "/second?q=2", nil))
	if err != nil {
		t.Fatalf("NewRequest second failed: %v", err)
	}

	first.SetAttribute("goark.dev/arkarta/servlet/test", "first")
	second.SetAttribute("goark.dev/arkarta/servlet/test", "second")

	value, ok := first.Attribute("goark.dev/arkarta/servlet/test")
	if !ok || value != "first" {
		t.Fatalf("first attribute = %v/%v, want first/true", value, ok)
	}
	value, ok = second.Attribute("goark.dev/arkarta/servlet/test")
	if !ok || value != "second" {
		t.Fatalf("second attribute = %v/%v, want second/true", value, ok)
	}
	if got := first.Query().Get("q"); got != "1" {
		t.Fatalf("first query = %s, want 1", got)
	}
}

func TestHTTPErrorImplementsStatusError(t *testing.T) {
	t.Parallel()

	cause := errors.New("database unavailable")
	err := NewHTTPError(http.StatusServiceUnavailable, "service unavailable", cause)

	var statusErr StatusError
	if !errors.As(err, &statusErr) {
		t.Fatal("HTTPError 未实现 StatusError")
	}
	if statusErr.StatusCode() != http.StatusServiceUnavailable {
		t.Fatalf("StatusCode = %d, want %d", statusErr.StatusCode(), http.StatusServiceUnavailable)
	}
	if statusErr.PublicMessage() != "service unavailable" {
		t.Fatalf("PublicMessage = %q", statusErr.PublicMessage())
	}
	if !errors.Is(err, cause) {
		t.Fatal("HTTPError 未保留底层 cause")
	}
}

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

	req, err := NewRequest(
		httptest.NewRequest(http.MethodGet, "/", nil),
		WithDispatchType(DispatchForward),
	)
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

	binding, err := NewFilterBinding(
		"audit", FilterFunc(noopBindingFilter), WithFilterInitParam("level", "full"),
	)
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
	binding, err := NewFilterBinding(
		"secure", recordFilter("secure", &calls), WithFilterURLPattern("/secure/*"),
	)
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
	_, err := NewFilterBinding(
		"bad", FilterFunc(noopBindingFilter),
		WithFilterDispatchTypes(DispatchTypes(1<<7)),
	)
	if !errors.Is(err, ErrInvalidDispatchTypes) {
		t.Fatalf("invalid dispatch err = %v, want ErrInvalidDispatchTypes", err)
	}
	_, err = NewFilterBinding(
		"bad", FilterFunc(noopBindingFilter), WithFilterInitParam("", "bad"),
	)
	if !errors.Is(err, ErrInvalidFilterConfig) {
		t.Fatalf("invalid init param err = %v, want ErrInvalidFilterConfig", err)
	}
	_, err = NewFilterBinding(
		"bad", FilterFunc(noopBindingFilter), WithFilterURLPattern("bad"),
	)
	if !errors.Is(err, ErrInvalidMappingPattern) {
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
