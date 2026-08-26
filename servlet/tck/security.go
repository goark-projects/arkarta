package tck

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/security"
)

// RunSecurity 执行 Servlet Security Profile 的兼容性测试。
func RunSecurity(t *testing.T) {
	t.Helper()
	t.Run("basic_authentication_binds_identity", runBasicAuthenticationBindsIdentity)
	t.Run("basic_authentication_challenges_missing_credentials", runBasicAuthenticationChallengesMissingCredentials)
	t.Run("method_constraints_and_role_mapping", runMethodConstraintsAndRoleMapping)
	t.Run("run_as_scope_restores_previous_role", runRunAsScopeRestoresPreviousRole)
}

func runBasicAuthenticationBindsIdentity(t *testing.T) {
	t.Helper()
	realm := security.NewStaticRealm(security.WithStaticUser("alice", "secret", "orders"))
	authenticator := security.NewBasicAuthenticator(realm, security.WithBasicRealmName("tck"))
	httpRequest := httptest.NewRequest(http.MethodGet, "/secure", nil)
	httpRequest.SetBasicAuth("alice", "secret")
	req, err := servlet.NewRequest(httpRequest)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	ok, err := security.Authenticate(context.Background(), req, newResponseStub(), authenticator)
	if err != nil || !ok {
		t.Fatalf("Authenticate ok/err = %v/%v, want true/nil", ok, err)
	}
	if security.RemoteUser(req) != "alice" || security.AuthType(req) != security.AuthTypeBasic {
		t.Fatalf("identity user/auth = %q/%q, want alice/BASIC", security.RemoteUser(req), security.AuthType(req))
	}
	if !security.UserInRole(req, "orders") {
		t.Fatal("authenticated identity should expose orders role")
	}
}

func runBasicAuthenticationChallengesMissingCredentials(t *testing.T) {
	t.Helper()
	realm := security.NewStaticRealm(security.WithStaticUser("alice", "secret", "orders"))
	authenticator := security.NewBasicAuthenticator(realm, security.WithBasicRealmName("tck"))
	req := newTCKRequest(t, http.MethodGet, "/secure")
	res := newResponseStub()

	ok, err := security.Authenticate(context.Background(), req, res, authenticator)
	if err != nil || ok {
		t.Fatalf("Authenticate ok/err = %v/%v, want false/nil", ok, err)
	}
	if res.Status() != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.Status())
	}
	if res.Header().Get("WWW-Authenticate") != `Basic realm="tck"` {
		t.Fatalf("challenge = %q, want Basic realm", res.Header().Get("WWW-Authenticate"))
	}
}

func runMethodConstraintsAndRoleMapping(t *testing.T) {
	t.Helper()
	constraint := security.NewConstraint(
		security.WithRoles("reader"),
		security.WithMethodConstraint(http.MethodPost, security.NewConstraint(
			security.WithRoles("writer"),
			security.WithRoleMapping("writer", "orders:write"),
		)),
	)

	readReq := newTCKRequest(t, http.MethodGet, "/orders")
	security.BindIdentity(readReq, security.NewIdentity(security.PrincipalFunc(func() string { return "reader" }), security.AuthTypeBasic, "reader"))
	if err := constraint.Authorize(context.Background(), readReq); err != nil {
		t.Fatalf("GET Authorize failed: %v", err)
	}

	writeReq := newTCKRequest(t, http.MethodPost, "/orders")
	security.BindIdentity(writeReq, security.NewIdentity(security.PrincipalFunc(func() string { return "writer" }), security.AuthTypeBasic, "orders:write"))
	if err := constraint.Authorize(context.Background(), writeReq); err != nil {
		t.Fatalf("POST Authorize with role mapping failed: %v", err)
	}

	deniedReq := newTCKRequest(t, http.MethodPost, "/orders")
	security.BindIdentity(deniedReq, security.NewIdentity(security.PrincipalFunc(func() string { return "reader" }), security.AuthTypeBasic, "reader"))
	err := constraint.Authorize(context.Background(), deniedReq)
	var status servlet.StatusError
	if !errors.As(err, &status) || status.StatusCode() != http.StatusForbidden {
		t.Fatalf("POST denied err = %v, want 403 status error", err)
	}
}

func runRunAsScopeRestoresPreviousRole(t *testing.T) {
	t.Helper()
	req := newTCKRequest(t, http.MethodGet, "/run-as")
	security.BindIdentity(req, security.NewIdentity(security.PrincipalFunc(func() string { return "svc" }), security.AuthTypeBasic, "user"))

	if security.UserInRole(req, "admin") {
		t.Fatal("admin role should not be active before RunAs")
	}
	if err := security.RunAs(req, "admin", func() error {
		if !security.UserInRole(req, "admin") {
			t.Fatal("admin role should be active inside RunAs")
		}
		return nil
	}); err != nil {
		t.Fatalf("RunAs failed: %v", err)
	}
	if security.UserInRole(req, "admin") {
		t.Fatal("RunAs role should be restored after callback")
	}
	if !security.UserInRole(req, "user") {
		t.Fatal("original identity role should remain after RunAs")
	}
}
