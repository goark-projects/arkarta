package tck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/session"
)

// RunSessionRequestBinding 执行 Session 与 Request/Response 绑定兼容性测试。
func RunSessionRequestBinding(t *testing.T, factory SessionManagerFactory) {
	t.Helper()
	t.Run("create_session_writes_cookie", func(t *testing.T) {
		runCreateSessionWritesCookie(t, factory)
	})
	t.Run("load_requested_session", func(t *testing.T) {
		runLoadRequestedSession(t, factory)
	})
	t.Run("change_session_id", func(t *testing.T) {
		runChangeSessionID(t, factory)
	})
}

func runCreateSessionWritesCookie(t *testing.T, factory SessionManagerFactory) {
	t.Helper()
	accessor := newTCKAccessor(t, factory())
	req := newTCKSessionRequest(t, "")
	res := newMemoryResponse()

	current, ok, err := accessor.Get(context.Background(), req, res, true)
	if err != nil {
		t.Fatalf("Get create failed: %v", err)
	}
	if !ok || current.ID() == "" || !current.IsNew() {
		t.Fatalf("session = %v/%v, want new session", current, ok)
	}
	cookie := res.Header().Get("Set-Cookie")
	if !strings.Contains(cookie, session.DefaultCookieName+"="+current.ID()) || !strings.Contains(cookie, "HttpOnly") {
		t.Fatalf("Set-Cookie = %q, want session id and HttpOnly", cookie)
	}
}

func runLoadRequestedSession(t *testing.T, factory SessionManagerFactory) {
	t.Helper()
	manager := factory()
	created, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	accessor := newTCKAccessor(t, manager)
	req := newTCKSessionRequest(t, session.DefaultCookieName+"="+created.ID())

	valid, err := accessor.RequestedIDValid(context.Background(), req)
	if err != nil || !valid {
		t.Fatalf("RequestedIDValid = %v/%v, want true/nil", valid, err)
	}
	current, ok, err := accessor.Get(context.Background(), req, nil, false)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok || current.ID() != created.ID() || current.IsNew() {
		t.Fatalf("loaded session = %v/%v, want existing %s", current, ok, created.ID())
	}
}

func runChangeSessionID(t *testing.T, factory SessionManagerFactory) {
	t.Helper()
	manager := factory()
	created, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	oldID := created.ID()
	accessor := newTCKAccessor(t, manager)
	req := newTCKSessionRequest(t, session.DefaultCookieName+"="+oldID)
	if _, ok, err := accessor.Get(context.Background(), req, nil, false); err != nil || !ok {
		t.Fatalf("Get existing ok/err = %v/%v, want true/nil", ok, err)
	}
	res := newMemoryResponse()

	newID, err := accessor.ChangeID(context.Background(), req, res)
	if err != nil {
		t.Fatalf("ChangeID failed: %v", err)
	}
	if newID == "" || newID == oldID {
		t.Fatalf("new id = %q, old id = %q", newID, oldID)
	}
	if _, ok, err := manager.Get(context.Background(), oldID); err != nil || ok {
		t.Fatalf("old id ok/err = %v/%v, want false/nil", ok, err)
	}
	if _, ok, err := manager.Get(context.Background(), newID); err != nil || !ok {
		t.Fatalf("new id ok/err = %v/%v, want true/nil", ok, err)
	}
	if cookie := res.Header().Get("Set-Cookie"); !strings.Contains(cookie, session.DefaultCookieName+"="+newID) {
		t.Fatalf("Set-Cookie = %q, want new session id", cookie)
	}
}

func newTCKAccessor(t *testing.T, manager session.Manager) *session.Accessor {
	t.Helper()
	accessor, err := session.NewAccessor(manager)
	if err != nil {
		t.Fatalf("NewAccessor failed: %v", err)
	}
	return accessor
}

func newTCKSessionRequest(t *testing.T, cookie string) *servlet.Request {
	t.Helper()
	httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	if cookie != "" {
		httpRequest.Header.Set("Cookie", cookie)
	}
	req, err := servlet.NewRequest(httpRequest)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	return req
}
