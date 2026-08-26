package resource

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"testing/fstest"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/nethttp"
)

func TestDefaultServletServesWelcomeFiles(t *testing.T) {
	t.Parallel()

	handler := newWelcomeServlet(t)
	tests := map[string]string{
		"/":      "root",
		"/docs":  "docs",
		"/docs/": "docs",
	}
	for target, want := range tests {
		recorder := httptest.NewRecorder()
		req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, target, nil))
		if err != nil {
			t.Fatalf("NewRequest %s failed: %v", target, err)
		}
		if err := handler.Serve(context.Background(), req, nethttp.NewResponse(recorder)); err != nil {
			t.Fatalf("Serve %s failed: %v", target, err)
		}
		if recorder.Code != http.StatusOK || recorder.Body.String() != want {
			t.Fatalf("%s status/body = %d/%q, want 200/%q", target, recorder.Code, recorder.Body.String(), want)
		}
	}
}

func TestDefaultServletCustomWelcomeFiles(t *testing.T) {
	t.Parallel()

	provider := newWelcomeProvider(t)
	defaultHandler, err := NewDefaultServlet(provider, WithWelcomeFiles())
	if err != nil {
		t.Fatalf("NewDefaultServlet default failed: %v", err)
	}
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/custom", nil))
	if err != nil {
		t.Fatalf("NewRequest default failed: %v", err)
	}
	err = defaultHandler.Serve(context.Background(), req, nethttp.NewResponse(httptest.NewRecorder()))
	var status servlet.StatusError
	if !errors.As(err, &status) || status.StatusCode() != http.StatusNotFound {
		t.Fatalf("disabled welcome err = %v, want 404", err)
	}

	customHandler, err := NewDefaultServlet(provider, WithWelcomeFiles("home.html"))
	if err != nil {
		t.Fatalf("NewDefaultServlet custom failed: %v", err)
	}
	recorder := httptest.NewRecorder()
	req, err = servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/custom", nil))
	if err != nil {
		t.Fatalf("NewRequest custom failed: %v", err)
	}
	if err := customHandler.Serve(context.Background(), req, nethttp.NewResponse(recorder)); err != nil {
		t.Fatalf("Serve custom failed: %v", err)
	}
	if recorder.Body.String() != "custom" {
		t.Fatalf("custom body = %q, want custom", recorder.Body.String())
	}
}

func TestDefaultServletDirectoryWithoutWelcomeIsNotFound(t *testing.T) {
	t.Parallel()

	handler := newWelcomeServlet(t)
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/missing", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	err = handler.Serve(context.Background(), req, nethttp.NewResponse(httptest.NewRecorder()))
	var status servlet.StatusError
	if !errors.As(err, &status) || status.StatusCode() != http.StatusNotFound {
		t.Fatalf("missing welcome err = %v, want 404", err)
	}
}

func TestDefaultServletRejectsInvalidWelcomeFile(t *testing.T) {
	t.Parallel()

	_, err := NewDefaultServlet(newWelcomeProvider(t), WithWelcomeFiles("../index.html"))
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("invalid welcome err = %v, want ErrInvalidPath", err)
	}
}

func TestDefaultServletWelcomeFilesAreIsolated(t *testing.T) {
	t.Parallel()

	handler := newWelcomeServlet(t)
	files := handler.WelcomeFiles()
	files[0] = "changed.html"
	want := []string{"index.html", "index.htm"}
	if !reflect.DeepEqual(handler.WelcomeFiles(), want) {
		t.Fatalf("welcome files = %#v, want %#v", handler.WelcomeFiles(), want)
	}
}

func newWelcomeServlet(t *testing.T) *DefaultServlet {
	t.Helper()
	handler, err := NewDefaultServlet(newWelcomeProvider(t))
	if err != nil {
		t.Fatalf("NewDefaultServlet failed: %v", err)
	}
	return handler
}

func newWelcomeProvider(t *testing.T) *FSProvider {
	t.Helper()
	provider, err := NewFSProvider(fstest.MapFS{
		"index.html":        &fstest.MapFile{Data: []byte("root")},
		"docs":              &fstest.MapFile{Mode: 0o755 | fs.ModeDir},
		"docs/index.html":   &fstest.MapFile{Data: []byte("docs")},
		"custom":            &fstest.MapFile{Mode: 0o755 | fs.ModeDir},
		"custom/home.html":  &fstest.MapFile{Data: []byte("custom")},
		"missing":           &fstest.MapFile{Mode: 0o755 | fs.ModeDir},
		"missing/readme.md": &fstest.MapFile{Data: []byte("readme")},
	})
	if err != nil {
		t.Fatalf("NewFSProvider failed: %v", err)
	}
	return provider
}
