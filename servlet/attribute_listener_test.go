package servlet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRequestAttributeListenerEvents(t *testing.T) {
	t.Parallel()

	var events []string
	listener := RequestAttributeListenerFunc{
		Added: func(_ context.Context, event RequestAttributeEvent) {
			events = append(events, "add:"+event.Name)
		},
		Replaced: func(_ context.Context, event RequestAttributeEvent) {
			events = append(events, "replace:"+event.Name)
		},
		Removed: func(_ context.Context, event RequestAttributeEvent) {
			events = append(events, "remove:"+event.Name)
		},
	}
	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/", nil), WithRequestAttributeListener(listener))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	req.SetAttribute("trace", "a")
	req.SetAttribute("trace", "b")
	req.SetAttribute("trace", nil)

	want := []string{"add:trace", "replace:trace", "remove:trace"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestContextAttributeListenerEvents(t *testing.T) {
	t.Parallel()

	var events []string
	app, err := NewWebApp("orders", WithContextAttributeListener(ContextAttributeListenerFunc{
		Added: func(_ context.Context, event ContextAttributeEvent) {
			events = append(events, "add:"+event.Name)
		},
		Replaced: func(_ context.Context, event ContextAttributeEvent) {
			events = append(events, "replace:"+event.Name)
		},
		Removed: func(_ context.Context, event ContextAttributeEvent) {
			events = append(events, "remove:"+event.Name)
		},
	}))
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}

	app.SetAttribute("mode", "blue")
	app.SetAttribute("mode", "green")
	app.SetAttribute("mode", nil)

	want := []string{"add:mode", "replace:mode", "remove:mode"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
