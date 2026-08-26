package servlet

import (
	"errors"
	"net/http"
	"testing"
)

func TestResponseHelpersSetHeaders(t *testing.T) {
	t.Parallel()

	res := newTestResponse()
	if err := SetContentType(res, "application/json"); err != nil {
		t.Fatalf("SetContentType failed: %v", err)
	}
	if err := SetCharacterEncoding(res, "utf-8"); err != nil {
		t.Fatalf("SetCharacterEncoding failed: %v", err)
	}
	if err := SetContentLength(res, 12); err != nil {
		t.Fatalf("SetContentLength failed: %v", err)
	}
	if err := AddCookie(res, &http.Cookie{Name: "sid", Value: "abc", HttpOnly: true}); err != nil {
		t.Fatalf("AddCookie failed: %v", err)
	}

	if ContentType(res) != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q, want application/json charset", ContentType(res))
	}
	if CharacterEncoding(res) != "utf-8" {
		t.Fatalf("charset = %q, want utf-8", CharacterEncoding(res))
	}
	if res.Header().Get("Content-Length") != "12" {
		t.Fatalf("content length = %q, want 12", res.Header().Get("Content-Length"))
	}
	if got := res.Header().Get("Set-Cookie"); got != "sid=abc; HttpOnly" {
		t.Fatalf("set-cookie = %q, want sid cookie", got)
	}
}

func TestRedirectResetsResponseAndSetsLocation(t *testing.T) {
	t.Parallel()

	res := newTestResponse()
	res.Header().Set("X-Old", "true")
	res.SetStatus(http.StatusCreated)

	if err := Redirect(res, "/login", http.StatusSeeOther); err != nil {
		t.Fatalf("Redirect failed: %v", err)
	}
	if res.Status() != http.StatusSeeOther {
		t.Fatalf("status = %d, want see other", res.Status())
	}
	if res.Header().Get("Location") != "/login" {
		t.Fatalf("location = %q, want /login", res.Header().Get("Location"))
	}
	if res.Header().Get("X-Old") != "" {
		t.Fatalf("old header leaked: %q", res.Header().Get("X-Old"))
	}
}

func TestSendErrorWritesSafeBody(t *testing.T) {
	t.Parallel()

	res := newTestResponse()
	if err := SendError(res, http.StatusForbidden, "denied"); err != nil {
		t.Fatalf("SendError failed: %v", err)
	}
	if res.Status() != http.StatusForbidden {
		t.Fatalf("status = %d, want forbidden", res.Status())
	}
	if res.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Fatalf("content type = %q, want text/plain utf-8", res.Header().Get("Content-Type"))
	}
	if res.body.String() != "denied\n" {
		t.Fatalf("body = %q, want denied newline", res.body.String())
	}
}

func TestResponseHelpersRejectCommittedResponse(t *testing.T) {
	t.Parallel()

	res := newTestResponse()
	if _, err := res.WriteString("committed"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	if err := Redirect(res, "/login", http.StatusFound); !errors.Is(err, ErrResponseCommitted) {
		t.Fatalf("Redirect err = %v, want ErrResponseCommitted", err)
	}
	if err := AddCookie(res, &http.Cookie{Name: "sid", Value: "abc"}); !errors.Is(err, ErrResponseCommitted) {
		t.Fatalf("AddCookie err = %v, want ErrResponseCommitted", err)
	}
}
