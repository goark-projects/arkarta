package tck

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/resource"
)

// RunStaticResources 执行静态资源 default servlet 兼容性测试。
func RunStaticResources(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	t.Run("serves_static_resource", func(t *testing.T) {
		t.Helper()
		handler := staticResourceHandler(t)
		recorder := httptest.NewRecorder()
		factory(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/public/app.json", nil))
		if recorder.Code != http.StatusOK || recorder.Body.String() != `{"ok":true}` {
			t.Fatalf("status/body = %d/%q, want 200/json", recorder.Code, recorder.Body.String())
		}
		if recorder.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("content type = %q, want application/json", recorder.Header().Get("Content-Type"))
		}
	})
	t.Run("honors_static_head", func(t *testing.T) {
		t.Helper()
		handler := staticResourceHandler(t)
		recorder := httptest.NewRecorder()
		factory(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodHead, "/public/app.json", nil))
		if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 {
			t.Fatalf("status/body length = %d/%d, want 200/0", recorder.Code, recorder.Body.Len())
		}
	})
	t.Run("serves_welcome_file", func(t *testing.T) {
		t.Helper()
		handler := staticResourceHandler(t)
		recorder := httptest.NewRecorder()
		factory(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/docs", nil))
		if recorder.Code != http.StatusOK || recorder.Body.String() != "welcome" {
			t.Fatalf("status/body = %d/%q, want 200/welcome", recorder.Code, recorder.Body.String())
		}
	})
}

func staticResourceHandler(t *testing.T) servlet.Handler {
	t.Helper()
	provider, err := resource.NewFSProvider(fstest.MapFS{
		"public/app.json": &fstest.MapFile{
			Data:    []byte(`{"ok":true}`),
			ModTime: time.Date(2026, 8, 26, 10, 0, 0, 0, time.UTC),
		},
		"docs":            &fstest.MapFile{Mode: 0o755 | fs.ModeDir},
		"docs/index.html": &fstest.MapFile{Data: []byte("welcome")},
	})
	if err != nil {
		t.Fatalf("NewFSProvider failed: %v", err)
	}
	handler, err := resource.NewDefaultServlet(provider)
	if err != nil {
		t.Fatalf("NewDefaultServlet failed: %v", err)
	}
	return handler
}
