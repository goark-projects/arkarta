package nethttp

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/nativeio"

	"goark.dev/arkarta/servlet"
)

func TestContainerDeploysApplicationAndServesHTTP(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithMapping("/orders", servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			_, err := res.WriteString("orders")
			return err
		})),
	)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}
	container := NewContainer()
	application, err := container.Deploy(context.Background(), deployment)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if err := container.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	container.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/orders", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Body.String() != "orders" {
		t.Fatalf("body = %q, want orders", recorder.Body.String())
	}
	if application.WebApp().State() != servlet.WebAppStateStarted {
		t.Fatalf("state = %v, want started", application.WebApp().State())
	}
}

func TestContainerServesApplicationRelativePath(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders", servlet.WithContextPath("/orders"))
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithMapping("/items", servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
			if req.ContextPath() != "/orders" {
				t.Fatalf("context path = %q, want /orders", req.ContextPath())
			}
			if req.Path() != "/items" {
				t.Fatalf("path = %q, want /items", req.Path())
			}
			_, err := res.WriteString(req.ServletPath())
			return err
		})),
	)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}
	container := NewContainer()
	if _, err := container.Deploy(context.Background(), deployment); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if err := container.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	container.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/orders/items", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Body.String() != "/items" {
		t.Fatalf("body = %q, want /items", recorder.Body.String())
	}
}

func TestContainerMetadataUsesReleaseVersion(t *testing.T) {
	t.Parallel()

	metadata := NewContainer().Metadata()
	if metadata.Name() != "arkarta-nethttp" {
		t.Fatalf("name = %q, want arkarta-nethttp", metadata.Name())
	}
	if metadata.Version() != "0.0.1" {
		t.Fatalf("version = %q, want 0.0.1", metadata.Version())
	}
	if !metadata.Supports(servletcontainer.ProfileCore) {
		t.Fatal("metadata must support core profile")
	}
	if !metadata.Supports(servletcontainer.ProfileNativeIO) {
		t.Fatal("metadata must support native I/O profile")
	}
}

func TestContainerRejectsUnsupportedProfile(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithProfile(servletcontainer.ProfileUpgrade),
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return nil
		})),
	)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}

	_, err = NewContainer().Deploy(context.Background(), deployment)
	if !errors.Is(err, ErrUnsupportedProfile) {
		t.Fatalf("Deploy err = %v, want ErrUnsupportedProfile", err)
	}
}

func TestContainerExposesNativeSender(t *testing.T) {
	t.Parallel()

	sender := NewContainer().NativeSender()
	region, err := nativeio.NewFileRegion(strings.NewReader("abcdef"), 1, 3)
	if err != nil {
		t.Fatalf("NewFileRegion failed: %v", err)
	}
	var dst bytes.Buffer
	result, err := sender.SendFile(context.Background(), &dst, region)
	if err != nil {
		t.Fatalf("SendFile failed: %v", err)
	}
	if dst.String() != "bcd" || result.Bytes() != 3 {
		t.Fatalf("native sender body/result = %q/%d, want bcd/3", dst.String(), result.Bytes())
	}
}

func TestContainerShutdownStopsApplications(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return nil
		})),
	)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}
	container := NewContainer()
	if _, err := container.Deploy(context.Background(), deployment); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	if err := container.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
	if app.State() != servlet.WebAppStateDestroyed {
		t.Fatalf("state = %v, want destroyed", app.State())
	}
}
