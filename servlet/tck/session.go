package tck

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"goark.dev/arkarta/servlet"
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
	t.Run("attribute_names_snapshot", func(t *testing.T) {
		runSessionAttributeNamesSnapshot(t, factory)
	})
}

// MemorySessionManagerFactory 创建内存会话管理器参考实现。
type MemorySessionManagerFactory func(options ...session.MemoryManagerOption) *session.MemoryManager

// RunMemorySessionProfile 执行 Arkarta 内存会话参考实现兼容性测试。
func RunMemorySessionProfile(t *testing.T, factory MemorySessionManagerFactory) {
	t.Helper()
	t.Run("store_passivate_activate_callbacks", func(t *testing.T) {
		runSessionStorePassivateActivateCallbacks(t, factory)
	})
	t.Run("ssl_tracking_binds_connection_id", func(t *testing.T) {
		runSSLTrackingBindsConnectionID(t, factory)
	})
	t.Run("webapp_cookie_config_is_inherited", func(t *testing.T) {
		runWebAppCookieConfigIsInherited(t, factory)
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

func runSessionAttributeNamesSnapshot(t *testing.T, factory SessionManagerFactory) {
	t.Helper()
	manager := factory()
	created, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := created.SetAttribute("b", 2); err != nil {
		t.Fatalf("SetAttribute b failed: %v", err)
	}
	if err := created.SetAttribute("a", 1); err != nil {
		t.Fatalf("SetAttribute a failed: %v", err)
	}
	names := created.AttributeNames()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Fatalf("AttributeNames = %#v, want sorted a/b", names)
	}
}

func runSessionStorePassivateActivateCallbacks(t *testing.T, factory MemorySessionManagerFactory) {
	t.Helper()
	manager := factory(session.WithIDGenerator(sequenceSessionIDs("S1")))
	current, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	value := &activationProbe{}
	if err := current.SetAttribute("token", value); err != nil {
		t.Fatalf("SetAttribute failed: %v", err)
	}
	store := session.NewMemoryStore()

	if err := manager.Passivate(context.Background(), current.ID(), store); err != nil {
		t.Fatalf("Passivate failed: %v", err)
	}
	if strings.Join(value.calls, ",") != "passivate" {
		t.Fatalf("activation calls = %#v, want passivate", value.calls)
	}

	restored, ok, err := factory().Activate(context.Background(), current.ID(), store)
	if err != nil || !ok {
		t.Fatalf("Activate ok/err = %v/%v, want true/nil", ok, err)
	}
	if restored.ID() != "S1" || restored.IsNew() {
		t.Fatalf("restored id/new = %q/%v, want S1/false", restored.ID(), restored.IsNew())
	}
	if got, exists := restored.Attribute("token"); !exists || got != value {
		t.Fatalf("restored token = %v/%v, want original value", got, exists)
	}
	if strings.Join(value.calls, ",") != "passivate,activate" {
		t.Fatalf("activation calls = %#v, want passivate/activate", value.calls)
	}
}

func runSSLTrackingBindsConnectionID(t *testing.T, factory MemorySessionManagerFactory) {
	t.Helper()
	manager := factory()
	accessor, err := session.NewAccessor(manager, session.WithTrackingModes(session.TrackingSSL))
	if err != nil {
		t.Fatalf("NewAccessor failed: %v", err)
	}
	req := newTCKRequest(t, http.MethodGet, "https://example.com/orders", servlet.WithRequestConnectionID("tls-conn-1"))

	current, ok, err := accessor.Get(context.Background(), req, nil, true)
	if err != nil || !ok {
		t.Fatalf("Get create ok/err = %v/%v, want true/nil", ok, err)
	}
	if current.ID() != "tls-conn-1" {
		t.Fatalf("session id = %q, want tls-conn-1", current.ID())
	}
	source, ok := accessor.RequestedIDSource(req)
	if !ok || source != session.TrackingSSL {
		t.Fatalf("requested source = %v/%v, want SSL/true", source, ok)
	}
}

func runWebAppCookieConfigIsInherited(t *testing.T, factory MemorySessionManagerFactory) {
	t.Helper()
	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	config, err := session.NewCookieConfig(
		session.WithCookieConfigName("ARKSESSION"),
		session.WithCookieConfigPath("/orders"),
		session.WithCookieConfigSecure(true),
	)
	if err != nil {
		t.Fatalf("NewCookieConfig failed: %v", err)
	}
	if err := session.ConfigureCookie(app, config); err != nil {
		t.Fatalf("ConfigureCookie failed: %v", err)
	}
	accessor, err := session.NewAccessorForWebApp(factory(session.WithIDGenerator(sequenceSessionIDs("S2"))), app)
	if err != nil {
		t.Fatalf("NewAccessorForWebApp failed: %v", err)
	}
	req := newTCKRequest(t, http.MethodGet, "http://example.com/orders/cart", servlet.WithRequestContextPath("/orders"))
	res := newResponseStub()

	if _, _, err := accessor.Get(context.Background(), req, res, true); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	header := res.Header().Get("Set-Cookie")
	if !strings.Contains(header, "ARKSESSION=S2") || !strings.Contains(header, "Path=/orders") || !strings.Contains(header, "Secure") {
		t.Fatalf("Set-Cookie = %q, want configured application cookie", header)
	}
}

func sequenceSessionIDs(ids ...string) session.IDGenerator {
	index := 0
	return func() (string, error) {
		if index >= len(ids) {
			return ids[len(ids)-1], nil
		}
		id := ids[index]
		index++
		return id, nil
	}
}

type activationProbe struct {
	calls []string
}

func (p *activationProbe) SessionWillPassivate(session.ActivationEvent) {
	p.calls = append(p.calls, "passivate")
}

func (p *activationProbe) SessionDidActivate(session.ActivationEvent) {
	p.calls = append(p.calls, "activate")
}
