package tck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
)

// ApplicationFactory 从部署描述创建容器应用。
type ApplicationFactory func(deployment *servletcontainer.Deployment) (servletcontainer.Application, error)

// RunLifecycle 执行 Servlet 生命周期兼容性测试。
func RunLifecycle(t *testing.T, factory ApplicationFactory) {
	t.Helper()
	t.Run("servlet_and_request_lifecycle_order", func(t *testing.T) {
		var calls []string
		app, err := servlet.NewWebApp("tck",
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
		target := &lifecycleServlet{calls: &calls}
		deployment, err := servletcontainer.NewDeployment(app, servletcontainer.WithServlet("/", "tckServlet", target))
		if err != nil {
			t.Fatalf("NewDeployment failed: %v", err)
		}
		application, err := factory(deployment)
		if err != nil {
			t.Fatalf("factory failed: %v", err)
		}
		req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/", nil))
		if err != nil {
			t.Fatalf("NewRequest failed: %v", err)
		}
		if err := application.Handler().Serve(context.Background(), req, nil); err != nil {
			t.Fatalf("Serve failed: %v", err)
		}
		if err := application.Stop(context.Background()); err != nil {
			t.Fatalf("Stop failed: %v", err)
		}

		want := []string{"init:tckServlet", "request-init", "serve", "request-destroy", "destroy"}
		if !reflect.DeepEqual(calls, want) {
			t.Fatalf("calls = %#v, want %#v", calls, want)
		}
	})
	t.Run("filter_lifecycle_order", func(t *testing.T) {
		var calls []string
		app, err := servlet.NewWebApp("tck")
		if err != nil {
			t.Fatalf("NewWebApp failed: %v", err)
		}
		target := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			calls = append(calls, "handler")
			return nil
		})
		deployment, err := servletcontainer.NewDeployment(app, servletcontainer.WithMapping("/", target, &lifecycleFilter{calls: &calls}))
		if err != nil {
			t.Fatalf("NewDeployment failed: %v", err)
		}
		application, err := factory(deployment)
		if err != nil {
			t.Fatalf("factory failed: %v", err)
		}
		req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/", nil))
		if err != nil {
			t.Fatalf("NewRequest failed: %v", err)
		}
		if err := application.Handler().Serve(context.Background(), req, nil); err != nil {
			t.Fatalf("Serve failed: %v", err)
		}
		if err := application.Stop(context.Background()); err != nil {
			t.Fatalf("Stop failed: %v", err)
		}

		want := []string{"filter-init:/#filter0", "filter", "handler", "filter-destroy"}
		if !reflect.DeepEqual(calls, want) {
			t.Fatalf("calls = %#v, want %#v", calls, want)
		}
	})
}

type lifecycleServlet struct {
	calls *[]string
}

func (s *lifecycleServlet) Init(_ context.Context, cfg servlet.ServletConfig) error {
	*s.calls = append(*s.calls, "init:"+cfg.Name())
	return nil
}

func (s *lifecycleServlet) Serve(context.Context, *servlet.Request, servlet.Response) error {
	*s.calls = append(*s.calls, "serve")
	return nil
}

func (s *lifecycleServlet) Destroy(context.Context) error {
	*s.calls = append(*s.calls, "destroy")
	return nil
}

type lifecycleFilter struct {
	calls *[]string
}

func (f *lifecycleFilter) Init(_ context.Context, cfg servlet.FilterConfig) error {
	*f.calls = append(*f.calls, "filter-init:"+cfg.Name())
	return nil
}

func (f *lifecycleFilter) Filter(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
	*f.calls = append(*f.calls, "filter")
	return chain.Next(ctx, req, res)
}

func (f *lifecycleFilter) Destroy(context.Context) error {
	*f.calls = append(*f.calls, "filter-destroy")
	return nil
}
