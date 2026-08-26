package tck

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
)

// HTTPErrorPageFactory 创建带错误页注册表的 HTTP 处理器。
type HTTPErrorPageFactory func(handler servlet.Handler, registry *servlet.ErrorPageRegistry) http.Handler

// RunErrorPages 执行错误页兼容性测试。
func RunErrorPages(t *testing.T, factory HTTPErrorPageFactory) {
	t.Helper()
	t.Run("status_error_page", func(t *testing.T) {
		registry := servlet.NewErrorPageRegistry()
		if err := registry.RegisterStatus(http.StatusForbidden, servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
			if req.DispatchType() != servlet.DispatchError {
				t.Fatalf("dispatch = %v, want error", req.DispatchType())
			}
			_, err := res.WriteString("forbidden-page")
			return err
		})); err != nil {
			t.Fatalf("RegisterStatus failed: %v", err)
		}
		handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return servlet.NewHTTPError(http.StatusForbidden, "forbidden", nil)
		})

		recorder := httptest.NewRecorder()
		factory(handler, registry).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/secure", nil))

		if recorder.Code != http.StatusForbidden || recorder.Body.String() != "forbidden-page" {
			t.Fatalf("status/body = %d/%q", recorder.Code, recorder.Body.String())
		}
	})
	t.Run("panic_error_page", func(t *testing.T) {
		registry := servlet.NewErrorPageRegistry()
		if err := registry.RegisterStatus(http.StatusInternalServerError, servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
			if _, ok := req.Attribute(servlet.AttributeErrorException); !ok {
				t.Fatal("panic should attach error exception")
			}
			_, err := res.WriteString("panic-page")
			return err
		})); err != nil {
			t.Fatalf("RegisterStatus failed: %v", err)
		}
		handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			panic("boom")
		})

		recorder := httptest.NewRecorder()
		factory(handler, registry).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))

		if recorder.Code != http.StatusInternalServerError || recorder.Body.String() != "panic-page" {
			t.Fatalf("status/body = %d/%q", recorder.Code, recorder.Body.String())
		}
	})
	t.Run("default_error_page", func(t *testing.T) {
		registry := servlet.NewErrorPageRegistry()
		if err := registry.RegisterDefault(servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			_, err := res.WriteString("default-page")
			return err
		})); err != nil {
			t.Fatalf("RegisterDefault failed: %v", err)
		}
		handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return servlet.NewHTTPError(http.StatusTeapot, "teapot", nil)
		})

		recorder := httptest.NewRecorder()
		factory(handler, registry).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/default", nil))

		if recorder.Code != http.StatusTeapot || recorder.Body.String() != "default-page" {
			t.Fatalf("status/body = %d/%q, want default error page", recorder.Code, recorder.Body.String())
		}
	})
	t.Run("error_type_last_registration_wins", func(t *testing.T) {
		registry := servlet.NewErrorPageRegistry()
		if err := servlet.RegisterErrorType[*tckTypedError](registry, servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			_, err := res.WriteString("first")
			return err
		})); err != nil {
			t.Fatalf("RegisterErrorType first failed: %v", err)
		}
		if err := servlet.RegisterErrorType[*tckTypedError](registry, servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			_, err := res.WriteString("last")
			return err
		})); err != nil {
			t.Fatalf("RegisterErrorType last failed: %v", err)
		}
		handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return &tckTypedError{}
		})

		recorder := httptest.NewRecorder()
		factory(handler, registry).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/typed", nil))

		if recorder.Code != http.StatusInternalServerError || recorder.Body.String() != "last" {
			t.Fatalf("status/body = %d/%q, want last typed error page", recorder.Code, recorder.Body.String())
		}
	})
	t.Run("error_page_loop_is_rejected", func(t *testing.T) {
		registry := servlet.NewErrorPageRegistry()
		if err := registry.RegisterStatus(http.StatusInternalServerError, servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
			handled, err := registry.Handle(ctx, req, res, http.StatusInternalServerError, servlet.NewHTTPError(http.StatusInternalServerError, "nested", nil))
			if handled || !errors.Is(err, servlet.ErrErrorPageLoop) {
				t.Fatalf("nested Handle handled/err = %v/%v, want false/ErrErrorPageLoop", handled, err)
			}
			_, writeErr := res.WriteString("loop-guard")
			return writeErr
		})); err != nil {
			t.Fatalf("RegisterStatus failed: %v", err)
		}
		handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return servlet.NewHTTPError(http.StatusInternalServerError, "root", nil)
		})

		recorder := httptest.NewRecorder()
		factory(handler, registry).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/loop", nil))

		if recorder.Code != http.StatusInternalServerError || recorder.Body.String() != "loop-guard" {
			t.Fatalf("status/body = %d/%q, want guarded error page", recorder.Code, recorder.Body.String())
		}
	})
}

type tckTypedError struct{}

func (e *tckTypedError) Error() string {
	return "typed"
}
