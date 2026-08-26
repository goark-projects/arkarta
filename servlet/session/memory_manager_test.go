package session

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryManagerCreatesAndLoadsSession(t *testing.T) {
	t.Parallel()

	manager := NewMemoryManager(WithIDGenerator(sequenceID("s1")))
	session, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	loaded, ok, err := manager.Get(context.Background(), session.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok {
		t.Fatal("Get should find session")
	}
	if loaded.ID() != "s1" {
		t.Fatalf("id = %s, want s1", loaded.ID())
	}
	if loaded.IsNew() {
		t.Fatal("loaded session should not remain new after access")
	}
}

func TestMemoryManagerRenewsSessionID(t *testing.T) {
	t.Parallel()

	manager := NewMemoryManager(WithIDGenerator(sequenceID("old", "new")))
	session, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	renewed, err := manager.RenewID(context.Background(), session.ID())
	if err != nil {
		t.Fatalf("RenewID failed: %v", err)
	}
	if renewed.ID() != "new" {
		t.Fatalf("renewed id = %s, want new", renewed.ID())
	}
	if _, ok, err := manager.Get(context.Background(), "old"); err != nil || ok {
		t.Fatalf("old id ok/err = %v/%v, want false/nil", ok, err)
	}
	if _, ok, err := manager.Get(context.Background(), "new"); err != nil || !ok {
		t.Fatalf("new id ok/err = %v/%v, want true/nil", ok, err)
	}
}

func TestSessionInvalidateRemovesSession(t *testing.T) {
	t.Parallel()

	manager := NewMemoryManager(WithIDGenerator(sequenceID("s1")))
	session, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := session.Invalidate(); err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}
	if session.IsValid() {
		t.Fatal("session should be invalid")
	}
	if err := session.SetAttribute("user", "alice"); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("SetAttribute err = %v, want ErrInvalidSession", err)
	}
	if _, ok, err := manager.Get(context.Background(), "s1"); err != nil || ok {
		t.Fatalf("Get ok/err = %v/%v, want false/nil", ok, err)
	}
}

func TestMemoryManagerExpiresIdleSession(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 26, 16, 0, 0, 0, time.UTC)
	manager := NewMemoryManager(
		WithIDGenerator(sequenceID("s1")),
		WithClock(func() time.Time { return now }),
		WithMaxInactiveInterval(time.Minute),
	)
	session, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	now = now.Add(time.Minute + time.Nanosecond)
	if _, ok, err := manager.Get(context.Background(), session.ID()); err != nil || ok {
		t.Fatalf("Get expired ok/err = %v/%v, want false/nil", ok, err)
	}
	if session.IsValid() {
		t.Fatal("expired session should be invalid")
	}
}

func TestSessionNegativeInactiveIntervalNeverExpires(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 26, 16, 0, 0, 0, time.UTC)
	manager := NewMemoryManager(
		WithIDGenerator(sequenceID("s1")),
		WithClock(func() time.Time { return now }),
		WithMaxInactiveInterval(-1),
	)
	session, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	now = now.Add(365 * 24 * time.Hour)
	if _, ok, err := manager.Get(context.Background(), session.ID()); err != nil || !ok {
		t.Fatalf("Get ok/err = %v/%v, want true/nil", ok, err)
	}
}

func TestSessionZeroInactiveIntervalExpiresImmediately(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 26, 16, 0, 0, 0, time.UTC)
	manager := NewMemoryManager(
		WithIDGenerator(sequenceID("s1")),
		WithClock(func() time.Time { return now }),
		WithMaxInactiveInterval(0),
	)
	session, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if _, ok, err := manager.Get(context.Background(), session.ID()); err != nil || ok {
		t.Fatalf("Get ok/err = %v/%v, want false/nil", ok, err)
	}
}

func sequenceID(values ...string) IDGenerator {
	index := 0
	return func() (string, error) {
		if index >= len(values) {
			return "", errors.New("no id left")
		}
		value := values[index]
		index++
		return value, nil
	}
}
