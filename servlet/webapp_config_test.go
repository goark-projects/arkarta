package servlet

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestWebAppDefaultCapabilities(t *testing.T) {
	t.Parallel()

	app, err := NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	if app.VirtualServerName() != DefaultVirtualServerName {
		t.Fatalf("virtual server = %q, want default", app.VirtualServerName())
	}
	if app.RequestCharacterEncoding() != DefaultCharacterEncoding || app.ResponseCharacterEncoding() != DefaultCharacterEncoding {
		t.Fatalf("charsets = %q/%q, want defaults", app.RequestCharacterEncoding(), app.ResponseCharacterEncoding())
	}
	if app.SessionTimeout() != DefaultSessionTimeout {
		t.Fatalf("session timeout = %s, want %s", app.SessionTimeout(), DefaultSessionTimeout)
	}
	if app.EffectiveMajorVersion() != ServletSpecMajorVersion || app.EffectiveMinorVersion() != ServletSpecMinorVersion {
		t.Fatalf("effective version = %d.%d, want %d.%d",
			app.EffectiveMajorVersion(),
			app.EffectiveMinorVersion(),
			ServletSpecMajorVersion,
			ServletSpecMinorVersion,
		)
	}
	if app.ArkartaMajorVersion() != ArkartaServletMajorVersion || app.ArkartaMinorVersion() != ArkartaServletMinorVersion {
		t.Fatalf("arkarta version = %d.%d, want %d.%d",
			app.ArkartaMajorVersion(),
			app.ArkartaMinorVersion(),
			ArkartaServletMajorVersion,
			ArkartaServletMinorVersion,
		)
	}
}

func TestWebAppCustomCapabilities(t *testing.T) {
	t.Parallel()

	app, err := NewWebApp("orders",
		WithVirtualServerName("api.internal"),
		WithRequestCharacterEncoding("gb18030"),
		WithResponseCharacterEncoding("utf-16"),
		WithSessionTimeout(45*time.Minute),
		WithMimeType("foo", "application/x-foo"),
		WithResourceFS(fstest.MapFS{"static/app.txt": &fstest.MapFile{Data: []byte("ok")}}),
	)
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	if app.VirtualServerName() != "api.internal" {
		t.Fatalf("virtual server = %q, want api.internal", app.VirtualServerName())
	}
	if app.RequestCharacterEncoding() != "gb18030" || app.ResponseCharacterEncoding() != "utf-16" {
		t.Fatalf("charsets = %q/%q, want custom", app.RequestCharacterEncoding(), app.ResponseCharacterEncoding())
	}
	if app.SessionTimeout() != 45*time.Minute {
		t.Fatalf("session timeout = %s, want 45m", app.SessionTimeout())
	}
	if app.MimeType("a.foo") != "application/x-foo" {
		t.Fatalf("custom mime = %q, want application/x-foo", app.MimeType("a.foo"))
	}
	file, err := app.OpenResource(context.Background(), "/static/app.txt")
	if err != nil {
		t.Fatalf("OpenResource failed: %v", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(data) != "ok" {
		t.Fatalf("resource body = %q, want ok", string(data))
	}
}

func TestWebAppMimeMappingsAreIsolated(t *testing.T) {
	t.Parallel()

	app, err := NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	mappings := app.MimeMappings()
	mappings[".json"] = "text/plain"

	want := map[string]string{
		".json": "application/json",
		".html": "text/html; charset=utf-8",
	}
	got := map[string]string{
		".json": app.MimeType("a.json"),
		".html": app.MimeType("a.html"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mime mappings = %#v, want %#v", got, want)
	}
}

func TestWebAppResourceLookup(t *testing.T) {
	t.Parallel()

	app, err := NewWebApp("orders", WithResourceFS(fstest.MapFS{
		"static/app.txt": &fstest.MapFile{Data: []byte("ok")},
	}))
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	exists, err := app.ResourceExists(context.Background(), "/static/app.txt")
	if err != nil || !exists {
		t.Fatalf("ResourceExists existing = %v/%v, want true/nil", exists, err)
	}
	exists, err = app.ResourceExists(context.Background(), "/missing.txt")
	if err != nil || exists {
		t.Fatalf("ResourceExists missing = %v/%v, want false/nil", exists, err)
	}
	if _, err := app.OpenResource(context.Background(), "/../secret.txt"); !errors.Is(err, ErrInvalidWebAppConfig) {
		t.Fatalf("unsafe resource err = %v, want ErrInvalidWebAppConfig", err)
	}
	empty, err := NewWebApp("empty")
	if err != nil {
		t.Fatalf("NewWebApp empty failed: %v", err)
	}
	if _, err := empty.OpenResource(context.Background(), "/missing.txt"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("empty resource err = %v, want fs.ErrNotExist", err)
	}
}

func TestWebAppResourcePathsAndTempDir(t *testing.T) {
	t.Parallel()

	app, err := NewWebApp("orders",
		WithTempDir("/tmp/arkarta"),
		WithResourceFS(fstest.MapFS{
			"static/app.txt":      &fstest.MapFile{Data: []byte("ok")},
			"static/css/site.css": &fstest.MapFile{Data: []byte("body{}")},
		}),
	)
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	paths, err := app.ResourcePaths(context.Background(), "/static")
	if err != nil {
		t.Fatalf("ResourcePaths failed: %v", err)
	}
	want := []string{"/static/app.txt", "/static/css/"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %#v, want %#v", paths, want)
	}
	if app.TempDir() != "/tmp/arkarta" {
		t.Fatalf("temp dir = %q, want /tmp/arkarta", app.TempDir())
	}
	if value, ok := app.Attribute(AttributeTempDir); !ok || value != "/tmp/arkarta" {
		t.Fatalf("temp attr = %v/%v, want /tmp/arkarta/true", value, ok)
	}
}

func TestWebAppDispatcherProvider(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	if err := router.Handle("/target", HandlerFunc(func(_ context.Context, _ *Request, res Response) error {
		_, err := res.WriteString("target")
		return err
	})); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	registry := NewDispatcherRegistry(router)
	registry.RegisterName("targetServlet", "/target")
	app, err := NewWebApp("orders", WithDispatcherProvider(registry))
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	dispatcher, err := app.NamedDispatcher("targetServlet")
	if err != nil {
		t.Fatalf("NamedDispatcher failed: %v", err)
	}
	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/source", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	res := newTestResponse()
	if err := dispatcher.Forward(context.Background(), req, res); err != nil {
		t.Fatalf("Forward failed: %v", err)
	}
	if res.body.String() != "target" {
		t.Fatalf("body = %q, want target", res.body.String())
	}
}

func TestWebAppLogUsesConfiguredLogger(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{Level: slog.LevelInfo}))
	app, err := NewWebApp("orders", WithLogger(logger))
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	app.Log(context.Background(), "started", "app", app.Name())
	if !strings.Contains(output.String(), "started") || !strings.Contains(output.String(), "orders") {
		t.Fatalf("log output = %q, want message and app", output.String())
	}
}

func TestWebAppRejectsInvalidCapabilities(t *testing.T) {
	t.Parallel()

	tests := []WebAppOption{
		WithVirtualServerName(" "),
		WithRequestCharacterEncoding(""),
		WithResponseCharacterEncoding(""),
		WithSessionTimeout(-time.Second),
		WithMimeType("/", "text/plain"),
		WithMimeType("txt", "bad content type"),
		WithResourceFS(nil),
		WithLogger(nil),
	}
	for _, option := range tests {
		if _, err := NewWebApp("orders", option); !errors.Is(err, ErrInvalidWebAppConfig) {
			t.Fatalf("NewWebApp err = %v, want ErrInvalidWebAppConfig", err)
		}
	}
}
