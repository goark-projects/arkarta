package session

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

func TestMemoryManagerPassivatesAndActivatesSession(t *testing.T) {
	t.Parallel()

	manager := NewMemoryManager(WithIDGenerator(sequenceID("S1")))
	current, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	value := &activationValue{}
	if err := current.SetAttribute("token", value); err != nil {
		t.Fatalf("SetAttribute failed: %v", err)
	}
	store := NewMemoryStore()

	if err := manager.Passivate(context.Background(), current.ID(), store); err != nil {
		t.Fatalf("Passivate failed: %v", err)
	}
	if !reflect.DeepEqual(value.calls, []string{"passivate"}) {
		t.Fatalf("activation calls = %#v, want passivate", value.calls)
	}

	next := NewMemoryManager()
	activated, ok, err := next.Activate(context.Background(), current.ID(), store)
	if err != nil || !ok {
		t.Fatalf("Activate ok/err = %v/%v, want true/nil", ok, err)
	}
	if activated.ID() != "S1" || activated.IsNew() {
		t.Fatalf("activated session id/new = %q/%v, want S1/false", activated.ID(), activated.IsNew())
	}
	if got, exists := activated.Attribute("token"); !exists || got != value {
		t.Fatalf("activated attr = %v/%v, want original value", got, exists)
	}
	if !reflect.DeepEqual(value.calls, []string{"passivate", "activate"}) {
		t.Fatalf("activation calls = %#v, want passivate/activate", value.calls)
	}
}

func TestMemoryManagerConcurrentSessionAccess(t *testing.T) {
	manager := NewMemoryManager(WithIDGenerator(sequenceID("S1")))
	current, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			name := string(rune('a' + index%8))
			for j := 0; j < 100; j++ {
				if err := current.SetAttribute(name, j); err != nil {
					t.Errorf("SetAttribute failed: %v", err)
					return
				}
				if _, _, err := manager.Get(context.Background(), current.ID()); err != nil {
					t.Errorf("Get failed: %v", err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}

type activationValue struct {
	calls []string
}

func (v *activationValue) SessionWillPassivate(ActivationEvent) {
	v.calls = append(v.calls, "passivate")
}

func (v *activationValue) SessionDidActivate(ActivationEvent) {
	v.calls = append(v.calls, "activate")
}
