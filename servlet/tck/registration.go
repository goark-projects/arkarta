package tck

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/registration"
)

// RunRegistration 执行 Servlet 动态注册元模型兼容性测试。
func RunRegistration(t *testing.T) {
	t.Helper()
	t.Run("servlet_registration_conflicts", runServletRegistrationConflicts)
	t.Run("filter_dispatcher_mappings", runFilterDispatcherMappings)
	t.Run("registry_freeze", runRegistrationFreeze)
	t.Run("deployment_requires_frozen_snapshot", runDeploymentRequiresFrozenSnapshot)
	t.Run("context_closes_after_webapp_start", runRegistrationContextClosesAfterWebAppStart)
	t.Run("load_on_startup_order_and_single_init", runLoadOnStartupOrderAndSingleInit)
}

func runServletRegistrationConflicts(t *testing.T) {
	t.Helper()
	registry := registration.NewRegistry()
	orders, err := registry.AddServlet("orders", servlet.HandlerFunc(noopTCKServe))
	if err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	if ok, err := orders.SetInitParam("encoding", "utf-8"); err != nil || !ok {
		t.Fatalf("SetInitParam ok/err = %v/%v, want true/nil", ok, err)
	}
	if conflicts, err := orders.AddMapping("/orders", "/orders/*"); err != nil || len(conflicts) != 0 {
		t.Fatalf("AddMapping conflicts/err = %#v/%v, want none/nil", conflicts, err)
	}
	other, err := registry.AddServlet("other", servlet.HandlerFunc(noopTCKServe))
	if err != nil {
		t.Fatalf("AddServlet other failed: %v", err)
	}
	conflicts, err := other.AddMapping("/orders")
	if err != nil {
		t.Fatalf("AddMapping other failed: %v", err)
	}
	if !reflect.DeepEqual(conflicts, []string{"/orders"}) {
		t.Fatalf("conflicts = %#v, want /orders", conflicts)
	}
}

func runFilterDispatcherMappings(t *testing.T) {
	t.Helper()
	registry := registration.NewRegistry()
	filter, err := registry.AddFilter("audit", servlet.FilterFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
		return chain.Next(ctx, req, res)
	}))
	if err != nil {
		t.Fatalf("AddFilter failed: %v", err)
	}
	dispatchers, err := registration.NewDispatcherTypes(registration.DispatcherRequest, registration.DispatcherError)
	if err != nil {
		t.Fatalf("NewDispatcherTypes failed: %v", err)
	}
	if err := filter.AddMappingForURLPatterns(dispatchers, true, "/secure/*"); err != nil {
		t.Fatalf("AddMappingForURLPatterns failed: %v", err)
	}
	mappings := registry.Snapshot().Filters()[0].URLPatternMappings()
	if len(mappings) != 1 || !mappings[0].MatchAfter() || !mappings[0].DispatcherTypes().Contains(registration.DispatcherError) {
		t.Fatalf("filter mapping = %#v, want request/error matchAfter", mappings)
	}
}

func runRegistrationFreeze(t *testing.T) {
	t.Helper()
	registry := registration.NewRegistry()
	target, err := registry.AddServlet("target", servlet.HandlerFunc(noopTCKServe))
	if err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	if _, err := registry.Freeze(); err != nil {
		t.Fatalf("Freeze failed: %v", err)
	}
	if _, err := target.AddMapping("/target"); !errors.Is(err, registration.ErrRegistryFrozen) {
		t.Fatalf("AddMapping after freeze err = %v, want ErrRegistryFrozen", err)
	}
}

func runDeploymentRequiresFrozenSnapshot(t *testing.T) {
	t.Helper()
	registry := registration.NewRegistry()
	target, err := registry.AddServlet("target", servlet.HandlerFunc(noopTCKServe))
	if err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	if _, err := target.AddMapping("/target"); err != nil {
		t.Fatalf("AddMapping failed: %v", err)
	}
	app, err := servlet.NewWebApp("tck")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	_, err = servletcontainer.DeploymentFromRegistration(app, registry.Snapshot())
	if !errors.Is(err, registration.ErrSnapshotNotFrozen) {
		t.Fatalf("DeploymentFromRegistration err = %v, want ErrSnapshotNotFrozen", err)
	}
}

func runRegistrationContextClosesAfterWebAppStart(t *testing.T) {
	t.Helper()
	app, err := servlet.NewWebApp("tck")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	ctx, err := registration.NewContext(app, nil)
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}
	if err := app.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	if err := app.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	t.Cleanup(func() {
		_ = app.Stop(context.Background())
		_ = app.Destroy(context.Background())
	})

	if _, err := ctx.AddServlet("late", servlet.HandlerFunc(noopTCKServe)); !errors.Is(err, registration.ErrRegistrationClosed) {
		t.Fatalf("AddServlet after start err = %v, want ErrRegistrationClosed", err)
	}
	if _, err := ctx.SetInitParam("late", "true"); !errors.Is(err, registration.ErrRegistrationClosed) {
		t.Fatalf("SetInitParam after start err = %v, want ErrRegistrationClosed", err)
	}
}

func runLoadOnStartupOrderAndSingleInit(t *testing.T) {
	t.Helper()
	var calls []string
	app, err := servlet.NewWebApp("tck")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	ctx, err := registration.NewContext(app, nil)
	if err != nil {
		t.Fatalf("NewContext failed: %v", err)
	}
	late, err := ctx.AddServlet("late", &lifecycleServlet{calls: &calls})
	if err != nil {
		t.Fatalf("AddServlet late failed: %v", err)
	}
	if conflicts, err := late.AddMapping("/late", "/late/*"); err != nil || len(conflicts) != 0 {
		t.Fatalf("late AddMapping conflicts/err = %#v/%v", conflicts, err)
	}
	if err := late.SetLoadOnStartup(10); err != nil {
		t.Fatalf("late SetLoadOnStartup failed: %v", err)
	}
	early, err := ctx.AddServlet("early", &lifecycleServlet{calls: &calls})
	if err != nil {
		t.Fatalf("AddServlet early failed: %v", err)
	}
	if conflicts, err := early.AddMapping("/early"); err != nil || len(conflicts) != 0 {
		t.Fatalf("early AddMapping conflicts/err = %#v/%v", conflicts, err)
	}
	if err := early.SetLoadOnStartup(1); err != nil {
		t.Fatalf("early SetLoadOnStartup failed: %v", err)
	}
	snapshot, err := ctx.Freeze()
	if err != nil {
		t.Fatalf("Freeze failed: %v", err)
	}
	deployment, err := servletcontainer.DeploymentFromRegistration(app, snapshot)
	if err != nil {
		t.Fatalf("DeploymentFromRegistration failed: %v", err)
	}
	application, err := servletcontainer.NewApplication(context.Background(), deployment)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	defer func() {
		if err := application.Stop(context.Background()); err != nil {
			t.Fatalf("Stop failed: %v", err)
		}
	}()

	want := []string{"init:early", "init:late"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("init calls = %#v, want %#v", calls, want)
	}
}

func noopTCKServe(context.Context, *servlet.Request, servlet.Response) error {
	return servlet.NewHTTPError(http.StatusNoContent, "", nil)
}
