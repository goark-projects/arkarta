package session

import (
	"context"
	"reflect"
	"testing"
)

func TestMemoryManagerSessionListenerEvents(t *testing.T) {
	t.Parallel()

	var calls []string
	manager := NewMemoryManager(
		WithIDGenerator(sequenceID("s1", "s2")),
		WithListener(ListenerFunc{
			Created: func(_ context.Context, event Event) error {
				calls = append(calls, "created:"+event.Session.ID())
				return nil
			},
			Destroyed: func(_ context.Context, event Event) error {
				calls = append(calls, "destroyed:"+event.Session.ID())
				return nil
			},
			IDChanged: func(_ context.Context, event IDChangedEvent) error {
				calls = append(calls, "id:"+event.OldID+"->"+event.NewID)
				return nil
			},
		}),
	)

	session, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	renewed, err := manager.RenewID(context.Background(), session.ID())
	if err != nil {
		t.Fatalf("RenewID failed: %v", err)
	}
	if err := renewed.Invalidate(); err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	want := []string{"created:s1", "id:s1->s2", "destroyed:s2"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}
