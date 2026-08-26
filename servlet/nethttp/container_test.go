package nethttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	servletcontainer "goark.dev/arkarta/servlet/container"

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
