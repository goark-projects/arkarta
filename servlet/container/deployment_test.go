package container

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
)

func TestDeploymentBuildsHandler(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := NewDeployment(app,
		WithMapping("/orders", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return nil
		}), servlet.FilterFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
			req.SetAttribute("filtered", true)
			return chain.Next(ctx, req, res)
		})),
	)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}

	handler, err := deployment.Handler()
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/orders", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if err := handler.Serve(context.Background(), req, nil); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}
	value, ok := req.Attribute("filtered")
	if !ok || value != true {
		t.Fatalf("filtered = %v/%v, want true/true", value, ok)
	}
}

func TestDeploymentRejectsDuplicateMappingAtBuild(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := NewDeployment(app,
		WithMapping("/orders", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return nil
		})),
		WithMapping("/orders", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return nil
		})),
	)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}

	_, err = deployment.Handler()
	if !errors.Is(err, servlet.ErrDuplicateMapping) {
		t.Fatalf("Handler err = %v, want ErrDuplicateMapping", err)
	}
}

func TestDeploymentRequiresWebApp(t *testing.T) {
	t.Parallel()

	_, err := NewDeployment(nil)
	if !errors.Is(err, ErrNilWebApp) {
		t.Fatalf("NewDeployment err = %v, want ErrNilWebApp", err)
	}
}
