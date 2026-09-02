package tck

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/resource"
)

// HTTPHandlerFactory 将 Servlet Handler 暴露为标准库 http.Handler。
type HTTPHandlerFactory func(servlet.Handler) http.Handler

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
	t.Run("honors_static_range", func(t *testing.T) {
		t.Helper()
		handler := staticResourceHandler(t)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/public/app.json", nil)
		request.Header.Set("Range", "bytes=1-4")
		factory(handler).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusPartialContent || recorder.Body.String() != `"ok"` {
			t.Fatalf("status/body = %d/%q, want 206/range", recorder.Code, recorder.Body.String())
		}
		if recorder.Header().Get("Content-Range") != "bytes 1-4/11" {
			t.Fatalf("content range = %q, want bytes 1-4/11", recorder.Header().Get("Content-Range"))
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
	t.Run("if_range_mismatch_serves_full_body", func(t *testing.T) {
		t.Helper()
		handler := staticResourceHandler(t)
		request := httptest.NewRequest(http.MethodGet, "/public/app.json", nil)
		request.Header.Set("Range", "bytes=1-4")
		request.Header.Set("If-Range", `"strong-but-stale"`)
		recorder := httptest.NewRecorder()
		factory(handler).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK || recorder.Body.String() != `{"ok":true}` {
			t.Fatalf("status/body = %d/%q, want 200/full body", recorder.Code, recorder.Body.String())
		}
		if recorder.Header().Get("Content-Range") != "" {
			t.Fatalf("Content-Range = %q, want empty", recorder.Header().Get("Content-Range"))
		}
	})
	t.Run("weak_if_range_does_not_allow_range", func(t *testing.T) {
		t.Helper()
		handler := staticResourceHandler(t)
		request := httptest.NewRequest(http.MethodGet, "/public/app.json", nil)
		request.Header.Set("Range", "bytes=1-4")
		request.Header.Set("If-Range", `W/"weak"`)
		recorder := httptest.NewRecorder()
		factory(handler).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK || recorder.Body.String() != `{"ok":true}` {
			t.Fatalf("status/body = %d/%q, want 200/full body", recorder.Code, recorder.Body.String())
		}
	})
	t.Run("serves_multiple_ranges", func(t *testing.T) {
		t.Helper()
		handler := staticResourceHandler(t)
		request := httptest.NewRequest(http.MethodGet, "/public/app.json", nil)
		request.Header.Set("Range", "bytes=0-0,2-3")
		recorder := httptest.NewRecorder()
		factory(handler).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusPartialContent {
			t.Fatalf("status = %d, want 206", recorder.Code)
		}
		if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "multipart/byteranges; boundary=") {
			t.Fatalf("Content-Type = %q, want multipart/byteranges", contentType)
		}
		body := recorder.Body.String()
		if !strings.Contains(body, "Content-Range: bytes 0-0/11") ||
			!strings.Contains(body, "Content-Range: bytes 2-3/11") {
			t.Fatalf("multi-range body = %q, want both content ranges", body)
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
