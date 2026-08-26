package tck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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
	t.Run("commits_status_without_body", func(t *testing.T) {
		runCommitStatusWithoutBody(t, factory)
	})
	t.Run("supports_response_helpers", func(t *testing.T) {
		runResponseHelpers(t, factory)
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
	t.Run("exposes_request_parameters", func(t *testing.T) {
		runRequestParameters(t, factory)
	})
	t.Run("exposes_mapping_elements", func(t *testing.T) {
		runMappingElements(t, factory)
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

func runCommitStatusWithoutBody(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	handler := servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		res.SetStatus(http.StatusNoContent)
		return nil
	})

	recorder := httptest.NewRecorder()
	factory(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/resource", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if recorder.Body.String() != "" {
		t.Fatalf("body = %q, want empty", recorder.Body.String())
	}
}

func runResponseHelpers(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	handler := servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		if err := servlet.SetContentType(res, "application/json"); err != nil {
			return err
		}
		if err := servlet.SetCharacterEncoding(res, "utf-8"); err != nil {
			return err
		}
		if err := servlet.SetContentLength(res, 2); err != nil {
			return err
		}
		if err := servlet.AddCookie(res, &http.Cookie{Name: "sid", Value: "abc", HttpOnly: true}); err != nil {
			return err
		}
		_, err := res.WriteString("ok")
		return err
	})

	recorder := httptest.NewRecorder()
	factory(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/helpers", nil))

	if recorder.Code != http.StatusOK || recorder.Body.String() != "ok" {
		t.Fatalf("status/body = %d/%q, want 200/ok", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q, want application/json charset", recorder.Header().Get("Content-Type"))
	}
	if recorder.Header().Get("Content-Length") != "2" {
		t.Fatalf("content length = %q, want 2", recorder.Header().Get("Content-Length"))
	}
	if got := recorder.Header().Get("Set-Cookie"); got != "sid=abc; HttpOnly" {
		t.Fatalf("set-cookie = %q, want sid cookie", got)
	}

	redirect := servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		return servlet.Redirect(res, "/login", http.StatusSeeOther)
	})
	recorder = httptest.NewRecorder()
	factory(redirect).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/secure", nil))

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("redirect status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if recorder.Header().Get("Location") != "/login" {
		t.Fatalf("location = %q, want /login", recorder.Header().Get("Location"))
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

func runRequestParameters(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	handler := servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		values, ok, err := req.ParameterValues("q")
		if err != nil {
			t.Fatalf("ParameterValues failed: %v", err)
		}
		want := []string{"query", "form"}
		if !ok || !reflect.DeepEqual(values, want) {
			t.Fatalf("q values = %#v/%v, want %#v/true", values, ok, want)
		}
		body, _, err := req.Parameter("body")
		if err != nil {
			t.Fatalf("Parameter failed: %v", err)
		}
		_, err = res.WriteString(body)
		return err
	})

	request := httptest.NewRequest(http.MethodPost, "/params?q=query", strings.NewReader("q=form&body=ok"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	factory(handler).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "ok" {
		t.Fatalf("status/body = %d/%q, want 200/ok", recorder.Code, recorder.Body.String())
	}
}

func runMappingElements(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	router := servlet.NewRouter()
	mustHandle(t, router, "/api/*", servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		if req.Mapping().Type() != servlet.MappingPrefix {
			t.Fatalf("mapping type = %v, want prefix", req.Mapping().Type())
		}
		if req.ServletPath() != "/api" || req.PathInfo() != "/orders" {
			t.Fatalf("mapping paths = %q/%q, want /api//orders", req.ServletPath(), req.PathInfo())
		}
		_, err := res.WriteString(req.Mapping().Pattern())
		return err
	}))

	recorder := httptest.NewRecorder()
	factory(router).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/orders", nil))

	if recorder.Code != http.StatusOK || recorder.Body.String() != "/api/*" {
		t.Fatalf("status/body = %d/%q, want 200//api/*", recorder.Code, recorder.Body.String())
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
