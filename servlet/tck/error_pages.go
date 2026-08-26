package tck

import (
	"context"
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
}
