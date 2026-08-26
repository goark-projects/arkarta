package session

import (
	"context"
	"reflect"
	"testing"
)

func TestMemorySessionAttributeAndBindingListeners(t *testing.T) {
	t.Parallel()

	var events []string
	manager := NewMemoryManager(
		WithIDGenerator(sequenceID("S1")),
		WithAttributeListener(AttributeListenerFunc{
			Added: func(event AttributeEvent) {
				events = append(events, "add:"+event.Name)
			},
			Replaced: func(event AttributeEvent) {
				events = append(events, "replace:"+event.Name)
			},
			Removed: func(event AttributeEvent) {
				events = append(events, "remove:"+event.Name)
			},
		}),
	)
	current, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	first := &bindingProbe{events: &events, name: "first"}
	second := &bindingProbe{events: &events, name: "second"}
	if err := current.SetAttribute("user", first); err != nil {
		t.Fatalf("SetAttribute first failed: %v", err)
	}
	if err := current.SetAttribute("user", second); err != nil {
		t.Fatalf("SetAttribute second failed: %v", err)
	}
	if err := current.RemoveAttribute("user"); err != nil {
		t.Fatalf("RemoveAttribute failed: %v", err)
	}

	want := []string{
		"bound:first",
		"add:user",
		"unbound:first",
		"bound:second",
		"replace:user",
		"unbound:second",
		"remove:user",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

type bindingProbe struct {
	events *[]string
	name   string
}

func (p *bindingProbe) ValueBound(BindingEvent) {
	*p.events = append(*p.events, "bound:"+p.name)
}

func (p *bindingProbe) ValueUnbound(BindingEvent) {
	*p.events = append(*p.events, "unbound:"+p.name)
}
