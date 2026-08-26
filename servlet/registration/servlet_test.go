package registration_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/registration"
)

func TestServletRegistrationMappingsAndSnapshot(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	orders, err := registry.AddServlet("orders", servlet.HandlerFunc(noopServe))
	if err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	if orders.Name() != "orders" {
		t.Fatalf("name = %q, want orders", orders.Name())
	}
	if orders.ClassName() == "" {
		t.Fatal("class name should be inferred from Go type")
	}
	added, err := orders.SetInitParam("encoding", "utf-8")
	if err != nil || !added {
		t.Fatalf("SetInitParam added/err = %v/%v, want true/nil", added, err)
	}
	added, err = orders.SetInitParam("encoding", "gbk")
	if err != nil || added {
		t.Fatalf("duplicate init param added/err = %v/%v, want false/nil", added, err)
	}
	conflicts, err := orders.SetInitParams(map[string]string{
		"encoding": "gbk",
		"pageSize": "100",
	})
	if err != nil {
		t.Fatalf("SetInitParams failed: %v", err)
	}
	if !reflect.DeepEqual(conflicts, []string{"encoding"}) {
		t.Fatalf("conflicts = %#v, want encoding", conflicts)
	}
	if _, ok := orders.InitParam("pageSize"); ok {
		t.Fatal("conflicting bulk init params must not be partially written")
	}
	if conflicts, err := orders.AddMapping("/orders", "/orders/*"); err != nil || len(conflicts) != 0 {
		t.Fatalf("AddMapping conflicts/err = %#v/%v, want none/nil", conflicts, err)
	}
	if err := orders.SetAsyncSupported(true); err != nil {
		t.Fatalf("SetAsyncSupported failed: %v", err)
	}
	if err := orders.SetLoadOnStartup(1); err != nil {
		t.Fatalf("SetLoadOnStartup failed: %v", err)
	}
	if err := orders.SetRunAsRole("admin"); err != nil {
		t.Fatalf("SetRunAsRole failed: %v", err)
	}

	other, err := registry.AddServlet("other", servlet.HandlerFunc(noopServe))
	if err != nil {
		t.Fatalf("AddServlet other failed: %v", err)
	}
	conflicts, err = other.AddMapping("/orders")
	if err != nil {
		t.Fatalf("AddMapping other failed: %v", err)
	}
	if !reflect.DeepEqual(conflicts, []string{"/orders"}) {
		t.Fatalf("conflicts = %#v, want /orders", conflicts)
	}
	if got := other.Mappings(); len(got) != 0 {
		t.Fatalf("other mappings = %#v, want empty", got)
	}

	snapshot := registry.Snapshot()
	servlets := snapshot.Servlets()
	if len(servlets) != 2 {
		t.Fatalf("servlet count = %d, want 2", len(servlets))
	}
	first := servlets[0]
	if first.Name() != "orders" || !first.AsyncSupported() || first.RunAsRole() != "admin" {
		t.Fatalf("servlet descriptor = %q/%v/%q, want orders/true/admin", first.Name(), first.AsyncSupported(), first.RunAsRole())
	}
	if order, ok := first.LoadOnStartup(); !ok || order != 1 {
		t.Fatalf("loadOnStartup = %d/%v, want 1/true", order, ok)
	}
	if !reflect.DeepEqual(first.Mappings(), []string{"/orders", "/orders/*"}) {
		t.Fatalf("mappings = %#v, want orders mappings", first.Mappings())
	}
	params := first.InitParams()
	params["encoding"] = "changed"
	if again := snapshot.Servlets()[0].InitParams()["encoding"]; again != "utf-8" {
		t.Fatalf("snapshot init param mutated to %q", again)
	}
}

func TestServletRegistrationRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	if _, err := registry.AddServlet("", servlet.HandlerFunc(noopServe)); !errors.Is(err, registration.ErrInvalidName) {
		t.Fatalf("empty servlet name err = %v, want ErrInvalidName", err)
	}
	var nilHandler servlet.HandlerFunc
	if _, err := registry.AddServlet("nil", nilHandler); !errors.Is(err, registration.ErrNilServlet) {
		t.Fatalf("nil servlet err = %v, want ErrNilServlet", err)
	}
	if _, err := registry.AddServlet("orders", servlet.HandlerFunc(noopServe)); err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	if _, err := registry.AddServlet("orders", servlet.HandlerFunc(noopServe)); !errors.Is(err, registration.ErrDuplicateRegistration) {
		t.Fatalf("duplicate servlet err = %v, want ErrDuplicateRegistration", err)
	}
	orders, ok := registry.Servlet("orders")
	if !ok {
		t.Fatal("Servlet should find orders")
	}
	if _, err := orders.SetInitParam("", "bad"); !errors.Is(err, registration.ErrInvalidInitParamName) {
		t.Fatalf("empty init param err = %v, want ErrInvalidInitParamName", err)
	}
	if _, err := orders.AddMapping("orders"); !errors.Is(err, servlet.ErrInvalidMappingPattern) {
		t.Fatalf("invalid mapping err = %v, want ErrInvalidMappingPattern", err)
	}
}

func TestRegistryFreezeRejectsFurtherMutation(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	orders, err := registry.AddServlet("orders", servlet.HandlerFunc(noopServe))
	if err != nil {
		t.Fatalf("AddServlet failed: %v", err)
	}
	snapshot, err := registry.Freeze()
	if err != nil {
		t.Fatalf("Freeze failed: %v", err)
	}
	if !snapshot.Frozen() || !registry.Frozen() {
		t.Fatal("registry and snapshot should be frozen")
	}
	if _, err := orders.SetInitParam("encoding", "utf-8"); !errors.Is(err, registration.ErrRegistryFrozen) {
		t.Fatalf("SetInitParam err = %v, want ErrRegistryFrozen", err)
	}
	if _, err := orders.AddMapping("/orders"); !errors.Is(err, registration.ErrRegistryFrozen) {
		t.Fatalf("AddMapping err = %v, want ErrRegistryFrozen", err)
	}
	if _, err := registry.AddFilter("audit", servlet.FilterFunc(noopFilter)); !errors.Is(err, registration.ErrRegistryFrozen) {
		t.Fatalf("AddFilter after freeze err = %v, want ErrRegistryFrozen", err)
	}
}

func noopServe(context.Context, *servlet.Request, servlet.Response) error {
	return nil
}

func noopFilter(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
	return chain.Next(ctx, req, res)
}
