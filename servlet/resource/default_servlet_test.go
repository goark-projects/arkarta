package resource

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/nethttp"
)

func TestDefaultServletServesGETAndHEAD(t *testing.T) {
	t.Parallel()

	handler := newTestDefaultServlet(t)

	getRecorder := httptest.NewRecorder()
	getReq, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/assets/app.json", nil))
	if err != nil {
		t.Fatalf("NewRequest GET failed: %v", err)
	}
	if err := handler.Serve(context.Background(), getReq, nethttp.NewResponse(getRecorder)); err != nil {
		t.Fatalf("Serve GET failed: %v", err)
	}
	if getRecorder.Code != http.StatusOK || getRecorder.Body.String() != `{"ok":true}` {
		t.Fatalf("GET status/body = %d/%q, want 200/json", getRecorder.Code, getRecorder.Body.String())
	}
	if getRecorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("GET content type = %q, want application/json", getRecorder.Header().Get("Content-Type"))
	}
	if getRecorder.Header().Get("Content-Length") != "11" {
		t.Fatalf("GET content length = %q, want 11", getRecorder.Header().Get("Content-Length"))
	}
	if getRecorder.Header().Get("ETag") == "" || getRecorder.Header().Get("Last-Modified") == "" {
		t.Fatal("GET should set ETag and Last-Modified")
	}

	headRecorder := httptest.NewRecorder()
	headReq, err := servlet.NewRequest(httptest.NewRequest(http.MethodHead, "/assets/app.json", nil))
	if err != nil {
		t.Fatalf("NewRequest HEAD failed: %v", err)
	}
	if err := handler.Serve(context.Background(), headReq, nethttp.NewResponse(headRecorder)); err != nil {
		t.Fatalf("Serve HEAD failed: %v", err)
	}
	if headRecorder.Code != http.StatusOK || headRecorder.Body.Len() != 0 {
		t.Fatalf("HEAD status/body length = %d/%d, want 200/0", headRecorder.Code, headRecorder.Body.Len())
	}
	if headRecorder.Header().Get("Content-Length") != "11" {
		t.Fatalf("HEAD content length = %q, want 11", headRecorder.Header().Get("Content-Length"))
	}
}

func TestDefaultServletHandlesConditionalGET(t *testing.T) {
	t.Parallel()

	handler := newTestDefaultServlet(t)
	firstRecorder := httptest.NewRecorder()
	firstReq, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/assets/app.json", nil))
	if err != nil {
		t.Fatalf("NewRequest first failed: %v", err)
	}
	if err := handler.Serve(context.Background(), firstReq, nethttp.NewResponse(firstRecorder)); err != nil {
		t.Fatalf("Serve first failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/assets/app.json", nil)
	request.Header.Set("If-None-Match", `"other", `+firstRecorder.Header().Get("ETag"))
	conditionalReq, err := servlet.NewRequest(request)
	if err != nil {
		t.Fatalf("NewRequest conditional failed: %v", err)
	}
	conditionalRecorder := httptest.NewRecorder()
	conditionalResponse := nethttp.NewResponse(conditionalRecorder)
	if err := handler.Serve(context.Background(), conditionalReq, conditionalResponse); err != nil {
		t.Fatalf("Serve conditional failed: %v", err)
	}

	if conditionalResponse.Status() != http.StatusNotModified || conditionalRecorder.Body.Len() != 0 {
		t.Fatalf("conditional status/body = %d/%q, want 304/empty", conditionalResponse.Status(), conditionalRecorder.Body.String())
	}
}

func TestDefaultServletHandlesRangeRequests(t *testing.T) {
	t.Parallel()

	handler := newTestDefaultServlet(t)
	request := httptest.NewRequest(http.MethodGet, "/assets/app.json", nil)
	request.Header.Set("Range", "bytes=1-4")
	req, err := servlet.NewRequest(request)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	recorder := httptest.NewRecorder()
	response := nethttp.NewResponse(recorder)
	if err := handler.Serve(context.Background(), req, response); err != nil {
		t.Fatalf("Serve range failed: %v", err)
	}

	if response.Status() != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", response.Status())
	}
	if recorder.Body.String() != `"ok"` {
		t.Fatalf("range body = %q, want %q", recorder.Body.String(), `"ok"`)
	}
	if recorder.Header().Get("Content-Range") != "bytes 1-4/11" {
		t.Fatalf("Content-Range = %q", recorder.Header().Get("Content-Range"))
	}
	if recorder.Header().Get("Content-Length") != "4" {
		t.Fatalf("Content-Length = %q, want 4", recorder.Header().Get("Content-Length"))
	}
}

func TestDefaultServletIgnoresRangeWhenIfRangeDoesNotMatch(t *testing.T) {
	t.Parallel()

	handler := newTestDefaultServlet(t)
	request := httptest.NewRequest(http.MethodGet, "/assets/app.json", nil)
	request.Header.Set("Range", "bytes=1-4")
	request.Header.Set("If-Range", `"other"`)
	req, err := servlet.NewRequest(request)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	recorder := httptest.NewRecorder()
	response := nethttp.NewResponse(recorder)
	if err := handler.Serve(context.Background(), req, response); err != nil {
		t.Fatalf("Serve If-Range failed: %v", err)
	}

	if response.Status() != http.StatusOK || recorder.Body.String() != `{"ok":true}` {
		t.Fatalf("status/body = %d/%q, want 200/full", response.Status(), recorder.Body.String())
	}
	if recorder.Header().Get("Content-Range") != "" {
		t.Fatalf("Content-Range = %q, want empty", recorder.Header().Get("Content-Range"))
	}
}

func TestDefaultServletIgnoresRangeForWeakETagValidator(t *testing.T) {
	t.Parallel()

	handler := newTestDefaultServlet(t)
	request := httptest.NewRequest(http.MethodGet, "/assets/app.json", nil)
	request.Header.Set("Range", "bytes=1-4")
	request.Header.Set("If-Range", `W/"weak"`)
	req, err := servlet.NewRequest(request)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	recorder := httptest.NewRecorder()
	response := nethttp.NewResponse(recorder)
	if err := handler.Serve(context.Background(), req, response); err != nil {
		t.Fatalf("Serve weak If-Range failed: %v", err)
	}

	if response.Status() != http.StatusOK || recorder.Body.String() != `{"ok":true}` {
		t.Fatalf("status/body = %d/%q, want 200/full", response.Status(), recorder.Body.String())
	}
}

func TestDefaultServletHandlesMultipleRanges(t *testing.T) {
	t.Parallel()

	handler := newTestDefaultServlet(t)
	request := httptest.NewRequest(http.MethodGet, "/assets/app.json", nil)
	request.Header.Set("Range", "bytes=0-0,2-3")
	req, err := servlet.NewRequest(request)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	recorder := httptest.NewRecorder()
	response := nethttp.NewResponse(recorder)
	if err := handler.Serve(context.Background(), req, response); err != nil {
		t.Fatalf("Serve multi-range failed: %v", err)
	}

	if response.Status() != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", response.Status())
	}
	contentType := recorder.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/byteranges; boundary=") {
		t.Fatalf("Content-Type = %q, want multipart/byteranges", contentType)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "Content-Range: bytes 0-0/11") ||
		!strings.Contains(body, "Content-Range: bytes 2-3/11") ||
		!strings.Contains(body, "{") ||
		!strings.Contains(body, "ok") {
		t.Fatalf("multi-range body = %q", body)
	}
}

func TestDefaultServletRejectsInvalidRange(t *testing.T) {
	t.Parallel()

	handler := newTestDefaultServlet(t)
	request := httptest.NewRequest(http.MethodGet, "/assets/app.json", nil)
	request.Header.Set("Range", "bytes=99-100")
	req, err := servlet.NewRequest(request)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	recorder := httptest.NewRecorder()
	response := nethttp.NewResponse(recorder)
	if err := handler.Serve(context.Background(), req, response); err != nil {
		t.Fatalf("Serve range failed: %v", err)
	}

	if response.Status() != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("status = %d, want 416", response.Status())
	}
	if recorder.Header().Get("Content-Range") != "bytes */11" {
		t.Fatalf("Content-Range = %q, want bytes */11", recorder.Header().Get("Content-Range"))
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("body length = %d, want 0", recorder.Body.Len())
	}
}

func TestDefaultServletMapsMissingAndDirectoryToNotFound(t *testing.T) {
	t.Parallel()

	handler := newTestDefaultServlet(t)
	for _, target := range []string{"/missing.txt", "/assets"} {
		req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, target, nil))
		if err != nil {
			t.Fatalf("NewRequest %s failed: %v", target, err)
		}
		err = handler.Serve(context.Background(), req, nethttp.NewResponse(httptest.NewRecorder()))
		var status servlet.StatusError
		if !errors.As(err, &status) || status.StatusCode() != http.StatusNotFound {
			t.Fatalf("%s err = %v, want 404 StatusError", target, err)
		}
	}
}

func TestDefaultServletRejectsUnsupportedMethod(t *testing.T) {
	t.Parallel()

	handler := newTestDefaultServlet(t)
	recorder := httptest.NewRecorder()
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodPost, "/assets/app.json", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	err = handler.Serve(context.Background(), req, nethttp.NewResponse(recorder))
	var status servlet.StatusError
	if !errors.As(err, &status) || status.StatusCode() != http.StatusMethodNotAllowed {
		t.Fatalf("POST err = %v, want 405 StatusError", err)
	}
	if recorder.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("Allow = %q, want GET, HEAD", recorder.Header().Get("Allow"))
	}
}

func newTestDefaultServlet(t *testing.T) *DefaultServlet {
	t.Helper()
	provider, err := NewFSProvider(fstest.MapFS{
		"assets":          &fstest.MapFile{Mode: 0o755 | fs.ModeDir},
		"assets/app.json": &fstest.MapFile{Data: []byte(`{"ok":true}`), ModTime: time.Date(2026, 8, 26, 10, 0, 0, 0, time.UTC)},
	})
	if err != nil {
		t.Fatalf("NewFSProvider failed: %v", err)
	}
	handler, err := NewDefaultServlet(provider)
	if err != nil {
		t.Fatalf("NewDefaultServlet failed: %v", err)
	}
	return handler
}
