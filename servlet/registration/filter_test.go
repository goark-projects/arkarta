package registration_test

import (
	"errors"
	"reflect"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/registration"
)

func TestFilterRegistrationMappings(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	filter, err := registry.AddFilter("audit", servlet.FilterFunc(noopFilter))
	if err != nil {
		t.Fatalf("AddFilter failed: %v", err)
	}
	if ok, err := filter.SetInitParam("level", "full"); err != nil || !ok {
		t.Fatalf("SetInitParam ok/err = %v/%v, want true/nil", ok, err)
	}
	if err := filter.SetAsyncSupported(true); err != nil {
		t.Fatalf("SetAsyncSupported failed: %v", err)
	}
	if err := filter.AddMappingForURLPatterns(0, false, "/secure/*", "/admin/*"); err != nil {
		t.Fatalf("AddMappingForURLPatterns failed: %v", err)
	}
	dispatchers, err := registration.NewDispatcherTypes(registration.DispatcherRequest, registration.DispatcherError)
	if err != nil {
		t.Fatalf("NewDispatcherTypes failed: %v", err)
	}
	if err := filter.AddMappingForServletNames(dispatchers, true, "orders"); err != nil {
		t.Fatalf("AddMappingForServletNames failed: %v", err)
	}

	urlMappings := filter.URLPatternMappings()
	if len(urlMappings) != 1 {
		t.Fatalf("url mapping count = %d, want 1", len(urlMappings))
	}
	if !urlMappings[0].DispatcherTypes().Contains(registration.DispatcherRequest) {
		t.Fatal("zero dispatcher set should default to REQUEST")
	}
	if !reflect.DeepEqual(urlMappings[0].URLPatterns(), []string{"/secure/*", "/admin/*"}) {
		t.Fatalf("url patterns = %#v, want secure/admin", urlMappings[0].URLPatterns())
	}
	nameMappings := filter.ServletNameMappings()
	if len(nameMappings) != 1 {
		t.Fatalf("servlet-name mapping count = %d, want 1", len(nameMappings))
	}
	if !nameMappings[0].MatchAfter() || !nameMappings[0].DispatcherTypes().Contains(registration.DispatcherError) {
		t.Fatalf("name mapping flags invalid")
	}
	if !reflect.DeepEqual(nameMappings[0].ServletNames(), []string{"orders"}) {
		t.Fatalf("servlet names = %#v, want orders", nameMappings[0].ServletNames())
	}

	snapshot := registry.Snapshot()
	descriptor := snapshot.Filters()[0]
	if descriptor.Name() != "audit" || !descriptor.AsyncSupported() || descriptor.InitParams()["level"] != "full" {
		t.Fatalf("filter descriptor mismatch")
	}
	patterns := descriptor.URLPatternMappings()[0].URLPatterns()
	patterns[0] = "/changed/*"
	if got := snapshot.Filters()[0].URLPatternMappings()[0].URLPatterns()[0]; got != "/secure/*" {
		t.Fatalf("snapshot URL pattern mutated to %q", got)
	}
}

func TestFilterRegistrationRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	if _, err := registry.AddFilter(" ", servlet.FilterFunc(noopFilter)); !errors.Is(err, registration.ErrInvalidName) {
		t.Fatalf("blank filter name err = %v, want ErrInvalidName", err)
	}
	var nilFilter servlet.FilterFunc
	if _, err := registry.AddFilter("nil", nilFilter); !errors.Is(err, registration.ErrNilFilter) {
		t.Fatalf("nil filter err = %v, want ErrNilFilter", err)
	}
	filter, err := registry.AddFilter("audit", servlet.FilterFunc(noopFilter))
	if err != nil {
		t.Fatalf("AddFilter failed: %v", err)
	}
	if err := filter.AddMappingForURLPatterns(registration.DispatcherTypes(1<<7), false, "/secure/*"); !errors.Is(err, registration.ErrInvalidDispatcherTypes) {
		t.Fatalf("invalid dispatchers err = %v, want ErrInvalidDispatcherTypes", err)
	}
	if err := filter.AddMappingForURLPatterns(0, false, "bad"); !errors.Is(err, servlet.ErrInvalidMappingPattern) {
		t.Fatalf("invalid URL pattern err = %v, want ErrInvalidMappingPattern", err)
	}
	if err := filter.AddMappingForServletNames(0, false, ""); !errors.Is(err, registration.ErrInvalidName) {
		t.Fatalf("invalid servlet name err = %v, want ErrInvalidName", err)
	}
}

func TestDispatcherTypesList(t *testing.T) {
	t.Parallel()

	dispatchers, err := registration.NewDispatcherTypes(registration.DispatcherError, registration.DispatcherRequest)
	if err != nil {
		t.Fatalf("NewDispatcherTypes failed: %v", err)
	}
	want := []registration.DispatcherType{registration.DispatcherRequest, registration.DispatcherError}
	if !reflect.DeepEqual(dispatchers.List(), want) {
		t.Fatalf("list = %#v, want %#v", dispatchers.List(), want)
	}
	if _, err := registration.NewDispatcherTypes(registration.DispatcherType(99)); !errors.Is(err, registration.ErrInvalidDispatcherTypes) {
		t.Fatalf("invalid dispatcher err = %v, want ErrInvalidDispatcherTypes", err)
	}
}
