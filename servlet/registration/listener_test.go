package registration_test

import (
	"errors"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/registration"
	"goark.dev/arkarta/servlet/session"
)

func TestListenerRegistrationSnapshot(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	contextListener, err := registry.AddContextListener(servlet.ContextListenerFunc{})
	if err != nil {
		t.Fatalf("AddContextListener failed: %v", err)
	}
	requestListener, err := registry.AddListener(servlet.RequestListenerFunc{})
	if err != nil {
		t.Fatalf("AddListener request failed: %v", err)
	}
	sessionListener, err := registry.AddSessionListener(session.ListenerFunc{})
	if err != nil {
		t.Fatalf("AddSessionListener failed: %v", err)
	}
	contextAttributeListener, err := registry.AddListener(servlet.ContextAttributeListenerFunc{})
	if err != nil {
		t.Fatalf("AddListener context attribute failed: %v", err)
	}
	requestAttributeListener, err := registry.AddListener(servlet.RequestAttributeListenerFunc{})
	if err != nil {
		t.Fatalf("AddListener request attribute failed: %v", err)
	}
	sessionAttributeListener, err := registry.AddSessionAttributeListener(session.AttributeListenerFunc{})
	if err != nil {
		t.Fatalf("AddSessionAttributeListener failed: %v", err)
	}
	if err := sessionListener.SetClassName("custom.SessionListener"); err != nil {
		t.Fatalf("SetClassName failed: %v", err)
	}

	if contextListener.Kind() != registration.ListenerContext || requestListener.Kind() != registration.ListenerRequest {
		t.Fatalf("listener kinds = %q/%q, want context/request", contextListener.Kind(), requestListener.Kind())
	}
	if contextAttributeListener.Kind() != registration.ListenerContextAttribute ||
		requestAttributeListener.Kind() != registration.ListenerRequestAttribute ||
		sessionAttributeListener.Kind() != registration.ListenerSessionAttribute {
		t.Fatalf("attribute listener kinds = %q/%q/%q", contextAttributeListener.Kind(), requestAttributeListener.Kind(), sessionAttributeListener.Kind())
	}
	snapshot := registry.Snapshot()
	listeners := snapshot.Listeners()
	if len(listeners) != 6 {
		t.Fatalf("listener count = %d, want 6", len(listeners))
	}
	if listeners[0].Order() != 0 || listeners[1].Order() != 1 || listeners[5].Order() != 5 {
		t.Fatalf("listener order = %d/%d/%d, want 0/1/5", listeners[0].Order(), listeners[1].Order(), listeners[5].Order())
	}
	if listeners[2].Kind() != registration.ListenerSession || listeners[2].ClassName() != "custom.SessionListener" {
		t.Fatalf("session listener descriptor = %q/%q", listeners[2].Kind(), listeners[2].ClassName())
	}
}

func TestListenerRegistrationRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	registry := registration.NewRegistry()
	var nilContextListener servlet.ContextListener
	if _, err := registry.AddContextListener(nilContextListener); !errors.Is(err, registration.ErrNilListener) {
		t.Fatalf("nil context listener err = %v, want ErrNilListener", err)
	}
	if _, err := registry.AddListener(struct{}{}); !errors.Is(err, registration.ErrNilListener) {
		t.Fatalf("unsupported listener err = %v, want ErrNilListener", err)
	}
}
