package servlet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterUsesServletMappingPriority(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	mustHandle(t, router, "/", appendName("default"))
	mustHandle(t, router, "*.json", appendName("extension"))
	mustHandle(t, router, "/api/*", appendName("prefix"))
	mustHandle(t, router, "/api/users/me", appendName("exact"))

	tests := []struct {
		path string
		want string
	}{
		{path: "/api/users/me", want: "exact"},
		{path: "/api/orders", want: "prefix"},
		{path: "/report.json", want: "extension"},
		{path: "/assets/style.css", want: "default"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			req, err := NewRequest(httptest.NewRequest(http.MethodGet, tt.path, nil))
			if err != nil {
				t.Fatalf("NewRequest failed: %v", err)
			}
			if err := router.Serve(context.Background(), req, nil); err != nil {
				t.Fatalf("Serve failed: %v", err)
			}
			value, ok := req.Attribute("handler")
			if !ok || value != tt.want {
				t.Fatalf("handler = %v/%v, want %s/true", value, ok, tt.want)
			}
		})
	}
}

func TestRouterAppliesServletMappingElements(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	mustHandle(t, router, "/orders/*", HandlerFunc(func(_ context.Context, req *Request, _ Response) error {
		if req.Mapping().Type() != MappingPrefix {
			t.Fatalf("mapping type = %v, want prefix", req.Mapping().Type())
		}
		if req.Mapping().Pattern() != "/orders/*" {
			t.Fatalf("mapping pattern = %q, want /orders/*", req.Mapping().Pattern())
		}
		if req.ServletPath() != "/orders" {
			t.Fatalf("servlet path = %q, want /orders", req.ServletPath())
		}
		if req.PathInfo() != "/42" {
			t.Fatalf("path info = %q, want /42", req.PathInfo())
		}
		return nil
	}))

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/orders/42", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if err := router.Serve(context.Background(), req, nil); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}
}

func TestRouterReturnsNotFoundWithoutDefaultMapping(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/missing", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	err = router.Serve(context.Background(), req, nil)
	var statusErr StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("err = %v, want StatusError", err)
	}
	if statusErr.StatusCode() != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", statusErr.StatusCode())
	}
}

func TestRouterRejectsDuplicateMapping(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	mustHandle(t, router, "/users", appendName("first"))
	err := router.Handle("/users", appendName("second"))
	if !errors.Is(err, ErrDuplicateMapping) {
		t.Fatalf("err = %v, want ErrDuplicateMapping", err)
	}
}

func mustHandle(t *testing.T, router *Router, pattern string, handler Handler) {
	t.Helper()
	if err := router.Handle(pattern, handler); err != nil {
		t.Fatalf("Handle(%q) failed: %v", pattern, err)
	}
}

func appendName(name string) Handler {
	return HandlerFunc(func(_ context.Context, req *Request, _ Response) error {
		req.SetAttribute("handler", name)
		return nil
	})
}
