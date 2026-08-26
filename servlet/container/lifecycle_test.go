package container

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"goark.dev/arkarta/servlet"
)

func TestManagedApplicationInitializesServletAndRequestEvents(t *testing.T) {
	t.Parallel()

	var calls []string
	app, err := servlet.NewWebApp("orders",
		servlet.WithRequestListener(servlet.RequestListenerFunc{
			Initialized: func(context.Context, servlet.RequestEvent) error {
				calls = append(calls, "request-init")
				return nil
			},
			Destroyed: func(context.Context, servlet.RequestEvent) error {
				calls = append(calls, "request-destroy")
				return nil
			},
		}),
	)
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	target := &recordingServlet{calls: &calls}
	deployment, err := NewDeployment(app, WithServlet("/orders", "ordersServlet", target))
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}

	application, err := NewApplication(context.Background(), deployment)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/orders", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if err := application.Handler().Serve(context.Background(), req, nil); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}
	if err := application.Stop(context.Background()); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	want := []string{"init:ordersServlet", "request-init", "serve", "request-destroy", "destroy"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
	if app.State() != servlet.WebAppStateDestroyed {
		t.Fatalf("state = %v, want destroyed", app.State())
	}
}

type recordingServlet struct {
	calls *[]string
}

func (s *recordingServlet) Init(_ context.Context, cfg servlet.ServletConfig) error {
	*s.calls = append(*s.calls, "init:"+cfg.Name())
	return nil
}

func (s *recordingServlet) Serve(context.Context, *servlet.Request, servlet.Response) error {
	*s.calls = append(*s.calls, "serve")
	return nil
}

func (s *recordingServlet) Destroy(context.Context) error {
	*s.calls = append(*s.calls, "destroy")
	return nil
}
