package container

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/registration"
)

func TestDeploymentFromRegistrationBuildsServletAndFilterMappings(t *testing.T) {
	t.Parallel()

	var calls []string
	registry := registration.NewRegistry()
	orders, err := registry.AddServlet("orders", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		calls = append(calls, "orders")
		return nil
	}))
	if err != nil {
		t.Fatalf("AddServlet orders failed: %v", err)
	}
	if conflicts, err := orders.AddMapping("/orders/*"); err != nil || len(conflicts) != 0 {
		t.Fatalf("AddMapping orders conflicts/err = %#v/%v, want none/nil", conflicts, err)
	}
	other, err := registry.AddServlet("other", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		calls = append(calls, "other")
		return nil
	}))
	if err != nil {
		t.Fatalf("AddServlet other failed: %v", err)
	}
	if conflicts, err := other.AddMapping("/other/*"); err != nil || len(conflicts) != 0 {
		t.Fatalf("AddMapping other conflicts/err = %#v/%v, want none/nil", conflicts, err)
	}
	audit, err := registry.AddFilter("audit", recordContainerFilter("audit", &calls))
	if err != nil {
		t.Fatalf("AddFilter audit failed: %v", err)
	}
	if err := audit.AddMappingForURLPatterns(0, false, "/orders/*"); err != nil {
		t.Fatalf("AddMappingForURLPatterns failed: %v", err)
	}
	named, err := registry.AddFilter("named", recordContainerFilter("named", &calls))
	if err != nil {
		t.Fatalf("AddFilter named failed: %v", err)
	}
	if err := named.AddMappingForServletNames(0, true, "other"); err != nil {
		t.Fatalf("AddMappingForServletNames failed: %v", err)
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
	ordersReq, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/orders/42", nil))
	if err != nil {
		t.Fatalf("NewRequest orders failed: %v", err)
	}
	if err := handler.Serve(context.Background(), ordersReq, nil); err != nil {
		t.Fatalf("Serve orders failed: %v", err)
	}
	otherReq, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/other/42", nil))
	if err != nil {
		t.Fatalf("NewRequest other failed: %v", err)
	}
	if err := handler.Serve(context.Background(), otherReq, nil); err != nil {
		t.Fatalf("Serve other failed: %v", err)
	}

	want := []string{"audit", "orders", "named", "other"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestDeploymentFromRegistrationInitializesConfigsAndListeners(t *testing.T) {
	t.Parallel()

	var calls []string
	registry := registration.NewRegistry()
	if _, err := registry.AddContextListener(servlet.ContextListenerFunc{
		Initialized: func(context.Context, servlet.ContextEvent) error {
			calls = append(calls, "context-init")
			return nil
		},
		Destroyed: func(context.Context, servlet.ContextEvent) error {
			calls = append(calls, "context-destroy")
			return nil
		},
	}); err != nil {
		t.Fatalf("AddContextListener failed: %v", err)
	}
	if _, err := registry.AddRequestListener(servlet.RequestListenerFunc{
		Initialized: func(context.Context, servlet.RequestEvent) error {
			calls = append(calls, "request-init")
			return nil
		},
		Destroyed: func(context.Context, servlet.RequestEvent) error {
			calls = append(calls, "request-destroy")
			return nil
		},
	}); err != nil {
		t.Fatalf("AddRequestListener failed: %v", err)
	}
	target := &configRecordingServlet{calls: &calls}
	orders, err := registry.AddServlet("orders", target)
	if err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	if ok, err := orders.SetInitParam("encoding", "utf-8"); err != nil || !ok {
		t.Fatalf("SetInitParam servlet ok/err = %v/%v, want true/nil", ok, err)
	}
	if conflicts, err := orders.AddMapping("/orders"); err != nil || len(conflicts) != 0 {
		t.Fatalf("AddMapping conflicts/err = %#v/%v, want none/nil", conflicts, err)
	}
	filter := &configRecordingFilter{calls: &calls}
	audit, err := registry.AddFilter("audit", filter)
	if err != nil {
		t.Fatalf("AddFilter failed: %v", err)
	}
	if ok, err := audit.SetInitParam("level", "full"); err != nil || !ok {
		t.Fatalf("SetInitParam filter ok/err = %v/%v, want true/nil", ok, err)
	}
	if err := audit.AddMappingForURLPatterns(0, true, "/orders"); err != nil {
		t.Fatalf("AddMappingForURLPatterns failed: %v", err)
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

	want := []string{
		"context-init",
		"filter-init:audit:full",
		"servlet-init:orders:utf-8",
		"request-init",
		"filter",
		"servlet",
		"request-destroy",
		"servlet-destroy",
		"filter-destroy",
		"context-destroy",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestDeploymentFromRegistrationRejectsUnknownServletName(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	filter, err := registry.AddFilter("audit", servlet.FilterFunc(noopContainerFilter))
	if err != nil {
		t.Fatalf("AddFilter failed: %v", err)
	}
	if err := filter.AddMappingForServletNames(0, false, "missing"); err != nil {
		t.Fatalf("AddMappingForServletNames failed: %v", err)
	}
	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	snapshot, err := registry.Freeze()
	if err != nil {
		t.Fatalf("Freeze failed: %v", err)
	}
	_, err = DeploymentFromRegistration(app, snapshot)
	if !errors.Is(err, ErrUnknownServletName) {
		t.Fatalf("DeploymentFromRegistration err = %v, want ErrUnknownServletName", err)
	}
}

func TestDeploymentFromRegistrationRequiresFrozenSnapshot(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	orders, err := registry.AddServlet("orders", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return nil
	}))
	if err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	if conflicts, err := orders.AddMapping("/orders"); err != nil || len(conflicts) != 0 {
		t.Fatalf("AddMapping conflicts/err = %#v/%v, want none/nil", conflicts, err)
	}
	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	_, err = DeploymentFromRegistration(app, registry.Snapshot())
	if !errors.Is(err, registration.ErrSnapshotNotFrozen) {
		t.Fatalf("DeploymentFromRegistration err = %v, want ErrSnapshotNotFrozen", err)
	}
}

func TestDeploymentInitializesRegisteredServletsOnceInStartupOrder(t *testing.T) {
	t.Parallel()

	var calls []string
	registry := registration.NewRegistry()
	late, err := registry.AddServlet("late", &configRecordingServlet{calls: &calls})
	if err != nil {
		t.Fatalf("AddServlet late failed: %v", err)
	}
	if err := late.SetLoadOnStartup(20); err != nil {
		t.Fatalf("SetLoadOnStartup late failed: %v", err)
	}
	if conflicts, err := late.AddMapping("/late", "/late/*"); err != nil || len(conflicts) != 0 {
		t.Fatalf("AddMapping late conflicts/err = %#v/%v, want none/nil", conflicts, err)
	}
	early, err := registry.AddServlet("early", &configRecordingServlet{calls: &calls})
	if err != nil {
		t.Fatalf("AddServlet early failed: %v", err)
	}
	if err := early.SetLoadOnStartup(1); err != nil {
		t.Fatalf("SetLoadOnStartup early failed: %v", err)
	}
	if conflicts, err := early.AddMapping("/early"); err != nil || len(conflicts) != 0 {
		t.Fatalf("AddMapping early conflicts/err = %#v/%v, want none/nil", conflicts, err)
	}
	lazy, err := registry.AddServlet("lazy", &configRecordingServlet{calls: &calls})
	if err != nil {
		t.Fatalf("AddServlet lazy failed: %v", err)
	}
	if conflicts, err := lazy.AddMapping("/lazy"); err != nil || len(conflicts) != 0 {
		t.Fatalf("AddMapping lazy conflicts/err = %#v/%v, want none/nil", conflicts, err)
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
	application, err := NewApplication(context.Background(), deployment)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := application.Stop(context.Background()); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	want := []string{
		"servlet-init:early:",
		"servlet-init:late:",
		"servlet-init:lazy:",
		"servlet-destroy",
		"servlet-destroy",
		"servlet-destroy",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

type configRecordingServlet struct {
	calls *[]string
}

func (s *configRecordingServlet) Init(_ context.Context, cfg servlet.ServletConfig) error {
	value, _ := cfg.InitParam("encoding")
	*s.calls = append(*s.calls, "servlet-init:"+cfg.Name()+":"+value)
	return nil
}

func (s *configRecordingServlet) Serve(context.Context, *servlet.Request, servlet.Response) error {
	*s.calls = append(*s.calls, "servlet")
	return nil
}

func (s *configRecordingServlet) Destroy(context.Context) error {
	*s.calls = append(*s.calls, "servlet-destroy")
	return nil
}

type configRecordingFilter struct {
	calls *[]string
}

func (f *configRecordingFilter) Init(_ context.Context, cfg servlet.FilterConfig) error {
	value, _ := cfg.InitParam("level")
	*f.calls = append(*f.calls, "filter-init:"+cfg.Name()+":"+value)
	return nil
}

func (f *configRecordingFilter) Filter(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
	*f.calls = append(*f.calls, "filter")
	return chain.Next(ctx, req, res)
}

func (f *configRecordingFilter) Destroy(context.Context) error {
	*f.calls = append(*f.calls, "filter-destroy")
	return nil
}

func recordContainerFilter(name string, calls *[]string) servlet.Filter {
	return servlet.FilterFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
		*calls = append(*calls, name)
		return chain.Next(ctx, req, res)
	})
}

func noopContainerFilter(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
	return chain.Next(ctx, req, res)
}
