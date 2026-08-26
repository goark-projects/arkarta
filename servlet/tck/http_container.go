package tck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
)

// HTTPContainer 表示同时实现 Servlet Container 和 net/http 入口的容器。
type HTTPContainer interface {
	servletcontainer.Container
	Handler() http.Handler
}

// HTTPContainerFactory 创建待验证 HTTP 容器。
type HTTPContainerFactory func() HTTPContainer

// RunHTTPContainer 执行 HTTP 容器兼容性测试。
func RunHTTPContainer(t *testing.T, factory HTTPContainerFactory) {
	t.Helper()
	t.Run("deploy_start_serve_shutdown", func(t *testing.T) {
		target := factory()
		if !target.Metadata().Supports(servletcontainer.ProfileCore) {
			t.Fatal("container must support core profile")
		}
		app, err := servlet.NewWebApp("tck")
		if err != nil {
			t.Fatalf("NewWebApp failed: %v", err)
		}
		deployment, err := servletcontainer.NewDeployment(app,
			servletcontainer.WithMapping("/", servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
				_, err := res.WriteString("container")
				return err
			})),
		)
		if err != nil {
			t.Fatalf("NewDeployment failed: %v", err)
		}
		application, err := target.Deploy(context.Background(), deployment)
		if err != nil {
			t.Fatalf("Deploy failed: %v", err)
		}
		if err := target.Start(context.Background()); err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		recorder := httptest.NewRecorder()
		target.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		if recorder.Code != http.StatusOK || recorder.Body.String() != "container" {
			t.Fatalf("status/body = %d/%q", recorder.Code, recorder.Body.String())
		}
		if err := target.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown failed: %v", err)
		}
		if application.WebApp().State() != servlet.WebAppStateDestroyed {
			t.Fatalf("state = %v, want destroyed", application.WebApp().State())
		}
	})
}
