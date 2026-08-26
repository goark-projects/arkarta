package servlet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
