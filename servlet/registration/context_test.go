package registration_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/registration"
)

func TestRegistrationContextWrapsWebAppAndRegistry(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	ctx, err := registration.NewContext(app, nil)
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}
	orders, err := ctx.AddServlet("orders", servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
		_, err := res.WriteString("orders")
		return err
	}))
	if err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	if conflicts, err := orders.AddMapping("/orders"); err != nil || len(conflicts) != 0 {
		t.Fatalf("AddMapping = %#v/%v, want none/nil", conflicts, err)
	}
	snapshot, err := ctx.Freeze()
	if err != nil {
		t.Fatalf("Freeze failed: %v", err)
	}
	deployment, err := container.DeploymentFromRegistration(app, snapshot)
	if err != nil {
		t.Fatalf("DeploymentFromRegistration failed: %v", err)
	}
	handler, err := deployment.Handler()
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}
	dispatcher, err := ctx.NamedDispatcher("orders")
	if err != nil {
		t.Fatalf("NamedDispatcher failed: %v", err)
	}
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/source", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	res := newRegistryResponse()
	if err := dispatcher.Forward(context.Background(), req, res); err != nil {
		t.Fatalf("Forward failed: %v", err)
	}
	if res.body != "orders" {
		t.Fatalf("body = %q, want orders", res.body)
	}
	if err := handler.Serve(context.Background(), req, res); err == nil {
		t.Fatal("direct handler should not match restored /source path")
	}
}

func TestRegistrationContextRejectsNilWebApp(t *testing.T) {
	t.Parallel()

	if _, err := registration.NewContext(nil, nil); !errors.Is(err, registration.ErrNilWebApp) {
		t.Fatalf("NewContext err = %v, want ErrNilWebApp", err)
	}
}

func TestRegistrationContextRejectsMutationAfterStart(t *testing.T) {
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
	ctx, err := registration.NewContext(app, nil)
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}

	_, err = ctx.AddServlet("orders", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return nil
	}))
	if !errors.Is(err, registration.ErrRegistrationClosed) {
		t.Fatalf("AddServlet err = %v, want ErrRegistrationClosed", err)
	}
	if ok, err := ctx.SetInitParam("encoding", "utf-8"); ok || !errors.Is(err, registration.ErrRegistrationClosed) {
		t.Fatalf("SetInitParam ok/err = %v/%v, want false/ErrRegistrationClosed", ok, err)
	}
}

type registryResponse struct {
	header http.Header
	status int
	body   string
}

func newRegistryResponse() *registryResponse {
	return &registryResponse{header: make(http.Header), status: http.StatusOK}
}

func (r *registryResponse) Header() http.Header {
	return r.header
}

func (r *registryResponse) SetStatus(code int) {
	r.status = code
}

func (r *registryResponse) Status() int {
	return r.status
}

func (r *registryResponse) Write(data []byte) (int, error) {
	r.body += string(data)
	return len(data), nil
}

func (r *registryResponse) WriteString(value string) (int, error) {
	r.body += value
	return len(value), nil
}

func (r *registryResponse) Flush() error {
	return nil
}

func (r *registryResponse) Committed() bool {
	return false
}

func (r *registryResponse) Reset() error {
	r.body = ""
	return nil
}

func (r *registryResponse) BodyWriter() io.Writer {
	return io.Discard
}
