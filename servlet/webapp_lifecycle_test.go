package servlet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestWebAppLifecycleAndRequestListeners(t *testing.T) {
	t.Parallel()

	var calls []string
	app, err := NewWebApp("orders",
		WithContextListener(ContextListenerFunc{
			Initialized: func(context.Context, ContextEvent) error {
				calls = append(calls, "context-init")
				return nil
			},
			Destroyed: func(context.Context, ContextEvent) error {
				calls = append(calls, "context-destroy")
				return nil
			},
		}),
		WithRequestListener(RequestListenerFunc{
			Initialized: func(context.Context, RequestEvent) error {
				calls = append(calls, "request-init")
				return nil
			},
			Destroyed: func(context.Context, RequestEvent) error {
				calls = append(calls, "request-destroy")
				return nil
			},
		}),
	)
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	if app.State() != WebAppStateNew {
		t.Fatalf("state = %v, want new", app.State())
	}
	if err := app.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	if err := app.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if err := app.RequestInitialized(context.Background(), req); err != nil {
		t.Fatalf("RequestInitialized failed: %v", err)
	}
	if err := app.RequestDestroyed(context.Background(), req, nil); err != nil {
		t.Fatalf("RequestDestroyed failed: %v", err)
	}
	if err := app.Stop(context.Background()); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if err := app.Destroy(context.Background()); err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	want := []string{"context-init", "request-init", "request-destroy", "context-destroy"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
	if app.State() != WebAppStateDestroyed {
		t.Fatalf("state = %v, want destroyed", app.State())
	}
}

func TestWebAppRejectsInvalidLifecycleTransition(t *testing.T) {
	t.Parallel()

	app, err := NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	err = app.Start(context.Background())
	if !errors.Is(err, ErrInvalidWebAppState) {
		t.Fatalf("Start err = %v, want ErrInvalidWebAppState", err)
	}
}
