package tck

import (
	"context"
	"testing"

	"goark.dev/arkarta/servlet/session"
)

// SessionManagerFactory 创建待验证的会话管理器。
type SessionManagerFactory func() session.Manager

// RunSessionManager 执行 Servlet Session Profile 的管理器兼容性测试。
func RunSessionManager(t *testing.T, factory SessionManagerFactory) {
	t.Helper()
	t.Run("create_get_and_attributes", func(t *testing.T) {
		runSessionCreateGetAttributes(t, factory)
	})
	t.Run("renew_id", func(t *testing.T) {
		runSessionRenewID(t, factory)
	})
	t.Run("destroy", func(t *testing.T) {
		runSessionDestroy(t, factory)
	})
}

func runSessionCreateGetAttributes(t *testing.T, factory SessionManagerFactory) {
	t.Helper()
	manager := factory()
	created, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID() == "" {
		t.Fatal("session id should not be empty")
	}
	if err := created.SetAttribute("principal", "alice"); err != nil {
		t.Fatalf("SetAttribute failed: %v", err)
	}

	loaded, ok, err := manager.Get(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok {
		t.Fatal("Get should find created session")
	}
	value, ok := loaded.Attribute("principal")
	if !ok || value != "alice" {
		t.Fatalf("principal = %v/%v, want alice/true", value, ok)
	}
}

func runSessionRenewID(t *testing.T, factory SessionManagerFactory) {
	t.Helper()
	manager := factory()
	created, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	oldID := created.ID()

	renewed, err := manager.RenewID(context.Background(), oldID)
	if err != nil {
		t.Fatalf("RenewID failed: %v", err)
	}
	if renewed.ID() == "" || renewed.ID() == oldID {
		t.Fatalf("renewed id = %q, old id = %q", renewed.ID(), oldID)
	}
	if _, ok, err := manager.Get(context.Background(), oldID); err != nil || ok {
		t.Fatalf("old id ok/err = %v/%v, want false/nil", ok, err)
	}
}

func runSessionDestroy(t *testing.T, factory SessionManagerFactory) {
	t.Helper()
	manager := factory()
	created, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := manager.Destroy(context.Background(), created.ID()); err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}
	if _, ok, err := manager.Get(context.Background(), created.ID()); err != nil || ok {
		t.Fatalf("Get destroyed ok/err = %v/%v, want false/nil", ok, err)
	}
	if created.IsValid() {
		t.Fatal("destroyed session should be invalid")
	}
}
