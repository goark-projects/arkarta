package servlet

import (
	"testing"
	"time"
)

func TestCookieStringWritesStandardAttributes(t *testing.T) {
	cookie := &Cookie{
		Name:        "sid",
		Value:       "abc",
		Path:        "/app",
		Domain:      ".example.test",
		Expires:     time.Date(2030, time.January, 2, 3, 4, 5, 0, time.FixedZone("offset", 8*60*60)),
		MaxAge:      3600,
		Secure:      true,
		HTTPOnly:    true,
		SameSite:    SameSiteStrictMode,
		Partitioned: true,
	}
	want := "sid=abc; Path=/app; Domain=example.test; Expires=Tue, 01 Jan 2030 19:04:05 GMT; Max-Age=3600; HttpOnly; Secure; SameSite=Strict; Partitioned"
	if got := cookie.String(); got != want {
		t.Fatalf("Cookie.String() = %q, want %q", got, want)
	}
}

func TestCookieStringRejectsInvalidName(t *testing.T) {
	if got := (&Cookie{Name: "bad name", Value: "value"}).String(); got != "" {
		t.Fatalf("Cookie.String() = %q, want empty", got)
	}
}
