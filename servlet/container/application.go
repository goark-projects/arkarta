package container

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"goark.dev/arkarta/servlet"
)

// ErrNilDeployment 表示应用构建时缺少部署描述。
var ErrNilDeployment = errors.New("arkarta/servlet/container: deployment is nil")

// ManagedApplication 是由部署描述构建的标准 Servlet 应用。
type ManagedApplication struct {
	webApp   *servlet.WebApp
	handler  servlet.Handler
	servlets []servlet.Servlet
	filters  []servlet.ManagedFilter

	mu      sync.Mutex
	stopped bool
}

// NewApplication 初始化并启动部署描述对应的标准应用。
func NewApplication(ctx context.Context, deployment *Deployment) (*ManagedApplication, error) {
	if deployment == nil {
		return nil, ErrNilDeployment
	}
	handler, err := deployment.Handler()
	if err != nil {
		return nil, err
	}
	app := &ManagedApplication{
		webApp:  deployment.WebApp(),
		handler: handler,
	}
	if err := app.initialize(ctx, deployment); err != nil {
		return nil, err
	}
	return app, nil
}

// WebApp 返回应用上下文。
func (a *ManagedApplication) WebApp() *servlet.WebApp {
	return a.webApp
}

// Handler 返回带生命周期保护和请求事件的处理器。
func (a *ManagedApplication) Handler() servlet.Handler {
	return servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		if a.webApp.State() != servlet.WebAppStateStarted {
			return servlet.NewHTTPError(http.StatusServiceUnavailable, http.StatusText(http.StatusServiceUnavailable), nil)
		}
		if err := a.webApp.RequestInitialized(ctx, req); err != nil {
			return err
		}
		err := a.handler.Serve(ctx, req, res)
		return errors.Join(err, a.webApp.RequestDestroyed(ctx, req, err))
	})
}

// Stop 停止应用并销毁 Servlet 与上下文。
func (a *ManagedApplication) Stop(ctx context.Context) error {
	a.mu.Lock()
	if a.stopped {
		a.mu.Unlock()
		return nil
	}
	a.stopped = true
	a.mu.Unlock()

	var result error
	result = errors.Join(result, a.webApp.Stop(ctx))
	for i := len(a.servlets) - 1; i >= 0; i-- {
		result = errors.Join(result, a.servlets[i].Destroy(ctx))
	}
	for i := len(a.filters) - 1; i >= 0; i-- {
		result = errors.Join(result, a.filters[i].Destroy(ctx))
	}
	result = errors.Join(result, a.webApp.Destroy(ctx))
	return result
}

func (a *ManagedApplication) initialize(ctx context.Context, deployment *Deployment) error {
	if err := a.webApp.Initialize(ctx); err != nil {
		return err
	}
	for _, item := range deployment.filterInitializations() {
		if err := item.filter.Init(ctx, servlet.NewFilterConfig(item.name, a.webApp, item.initParam)); err != nil {
			_ = a.destroyInitialized(ctx)
			return err
		}
		a.filters = append(a.filters, item.filter)
	}
	for _, mapping := range deployment.servletMappings() {
		target := mapping.Handler().(servlet.Servlet)
		if err := target.Init(ctx, mapping.servletConfig(a.webApp)); err != nil {
			_ = a.destroyInitialized(ctx)
			return err
		}
		a.servlets = append(a.servlets, target)
	}
	if err := a.webApp.Start(ctx); err != nil {
		_ = a.destroyInitialized(ctx)
		return err
	}
	return nil
}

func (a *ManagedApplication) destroyInitialized(ctx context.Context) error {
	var result error
	for i := len(a.servlets) - 1; i >= 0; i-- {
		result = errors.Join(result, a.servlets[i].Destroy(ctx))
	}
	for i := len(a.filters) - 1; i >= 0; i-- {
		result = errors.Join(result, a.filters[i].Destroy(ctx))
	}
	result = errors.Join(result, a.webApp.Destroy(ctx))
	return result
}
