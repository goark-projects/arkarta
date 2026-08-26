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

func TestDeploymentSetsServletNameAttribute(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := NewDeployment(app,
		WithServlet("/orders", "ordersServlet", servletNameServlet{t: t}),
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
	if value, ok := req.Attribute(servlet.AttributeServletName); ok || value != nil {
		t.Fatalf("servlet name should be restored, got %v/%v", value, ok)
	}
}

type servletNameServlet struct {
	t *testing.T
}

func (s servletNameServlet) Init(context.Context, servlet.ServletConfig) error {
	return nil
}

func (s servletNameServlet) Serve(_ context.Context, req *servlet.Request, _ servlet.Response) error {
	value, ok := req.Attribute(servlet.AttributeServletName)
	if !ok || value != "ordersServlet" {
		s.t.Fatalf("servlet name = %v/%v, want ordersServlet/true", value, ok)
	}
	return nil
}

func (s servletNameServlet) Destroy(context.Context) error {
	return nil
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
