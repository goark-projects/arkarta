package container

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/registration"
	"goark.dev/arkarta/servlet/security"
)

func TestDeploymentAppliesRegisteredSecurityConstraint(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	orders, err := registry.AddServlet(
		"orders",
		servlet.HandlerFunc(
			func(context.Context, *servlet.Request, servlet.Response) error {
				return nil
			},
		),
	)
	if err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	if conflicts, err := orders.AddMapping("/orders"); err != nil ||
		len(conflicts) != 0 {
		t.Fatalf("AddMapping conflicts/err = %#v/%v, want none/nil", conflicts, err)
	}
	constraint := security.NewConstraint(security.WithRoles("admin"))
	if err := orders.SetSecurityConfig(constraint); err != nil {
		t.Fatalf("SetSecurityConfig failed: %v", err)
	}
	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	snapshot, err := registry.Freeze()
	if err != nil {
		t.Fatalf("Freeze failed: %v", err)
	}
	deployment, err := DeploymentFromRegistration(app, snapshot)
	if err != nil {
		t.Fatalf("DeploymentFromRegistration failed: %v", err)
	}
	handler, err := deployment.Handler()
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	denied, err := servlet.NewRequest(
		httptest.NewRequest(http.MethodGet, "https://example.com/orders", nil),
	)
	if err != nil {
		t.Fatalf("NewRequest denied failed: %v", err)
	}
	if err := handler.Serve(context.Background(), denied, nil); !statusCodeIs(
		err,
		http.StatusUnauthorized,
	) {
		t.Fatalf("unauthenticated err = %v, want 401", err)
	}

	allowed, err := servlet.NewRequest(
		httptest.NewRequest(http.MethodGet, "https://example.com/orders", nil),
	)
	if err != nil {
		t.Fatalf("NewRequest allowed failed: %v", err)
	}
	security.SetPrincipal(
		allowed,
		security.PrincipalFunc(func() string { return "alice" }),
		security.AuthTypeBasic,
		"admin",
	)
	if err := handler.Serve(context.Background(), allowed, nil); err != nil {
		t.Fatalf("authenticated Serve failed: %v", err)
	}
}

func TestDeploymentAppliesRunAsDuringServletService(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	orders, err := registry.AddServlet(
		"orders",
		servlet.HandlerFunc(
			func(_ context.Context, req *servlet.Request, _ servlet.Response) error {
				if role := security.RunAsRole(req); role != "system" {
					t.Fatalf("run-as role = %q, want system", role)
				}
				if !security.UserInRole(req, "system") {
					t.Fatal(
						"run-as role should participate in UserInRole during servlet execution",
					)
				}
				return nil
			},
		),
	)
	if err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	if conflicts, err := orders.AddMapping("/orders"); err != nil ||
		len(conflicts) != 0 {
		t.Fatalf("AddMapping conflicts/err = %#v/%v, want none/nil", conflicts, err)
	}
	if err := orders.SetRunAsRole("system"); err != nil {
		t.Fatalf("SetRunAsRole failed: %v", err)
	}
	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	snapshot, err := registry.Freeze()
	if err != nil {
		t.Fatalf("Freeze failed: %v", err)
	}
	deployment, err := DeploymentFromRegistration(app, snapshot)
	if err != nil {
		t.Fatalf("DeploymentFromRegistration failed: %v", err)
	}
	handler, err := deployment.Handler()
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}
	req, err := servlet.NewRequest(
		httptest.NewRequest(http.MethodGet, "https://example.com/orders", nil),
	)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	if err := handler.Serve(context.Background(), req, nil); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}
	if role := security.RunAsRole(req); role != "" {
		t.Fatalf("run-as role after serve = %q, want empty", role)
	}
}

func statusCodeIs(err error, status int) bool {
	var statusErr servlet.StatusError
	return errors.As(err, &statusErr) && statusErr.StatusCode() == status
}
