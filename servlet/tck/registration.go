package tck

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/registration"
)

// RunRegistration 执行 Servlet 动态注册元模型兼容性测试。
func RunRegistration(t *testing.T) {
	t.Helper()
	t.Run("servlet_registration_conflicts", runServletRegistrationConflicts)
	t.Run("filter_dispatcher_mappings", runFilterDispatcherMappings)
	t.Run("registry_freeze", runRegistrationFreeze)
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

func noopTCKServe(context.Context, *servlet.Request, servlet.Response) error {
	return servlet.NewHTTPError(http.StatusNoContent, "", nil)
}
