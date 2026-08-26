package session

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/nethttp"
)

func TestAccessorGetWithoutCreateReturnsEmpty(t *testing.T) {
	t.Parallel()

	accessor, err := NewAccessor(NewMemoryManager())
	if err != nil {
		t.Fatalf("NewAccessor failed: %v", err)
	}
	req := newSessionRequest(t, http.MethodGet, "/orders", "")
	current, ok, err := accessor.Get(context.Background(), req, nil, false)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if current != nil || ok {
		t.Fatalf("session = %v/%v, want nil/false", current, ok)
	}
	if id, ok := accessor.RequestedID(req); id != "" || ok {
		t.Fatalf("requested id = %q/%v, want empty/false", id, ok)
	}
}

func TestAccessorCreatesSessionAndWritesCookie(t *testing.T) {
	t.Parallel()

	manager := NewMemoryManager(WithIDGenerator(sequenceID("s1")))
	accessor, err := NewAccessor(manager)
	if err != nil {
		t.Fatalf("NewAccessor failed: %v", err)
	}
	req := newSessionRequest(t, http.MethodGet, "https://example.com/app/orders", "",
		servlet.WithRequestContextPath("/app"),
	)
	recorder := httptest.NewRecorder()
	res := nethttp.NewResponse(recorder)

	current, ok, err := accessor.Get(context.Background(), req, res, true)
	if err != nil {
		t.Fatalf("Get create failed: %v", err)
	}
	if !ok || current.ID() != "s1" || !current.IsNew() {
		t.Fatalf("session = %v/%v, want new s1", current, ok)
	}
	header := recorder.Header().Get("Set-Cookie")
	if !strings.Contains(header, "JSESSIONID=s1") ||
		!strings.Contains(header, "Path=/app") ||
		!strings.Contains(header, "HttpOnly") ||
		!strings.Contains(header, "Secure") ||
		!strings.Contains(header, "SameSite=Lax") {
		t.Fatalf("Set-Cookie = %q, want secure context session cookie", header)
	}
	if cached, ok := Current(req); !ok || cached.ID() != "s1" {
		t.Fatalf("current session = %v/%v, want s1/true", cached, ok)
	}
}

func TestAccessorLoadsRequestedSession(t *testing.T) {
	t.Parallel()

	manager := NewMemoryManager(WithIDGenerator(sequenceID("s1")))
	created, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	accessor, err := NewAccessor(manager)
	if err != nil {
		t.Fatalf("NewAccessor failed: %v", err)
	}
	req := newSessionRequest(t, http.MethodGet, "/orders", "JSESSIONID="+created.ID())

	if valid, err := accessor.RequestedIDValid(context.Background(), req); err != nil || !valid {
		t.Fatalf("RequestedIDValid = %v/%v, want true/nil", valid, err)
	}
	current, ok, err := accessor.Get(context.Background(), req, nil, false)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok || current.ID() != created.ID() || current.IsNew() {
		t.Fatalf("loaded session = %v/%v, want existing s1", current, ok)
	}
	if id, ok := accessor.RequestedID(req); id != "s1" || !ok {
		t.Fatalf("requested id = %q/%v, want s1/true", id, ok)
	}
	if source, ok := accessor.RequestedIDSource(req); source != TrackingCookie || !ok {
		t.Fatalf("requested id source = %q/%v, want COOKIE/true", source, ok)
	}
}

func TestAccessorLoadsURLTrackedSession(t *testing.T) {
	t.Parallel()

	manager := NewMemoryManager(WithIDGenerator(sequenceID("s1")))
	created, err := manager.Create(context.Background())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	accessor, err := NewAccessor(manager, WithTrackingModes(TrackingURL))
	if err != nil {
		t.Fatalf("NewAccessor failed: %v", err)
	}
	req := newSessionRequest(t, http.MethodGet, "/orders;jsessionid="+created.ID(), "")

	current, ok, err := accessor.Get(context.Background(), req, nil, false)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok || current.ID() != created.ID() {
		t.Fatalf("loaded session = %v/%v, want existing s1", current, ok)
	}
	if source, ok := accessor.RequestedIDSource(req); source != TrackingURL || !ok {
		t.Fatalf("requested id source = %q/%v, want URL/true", source, ok)
	}
}

func TestAccessorURLOnlyCreateDoesNotWriteCookie(t *testing.T) {
	t.Parallel()

	manager := NewMemoryManager(WithIDGenerator(sequenceID("s1")))
	accessor, err := NewAccessor(manager, WithTrackingModes(TrackingURL))
	if err != nil {
		t.Fatalf("NewAccessor failed: %v", err)
	}
	req := newSessionRequest(t, http.MethodGet, "/orders", "")

	current, ok, err := accessor.Get(context.Background(), req, nil, true)
	if err != nil {
		t.Fatalf("Get create failed: %v", err)
	}
	if !ok || current.ID() != "s1" {
		t.Fatalf("created session = %v/%v, want s1/true", current, ok)
	}
	got, err := accessor.EncodeURL(req, "/orders")
	if err != nil {
		t.Fatalf("EncodeURL failed: %v", err)
	}
	if got != "/orders;jsessionid=s1" {
		t.Fatalf("encoded URL = %q, want /orders;jsessionid=s1", got)
	}
}

func TestAccessorChangeIDWritesNewCookie(t *testing.T) {
	t.Parallel()

	manager := NewMemoryManager(WithIDGenerator(sequenceID("old", "new")))
	accessor, err := NewAccessor(manager, WithCookiePath("/"))
	if err != nil {
		t.Fatalf("NewAccessor failed: %v", err)
	}
	req := newSessionRequest(t, http.MethodPost, "/login", "")
	recorder := httptest.NewRecorder()
	res := nethttp.NewResponse(recorder)
	current, ok, err := accessor.Get(context.Background(), req, res, true)
	if err != nil || !ok || current.ID() != "old" {
		t.Fatalf("Get create = %v/%v/%v, want old/true/nil", current, ok, err)
	}
	recorder.Header().Del("Set-Cookie")

	newID, err := accessor.ChangeID(context.Background(), req, res)
	if err != nil {
		t.Fatalf("ChangeID failed: %v", err)
	}
	if newID != "new" {
		t.Fatalf("new id = %q, want new", newID)
	}
	if _, ok, err := manager.Get(context.Background(), "old"); err != nil || ok {
		t.Fatalf("old id ok/err = %v/%v, want false/nil", ok, err)
	}
	if _, ok, err := manager.Get(context.Background(), "new"); err != nil || !ok {
		t.Fatalf("new id ok/err = %v/%v, want true/nil", ok, err)
	}
	if header := recorder.Header().Get("Set-Cookie"); !strings.Contains(header, "JSESSIONID=new") {
		t.Fatalf("Set-Cookie = %q, want new session id", header)
	}
}

func TestAccessorRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	if _, err := NewAccessor(nil); !errors.Is(err, ErrNilManager) {
		t.Fatalf("nil manager err = %v, want ErrNilManager", err)
	}
	manager := NewMemoryManager()
	accessor, err := NewAccessor(manager)
	if err != nil {
		t.Fatalf("NewAccessor failed: %v", err)
	}
	if _, _, err := accessor.Get(context.Background(), nil, nil, false); !errors.Is(err, ErrNilRequest) {
		t.Fatalf("nil request err = %v, want ErrNilRequest", err)
	}
	req := newSessionRequest(t, http.MethodGet, "/orders", "")
	if _, _, err := accessor.Get(context.Background(), req, nil, true); !errors.Is(err, servlet.ErrNilResponse) {
		t.Fatalf("nil response err = %v, want ErrNilResponse", err)
	}
	if _, err := NewAccessor(manager, WithCookieName("")); !errors.Is(err, ErrInvalidCookieConfig) {
		t.Fatalf("empty cookie name err = %v, want ErrInvalidCookieConfig", err)
	}
}

func newSessionRequest(t *testing.T, method, target, cookie string, options ...servlet.RequestOption) *servlet.Request {
	t.Helper()
	httpRequest := httptest.NewRequest(method, target, nil)
	if cookie != "" {
		httpRequest.Header.Set("Cookie", cookie)
	}
	req, err := servlet.NewRequest(httpRequest, options...)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	return req
}
