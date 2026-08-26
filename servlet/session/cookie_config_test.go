package session

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/nethttp"
)

func TestCookieConfigCanBeBoundToWebApp(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	config, err := NewCookieConfig(
		WithCookieConfigName("ARKSESSION"),
		WithCookieConfigPath("/app"),
		WithCookieConfigSecure(true),
	)
	if err != nil {
		t.Fatalf("NewCookieConfig failed: %v", err)
	}
	if err := ConfigureCookie(app, config); err != nil {
		t.Fatalf("ConfigureCookie failed: %v", err)
	}
	got, ok := CookieConfigFor(app)
	if !ok || got.Name() != "ARKSESSION" || got.Path() != "/app" || !got.Secure() {
		t.Fatalf("cookie config = %#v/%v", got, ok)
	}

	accessor, err := NewAccessorForWebApp(NewMemoryManager(WithIDGenerator(sequenceID("s1"))), app)
	if err != nil {
		t.Fatalf("NewAccessorForWebApp failed: %v", err)
	}
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "http://example.com/app/orders", nil),
		servlet.WithRequestContextPath("/app"),
	)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	recorder := httptest.NewRecorder()
	if _, _, err := accessor.Get(context.Background(), req, nethttp.NewResponse(recorder), true); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	header := recorder.Header().Get("Set-Cookie")
	if !strings.Contains(header, "ARKSESSION=s1") || !strings.Contains(header, "Path=/app") || !strings.Contains(header, "Secure") {
		t.Fatalf("Set-Cookie = %q, want configured cookie", header)
	}
}

func TestConfigureCookieRejectsStartedWebApp(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	if err := app.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	if err := app.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	config, err := NewCookieConfig()
	if err != nil {
		t.Fatalf("NewCookieConfig failed: %v", err)
	}

	err = ConfigureCookie(app, config)
	if !errors.Is(err, ErrCookieConfigLocked) {
		t.Fatalf("ConfigureCookie err = %v, want ErrCookieConfigLocked", err)
	}
}
