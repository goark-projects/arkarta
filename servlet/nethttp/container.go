package nethttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
)

const (
	containerName    = "arkarta-nethttp"
	containerVersion = "0.1.0"
)

// Container 是基于标准库 net/http 的 Arkarta Servlet 参考容器。
type Container struct {
	metadata servletcontainer.Metadata

	mu           sync.RWMutex
	applications []servletcontainer.Application
	started      bool
	shutdown     bool
}

// NewContainer 创建 net/http 参考容器。
func NewContainer() *Container {
	return &Container{
		metadata: servletcontainer.NewMetadata(
			containerName,
			containerVersion,
			[]servletcontainer.Profile{servletcontainer.ProfileCore},
			map[string]string{"transport": "net/http"},
		),
	}
}

// Metadata 返回容器元数据。
func (c *Container) Metadata() servletcontainer.Metadata {
	return c.metadata
}

// Deploy 部署 Web 应用。
func (c *Container) Deploy(ctx context.Context, deployment *servletcontainer.Deployment) (servletcontainer.Application, error) {
	if err := c.ensureProfiles(deployment); err != nil {
		return nil, err
	}
	application, err := servletcontainer.NewApplication(ctx, deployment)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.shutdown {
		_ = application.Stop(ctx)
		return nil, http.ErrServerClosed
	}
	c.applications = append(c.applications, application)
	return application, nil
}

// Start 标记容器开始对外服务。
func (c *Container) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.shutdown {
		return http.ErrServerClosed
	}
	c.started = true
	return nil
}

// Shutdown 停止所有已部署应用。
func (c *Container) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	if c.shutdown {
		c.mu.Unlock()
		return nil
	}
	c.shutdown = true
	c.started = false
	applications := append([]servletcontainer.Application(nil), c.applications...)
	c.mu.Unlock()

	var result error
	for i := len(applications) - 1; i >= 0; i-- {
		result = errors.Join(result, applications[i].Stop(ctx))
	}
	return result
}

// Handler 返回容器聚合后的标准库 HTTP 处理器。
func (c *Container) Handler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		application, ok := c.matchApplication(request.URL.Path)
		if !ok {
			http.NotFound(writer, request)
			return
		}
		Handler(application.Handler()).ServeHTTP(writer, request)
	})
}

func (c *Container) ensureProfiles(deployment *servletcontainer.Deployment) error {
	if deployment == nil {
		return servletcontainer.ErrNilDeployment
	}
	for _, profile := range deployment.Profiles() {
		if !c.metadata.Supports(profile) {
			return ErrUnsupportedProfile
		}
	}
	return nil
}

func (c *Container) matchApplication(path string) (servletcontainer.Application, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.started || c.shutdown {
		return serviceUnavailableApplication{}, true
	}
	for _, application := range c.applications {
		contextPath := application.WebApp().ContextPath()
		if matchContextPath(path, contextPath) {
			return application, true
		}
	}
	return nil, false
}

func matchContextPath(path, contextPath string) bool {
	if contextPath == "" || contextPath == "/" {
		return true
	}
	return path == contextPath || strings.HasPrefix(path, contextPath+"/")
}

type serviceUnavailableApplication struct {
}

func (serviceUnavailableApplication) WebApp() *servlet.WebApp {
	return nil
}

func (serviceUnavailableApplication) Handler() servlet.Handler {
	return servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return servlet.NewHTTPError(http.StatusServiceUnavailable, http.StatusText(http.StatusServiceUnavailable), nil)
	})
}

func (serviceUnavailableApplication) Stop(context.Context) error {
	return nil
}
