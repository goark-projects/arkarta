package tck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"goark.dev/arkarta/servlet"
)

// HTTPHandlerFactory 将 Servlet Handler 暴露为标准库 http.Handler。
type HTTPHandlerFactory func(servlet.Handler) http.Handler

// RunCoreHTTP 执行面向 net/http 互操作容器的 Core Profile 兼容性测试。
func RunCoreHTTP(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	t.Run("writes_status_header_and_body", func(t *testing.T) {
		runWriteResponse(t, factory)
	})
	t.Run("maps_status_error", func(t *testing.T) {
		runStatusError(t, factory)
	})
	t.Run("recovers_panic", func(t *testing.T) {
		runPanicRecovery(t, factory)
	})
	t.Run("preserves_filter_order", func(t *testing.T) {
		runFilterOrder(t, factory)
	})
	t.Run("uses_servlet_mapping_priority", func(t *testing.T) {
		runMappingPriority(t, factory)
	})
}

func runWriteResponse(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	handler := servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		res.Header().Set("X-TCK", "core")
		res.SetStatus(http.StatusCreated)
		_, err := res.WriteString("ok")
		return err
	})

	recorder := httptest.NewRecorder()
	factory(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/tck", nil))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if recorder.Header().Get("X-TCK") != "core" {
		t.Fatalf("X-TCK = %q, want core", recorder.Header().Get("X-TCK"))
	}
	if recorder.Body.String() != "ok" {
		t.Fatalf("body = %q, want ok", recorder.Body.String())
	}
}

func runStatusError(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return servlet.NewHTTPError(http.StatusNotFound, "not found", nil)
	})

	recorder := httptest.NewRecorder()
	factory(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/missing", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	if recorder.Body.String() != "not found\n" {
		t.Fatalf("body = %q, want not found newline", recorder.Body.String())
	}
}

func runPanicRecovery(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		panic("boom")
	})

	recorder := httptest.NewRecorder()
	factory(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if recorder.Body.String() != "Internal Server Error\n" {
		t.Fatalf("body = %q, want safe 500", recorder.Body.String())
	}
}

func runFilterOrder(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	var calls []string
	handler := servlet.ChainFilters(
		servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			calls = append(calls, "handler")
			return nil
		}),
		servlet.FilterFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
			calls = append(calls, "a-before")
			if err := chain.Next(ctx, req, res); err != nil {
				return err
			}
			calls = append(calls, "a-after")
			return nil
		}),
		servlet.FilterFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
			calls = append(calls, "b-before")
			if err := chain.Next(ctx, req, res); err != nil {
				return err
			}
			calls = append(calls, "b-after")
			return nil
		}),
	)

	factory(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/filters", nil))

	want := []string{"a-before", "b-before", "handler", "b-after", "a-after"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func runMappingPriority(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	router := servlet.NewRouter()
	mustHandle(t, router, "/", markHandler("default"))
	mustHandle(t, router, "*.json", markHandler("extension"))
	mustHandle(t, router, "/api/*", markHandler("prefix"))
	mustHandle(t, router, "/api/users/me", markHandler("exact"))

	tests := map[string]string{
		"/api/users/me": "exact",
		"/api/orders":   "prefix",
		"/report.json":  "extension",
		"/health":       "default",
	}
	for path, want := range tests {
		recorder := httptest.NewRecorder()
		factory(router).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Body.String() != want {
			t.Fatalf("%s body = %q, want %q", path, recorder.Body.String(), want)
		}
	}
}

func markHandler(value string) servlet.Handler {
	return servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		_, err := res.WriteString(value)
		return err
	})
}

func mustHandle(t *testing.T, router *servlet.Router, pattern string, handler servlet.Handler) {
	t.Helper()
	if err := router.Handle(pattern, handler); err != nil {
		t.Fatalf("Handle(%q) failed: %v", pattern, err)
	}
}
