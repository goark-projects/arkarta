package nethttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
)

func TestHandlerWritesServletResponse(t *testing.T) {
	t.Parallel()

	handler := Handler(servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		if req.Path() != "/orders" {
			t.Fatalf("path = %s, want /orders", req.Path())
		}
		res.Header().Set("X-Arkarta", "servlet")
		res.SetStatus(http.StatusCreated)
		_, err := res.WriteString("created")
		return err
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/orders", nil))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if recorder.Header().Get("X-Arkarta") != "servlet" {
		t.Fatalf("X-Arkarta = %q", recorder.Header().Get("X-Arkarta"))
	}
	if recorder.Body.String() != "created" {
		t.Fatalf("body = %q, want created", recorder.Body.String())
	}
}

func TestHandlerCommitsStatusWithoutBody(t *testing.T) {
	t.Parallel()

	handler := Handler(servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		res.SetStatus(http.StatusNoContent)
		return nil
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/resource", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want no content", recorder.Code)
	}
}

func TestHandlerSupportsResponseConvenienceHelpers(t *testing.T) {
	t.Parallel()

	handler := Handler(servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		if err := servlet.AddCookie(res, &http.Cookie{Name: "sid", Value: "abc", HttpOnly: true}); err != nil {
			return err
		}
		return servlet.Redirect(res, "/login", http.StatusSeeOther)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/secure", nil))

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want see other", recorder.Code)
	}
	if recorder.Header().Get("Location") != "/login" {
		t.Fatalf("location = %q, want /login", recorder.Header().Get("Location"))
	}
	if recorder.Header().Get("Set-Cookie") != "" {
		t.Fatalf("cookie must be reset by Redirect, got %q", recorder.Header().Get("Set-Cookie"))
	}
}

func TestHandlerMapsStatusError(t *testing.T) {
	t.Parallel()

	handler := Handler(servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return servlet.NewHTTPError(http.StatusForbidden, "forbidden", nil)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if recorder.Body.String() != "forbidden\n" {
		t.Fatalf("body = %q, want forbidden newline", recorder.Body.String())
	}
}

func TestHandlerUsesErrorPageRegistry(t *testing.T) {
	t.Parallel()

	registry := servlet.NewErrorPageRegistry()
	if err := registry.RegisterStatus(http.StatusForbidden, servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := res.WriteString("custom forbidden")
		return err
	})); err != nil {
		t.Fatalf("RegisterStatus failed: %v", err)
	}
	handler := HandlerWithOptions(servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return servlet.NewHTTPError(http.StatusForbidden, "forbidden", nil)
	}), WithErrorPages(registry))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if recorder.Body.String() != "custom forbidden" {
		t.Fatalf("body = %q, want custom forbidden", recorder.Body.String())
	}
}

func TestHandlerMapsWrappedStatusError(t *testing.T) {
	t.Parallel()

	handler := Handler(servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return errors.Join(servlet.NewHTTPError(http.StatusConflict, "conflict", nil), errors.New("write conflict"))
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
	if recorder.Body.String() != "conflict\n" {
		t.Fatalf("body = %q, want conflict newline", recorder.Body.String())
	}
}

func TestHandlerHidesPlainError(t *testing.T) {
	t.Parallel()

	handler := Handler(servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return errors.New("internal database detail")
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if recorder.Body.String() != "Internal Server Error\n" {
		t.Fatalf("body = %q, want safe 500", recorder.Body.String())
	}
}

func TestHandlerDispatchesPanicToErrorPage(t *testing.T) {
	t.Parallel()

	registry := servlet.NewErrorPageRegistry()
	if err := registry.RegisterStatus(http.StatusInternalServerError, servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		if req.DispatchType() != servlet.DispatchError {
			t.Fatalf("dispatch = %v, want error", req.DispatchType())
		}
		if _, ok := req.Attribute(servlet.AttributeErrorException); !ok {
			t.Fatal("panic error should be attached to request")
		}
		_, err := res.WriteString("panic-page")
		return err
	})); err != nil {
		t.Fatalf("RegisterStatus failed: %v", err)
	}
	handler := HandlerWithOptions(servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		panic("boom")
	}), WithErrorPages(registry))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if recorder.Body.String() != "panic-page" {
		t.Fatalf("body = %q, want panic-page", recorder.Body.String())
	}
}

func TestHandlerWritesTrailerFields(t *testing.T) {
	t.Parallel()

	handler := Handler(servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		if err := servlet.SetTrailerFields(res, func() servlet.Header {
			header := servlet.NewHeader()
			header.Set("X-Arkarta-Trailer", "done")
			return header
		}); err != nil {
			return err
		}
		_, err := res.WriteString("body")
		return err
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/trailers", nil))

	result := recorder.Result()
	if result.Trailer.Get("X-Arkarta-Trailer") != "done" {
		t.Fatalf("trailer = %q, want done", result.Trailer.Get("X-Arkarta-Trailer"))
	}
}
