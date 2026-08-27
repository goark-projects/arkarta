package nethttp

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
)

func TestServerServesDeployedApplicationAndStopsOnContextCancel(t *testing.T) {
	t.Parallel()

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			_, err := res.WriteString("ok")
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
	server, err := NewServer(container, WithReadHeaderTimeout(time.Second), WithIdleTimeout(time.Second))
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- server.Serve(ctx, listener)
	}()

	body := mustGetBody(t, "http://"+listener.Addr().String()+"/")
	if body != "ok" {
		t.Fatalf("body = %q, want ok", body)
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Serve err = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after context cancel")
	}
	if app.State() != servlet.WebAppStateDestroyed {
		t.Fatalf("state = %v, want destroyed", app.State())
	}
}

func TestServerRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	if _, err := NewServer(nil); !errors.Is(err, ErrNilContainer) {
		t.Fatalf("NewServer err = %v, want ErrNilContainer", err)
	}
	server, err := NewServer(NewContainer())
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	if err := server.Serve(context.Background(), nil); !errors.Is(err, ErrNilListener) {
		t.Fatalf("Serve err = %v, want ErrNilListener", err)
	}
}

func TestServerOptionsConfigureHTTPServer(t *testing.T) {
	t.Parallel()

	server, err := NewServer(NewContainer(),
		WithAddress("127.0.0.1:8081"),
		WithReadTimeout(time.Second),
		WithReadHeaderTimeout(2*time.Second),
		WithWriteTimeout(3*time.Second),
		WithIdleTimeout(4*time.Second),
		WithMaxHeaderBytes(8192),
	)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	httpServer := server.HTTPServer()
	if httpServer.Addr != "127.0.0.1:8081" ||
		httpServer.ReadTimeout != time.Second ||
		httpServer.ReadHeaderTimeout != 2*time.Second ||
		httpServer.WriteTimeout != 3*time.Second ||
		httpServer.IdleTimeout != 4*time.Second ||
		httpServer.MaxHeaderBytes != 8192 {
		t.Fatalf("http server config = %#v", httpServer)
	}
}

func mustGetBody(t *testing.T, url string) string {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%q", response.StatusCode, string(data))
	}
	return string(data)
}
