package tck

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"goark.dev/arkarta/servlet"
)

// RunCore 执行传输无关的 Servlet Core Profile 兼容性测试。
func RunCore(t *testing.T, driver Driver) {
	t.Helper()
	t.Run("writes_status_header_and_body", func(t *testing.T) {
		runWriteResponse(t, driver)
	})
	t.Run("commits_status_without_body", func(t *testing.T) {
		runCommitStatusWithoutBody(t, driver)
	})
	t.Run("supports_response_helpers", func(t *testing.T) {
		runResponseHelpers(t, driver)
	})
	t.Run("maps_status_error", func(t *testing.T) {
		runStatusError(t, driver)
	})
	t.Run("recovers_panic", func(t *testing.T) {
		runPanicRecovery(t, driver)
	})
	t.Run("preserves_filter_order", func(t *testing.T) {
		runFilterOrder(t, driver)
	})
	t.Run("uses_servlet_mapping_priority", func(t *testing.T) {
		runMappingPriority(t, driver)
	})
	t.Run("exposes_request_parameters", func(t *testing.T) {
		runRequestParameters(t, driver)
	})
	t.Run("exposes_request_cookies", func(t *testing.T) {
		runRequestCookies(t, driver)
	})
	t.Run("exposes_mapping_elements", func(t *testing.T) {
		runMappingElements(t, driver)
	})
}

func runWriteResponse(t *testing.T, driver Driver) {
	t.Helper()
	handler := servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		res.Header().Set("X-TCK", "core")
		res.SetStatus(http.StatusCreated)
		_, err := res.WriteString("ok")
		return err
	})

	response := exchange(t, driver, handler, NewRequest(http.MethodPost, "/tck"))

	if response.Status != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Status, http.StatusCreated)
	}
	if response.Header.Get("X-TCK") != "core" {
		t.Fatalf("X-TCK = %q, want core", response.Header.Get("X-TCK"))
	}
	if string(response.Body) != "ok" {
		t.Fatalf("body = %q, want ok", response.Body)
	}
}

func runCommitStatusWithoutBody(t *testing.T, driver Driver) {
	t.Helper()
	handler := servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		res.SetStatus(http.StatusNoContent)
		return nil
	})

	response := exchange(t, driver, handler, NewRequest(http.MethodDelete, "/resource"))

	if response.Status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Status, http.StatusNoContent)
	}
	if len(response.Body) != 0 {
		t.Fatalf("body = %q, want empty", response.Body)
	}
}

func runResponseHelpers(t *testing.T, driver Driver) {
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
		if err := servlet.AddCookie(res, &servlet.Cookie{Name: "sid", Value: "abc", HTTPOnly: true}); err != nil {
			return err
		}
		_, err := res.WriteString("ok")
		return err
	})

	response := exchange(t, driver, handler, NewRequest(http.MethodGet, "/helpers"))

	if response.Status != http.StatusOK || string(response.Body) != "ok" {
		t.Fatalf("status/body = %d/%q, want 200/ok", response.Status, response.Body)
	}
	if response.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q, want application/json charset", response.Header.Get("Content-Type"))
	}
	if response.Header.Get("Content-Length") != "2" {
		t.Fatalf("content length = %q, want 2", response.Header.Get("Content-Length"))
	}
	if got := response.Header.Get("Set-Cookie"); got != "sid=abc; HttpOnly" {
		t.Fatalf("set-cookie = %q, want sid cookie", got)
	}

	redirect := servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		return servlet.Redirect(res, "/login", http.StatusSeeOther)
	})
	response = exchange(t, driver, redirect, NewRequest(http.MethodGet, "/secure"))

	if response.Status != http.StatusSeeOther {
		t.Fatalf("redirect status = %d, want %d", response.Status, http.StatusSeeOther)
	}
	if response.Header.Get("Location") != "/login" {
		t.Fatalf("location = %q, want /login", response.Header.Get("Location"))
	}
}

func runStatusError(t *testing.T, driver Driver) {
	t.Helper()
	handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return servlet.NewHTTPError(http.StatusNotFound, "not found", nil)
	})

	response := exchange(t, driver, handler, NewRequest(http.MethodGet, "/missing"))

	if response.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Status, http.StatusNotFound)
	}
	if string(response.Body) != "not found\n" {
		t.Fatalf("body = %q, want not found newline", response.Body)
	}
}

func runPanicRecovery(t *testing.T, driver Driver) {
	t.Helper()
	handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		panic("boom")
	})

	response := exchange(t, driver, handler, NewRequest(http.MethodGet, "/panic"))

	if response.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Status, http.StatusInternalServerError)
	}
	if string(response.Body) != "Internal Server Error\n" {
		t.Fatalf("body = %q, want safe 500", response.Body)
	}
}

func runFilterOrder(t *testing.T, driver Driver) {
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

	exchange(t, driver, handler, NewRequest(http.MethodGet, "/filters"))

	want := []string{"a-before", "b-before", "handler", "b-after", "a-after"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func runMappingPriority(t *testing.T, driver Driver) {
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
		response := exchange(t, driver, router, NewRequest(http.MethodGet, path))
		if string(response.Body) != want {
			t.Fatalf("%s body = %q, want %q", path, response.Body, want)
		}
	}
}

func runRequestParameters(t *testing.T, driver Driver) {
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
		names, err := req.ParameterNames()
		if err != nil {
			t.Fatalf("ParameterNames failed: %v", err)
		}
		if wantNames := []string{"body", "q"}; !reflect.DeepEqual(names, wantNames) {
			t.Fatalf("parameter names = %#v, want %#v", names, wantNames)
		}
		body, _, err := req.Parameter("body")
		if err != nil {
			t.Fatalf("Parameter failed: %v", err)
		}
		_, err = res.WriteString(body)
		return err
	})

	request := NewRequest(http.MethodPost, "/params?q=query")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Body = []byte("q=form&body=ok")
	response := exchange(t, driver, handler, request)

	if response.Status != http.StatusOK || string(response.Body) != "ok" {
		t.Fatalf("status/body = %d/%q, want 200/ok", response.Status, response.Body)
	}
}

func runRequestCookies(t *testing.T, driver Driver) {
	t.Helper()
	handler := servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		cookies := req.Cookies()
		if len(cookies) != 2 || cookies[0].Name != "sid" || cookies[1].Name != "mode" {
			t.Fatalf("cookies = %#v, want sid/mode", cookies)
		}
		cookies[0].Value = "mutated"
		sid, err := req.Cookie("sid")
		if err != nil {
			t.Fatalf("Cookie failed: %v", err)
		}
		_, err = res.WriteString(sid.Value)
		return err
	})

	request := NewRequest(http.MethodGet, "/cookies")
	request.Header.Set("Cookie", "sid=abc; mode=dark")
	response := exchange(t, driver, handler, request)

	if response.Status != http.StatusOK || string(response.Body) != "abc" {
		t.Fatalf("status/body = %d/%q, want 200/abc", response.Status, response.Body)
	}
}

func runMappingElements(t *testing.T, driver Driver) {
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

	response := exchange(t, driver, router, NewRequest(http.MethodGet, "/api/orders"))

	if response.Status != http.StatusOK || string(response.Body) != "/api/*" {
		t.Fatalf("status/body = %d/%q, want 200//api/*", response.Status, response.Body)
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
