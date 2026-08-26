package security

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goark.dev/arkarta/servlet"
)

func TestBasicAuthenticatorBindsPrincipalAndLogoutClearsIt(t *testing.T) {
	t.Parallel()

	authenticator := NewBasicAuthenticator(NewStaticRealm(
		WithStaticUser("alice", "secret", "admin", "ops"),
	))
	httpRequest := httptest.NewRequest(http.MethodGet, "/admin", nil)
	httpRequest.SetBasicAuth("alice", "secret")
	req, err := servlet.NewRequest(httpRequest)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	res := newAuthResponse()

	ok, err := Authenticate(context.Background(), req, res, authenticator)
	if err != nil || !ok {
		t.Fatalf("Authenticate ok/err = %v/%v, want true/nil", ok, err)
	}
	if RemoteUser(req) != "alice" || AuthType(req) != AuthTypeBasic || !UserInRole(req, "admin") {
		t.Fatalf("principal = %q authType=%q admin=%v", RemoteUser(req), AuthType(req), UserInRole(req, "admin"))
	}
	if err := Logout(context.Background(), req, res, authenticator); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
	if _, ok := CurrentPrincipal(req); ok {
		t.Fatal("principal should be cleared after logout")
	}
}

func TestBasicAuthenticatorChallengesMissingCredentials(t *testing.T) {
	t.Parallel()

	authenticator := NewBasicAuthenticator(NewStaticRealm())
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/admin", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	res := newAuthResponse()

	ok, err := Authenticate(context.Background(), req, res, authenticator)
	if err != nil || ok {
		t.Fatalf("Authenticate ok/err = %v/%v, want false/nil", ok, err)
	}
	if res.Status() != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.Status())
	}
	if challenge := res.Header().Get("WWW-Authenticate"); !strings.Contains(challenge, `Basic realm="arkarta"`) {
		t.Fatalf("challenge = %q, want Basic realm", challenge)
	}
}

func TestBasicAuthenticatorRejectsInvalidPassword(t *testing.T) {
	t.Parallel()

	authenticator := NewBasicAuthenticator(NewStaticRealm(
		WithStaticUser("alice", "secret", "admin"),
	))
	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "/admin", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	err = Login(context.Background(), req, "alice", "bad", authenticator)
	if !errors.Is(err, ErrAuthenticationFailed) {
		t.Fatalf("Login err = %v, want ErrAuthenticationFailed", err)
	}
}

type authResponse struct {
	header http.Header
	status int
}

func newAuthResponse() *authResponse {
	return &authResponse{header: make(http.Header), status: http.StatusOK}
}

func (r *authResponse) Header() http.Header {
	return r.header
}

func (r *authResponse) SetStatus(code int) {
	r.status = code
}

func (r *authResponse) Status() int {
	return r.status
}

func (r *authResponse) Write(data []byte) (int, error) {
	return len(data), nil
}

func (r *authResponse) WriteString(value string) (int, error) {
	return len(value), nil
}

func (r *authResponse) Flush() error {
	return nil
}

func (r *authResponse) Committed() bool {
	return false
}

func (r *authResponse) Reset() error {
	return nil
}

func (r *authResponse) BodyWriter() io.Writer {
	return io.Discard
}
