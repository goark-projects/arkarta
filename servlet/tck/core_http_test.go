package tck_test

import (
	"net/http"
	"testing"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/nethttp"
	"goark.dev/arkarta/servlet/session"
	"goark.dev/arkarta/servlet/tck"
)

func TestRunCoreHTTPWithNetHTTPAdapter(t *testing.T) {
	tck.RunCoreHTTP(t, func(handler servlet.Handler) http.Handler {
		return nethttp.Handler(handler)
	})
}

func TestRunSessionManagerWithMemoryManager(t *testing.T) {
	tck.RunSessionManager(t, func() session.Manager {
		return session.NewMemoryManager()
	})
}

func TestRunLifecycleWithManagedApplication(t *testing.T) {
	tck.RunLifecycle(t, func(deployment *servletcontainer.Deployment) (servletcontainer.Application, error) {
		return servletcontainer.NewApplication(t.Context(), deployment)
	})
}

func TestRunDispatcher(t *testing.T) {
	tck.RunDispatcher(t)
}

func TestRunErrorPagesWithNetHTTPAdapter(t *testing.T) {
	tck.RunErrorPages(t, func(handler servlet.Handler, registry *servlet.ErrorPageRegistry) http.Handler {
		return nethttp.HandlerWithOptions(handler, nethttp.WithErrorPages(registry))
	})
}

func TestRunMultipartParser(t *testing.T) {
	tck.RunMultipartParser(t, func(options ...multipart.Option) *multipart.Parser {
		return multipart.NewParser(options...)
	})
}

func TestRunHTTPContainerWithNetHTTPContainer(t *testing.T) {
	tck.RunHTTPContainer(t, func() tck.HTTPContainer {
		return nethttp.NewContainer()
	})
}
