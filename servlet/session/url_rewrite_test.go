package session

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
)

func TestURLRewriterEncodesSessionIDInPath(t *testing.T) {
	t.Parallel()

	req := newRewriteRequest(t, "http://example.com/app/orders")
	rewriter, err := NewURLRewriter()
	if err != nil {
		t.Fatalf("NewURLRewriter failed: %v", err)
	}

	got, err := rewriter.EncodeURL(req, "/app/orders?q=1#top", "S123")
	if err != nil {
		t.Fatalf("EncodeURL failed: %v", err)
	}
	want := "/app/orders;jsessionid=S123?q=1#top"
	if got != want {
		t.Fatalf("encoded URL = %q, want %q", got, want)
	}
}

func TestURLRewriterReplacesExistingSessionID(t *testing.T) {
	t.Parallel()

	got, err := EncodeURL(newRewriteRequest(t, "http://example.com/orders"), "/orders;jsessionid=OLD?status=open", "NEW")
	if err != nil {
		t.Fatalf("EncodeURL failed: %v", err)
	}
	want := "/orders;jsessionid=NEW?status=open"
	if got != want {
		t.Fatalf("encoded URL = %q, want %q", got, want)
	}
}

func TestURLRewriterSkipsCookieAndExternalURLs(t *testing.T) {
	t.Parallel()

	httpReq := httptest.NewRequest(http.MethodGet, "http://example.com/orders", nil)
	httpReq.AddCookie(&http.Cookie{Name: DefaultCookieName, Value: "COOKIE"})
	req, err := servlet.NewRequest(httpReq)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	got, err := EncodeURL(req, "/orders", "S123")
	if err != nil {
		t.Fatalf("EncodeURL cookie failed: %v", err)
	}
	if got != "/orders" {
		t.Fatalf("cookie preferred URL = %q, want unchanged", got)
	}

	got, err = EncodeURL(newRewriteRequest(t, "http://example.com/orders"), "https://other.example/orders", "S123")
	if err != nil {
		t.Fatalf("EncodeURL external failed: %v", err)
	}
	if got != "https://other.example/orders" {
		t.Fatalf("external URL = %q, want unchanged", got)
	}
}

func TestURLRewriterSupportsCustomParameterAndCookiePolicy(t *testing.T) {
	t.Parallel()

	httpReq := httptest.NewRequest(http.MethodGet, "http://example.com/orders", nil)
	httpReq.AddCookie(&http.Cookie{Name: DefaultCookieName, Value: "COOKIE"})
	req, err := servlet.NewRequest(httpReq)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	rewriter, err := NewURLRewriter(WithRewriteParameterName("sid"), WithCookiePreferred(false))
	if err != nil {
		t.Fatalf("NewURLRewriter failed: %v", err)
	}
	got, err := rewriter.EncodeRedirectURL(req, "/orders", "S123")
	if err != nil {
		t.Fatalf("EncodeRedirectURL failed: %v", err)
	}
	if got != "/orders;sid=S123" {
		t.Fatalf("custom URL = %q, want /orders;sid=S123", got)
	}
}

func TestAccessorEncodesCurrentSessionURL(t *testing.T) {
	t.Parallel()

	accessor, err := NewAccessor(NewMemoryManager())
	if err != nil {
		t.Fatalf("NewAccessor failed: %v", err)
	}
	req := newRewriteRequest(t, "http://example.com/orders")
	session, _, err := accessor.Get(context.Background(), req, newNoopResponse(), true)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	got, err := accessor.EncodeURL(req, "/orders")
	if err != nil {
		t.Fatalf("Accessor EncodeURL failed: %v", err)
	}
	want := "/orders;jsessionid=" + session.ID()
	if got != want {
		t.Fatalf("accessor URL = %q, want %q", got, want)
	}
}

func TestURLRewriterRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	if _, err := NewURLRewriter(WithRewriteParameterName("bad/name")); !errors.Is(err, ErrInvalidURLRewriteConfig) {
		t.Fatalf("invalid parameter err = %v, want ErrInvalidURLRewriteConfig", err)
	}
}

func newRewriteRequest(t *testing.T, target string) *servlet.Request {
	t.Helper()
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, target, nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	return req
}

type noopResponse struct {
	header http.Header
}

func newNoopResponse() *noopResponse {
	return &noopResponse{header: make(http.Header)}
}

func (r *noopResponse) Header() http.Header {
	return r.header
}

func (r *noopResponse) SetStatus(int) {
}

func (r *noopResponse) Status() int {
	return http.StatusOK
}

func (r *noopResponse) Write(data []byte) (int, error) {
	return len(data), nil
}

func (r *noopResponse) WriteString(value string) (int, error) {
	return len(value), nil
}

func (r *noopResponse) Flush() error {
	return nil
}

func (r *noopResponse) Committed() bool {
	return false
}

func (r *noopResponse) Reset() error {
	return nil
}

func (r *noopResponse) BodyWriter() io.Writer {
	return io.Discard
}
