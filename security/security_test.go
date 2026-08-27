package security_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"goark.dev/arkarta/security"
)

func TestAuthenticationNormalizesAuthoritiesAndContext(t *testing.T) {
	t.Parallel()

	auth := security.NewAuthentication(
		security.PrincipalFunc(func() string { return "alice" }),
		security.WithAuthorities("ROLE_USER", "ROLE_ADMIN", "ROLE_USER"),
		security.WithAuthenticated(true),
	)
	if !auth.Authenticated() || auth.Principal().Name() != "alice" {
		t.Fatalf("authentication = %#v, want authenticated alice", auth)
	}
	if got := auth.Authorities(); !reflect.DeepEqual(got, []security.Authority{"ROLE_ADMIN", "ROLE_USER"}) {
		t.Fatalf("authorities = %#v, want sorted unique", got)
	}
	if !auth.HasAuthority("ROLE_ADMIN") || auth.HasAuthority("ROLE_GUEST") {
		t.Fatalf("authority lookup failed")
	}

	ctx := security.ContextWithSecurity(context.Background(), security.NewContext(auth))
	current, ok := security.AuthenticationFromContext(ctx)
	if !ok || current.Principal().Name() != "alice" {
		t.Fatalf("context auth = %#v, ok=%v", current, ok)
	}
}

func TestAuthenticationManagerFuncAuthenticatesCredential(t *testing.T) {
	t.Parallel()

	manager := security.AuthenticationManagerFunc(func(ctx context.Context, credential security.Credential) (security.Authentication, error) {
		if err := ctx.Err(); err != nil {
			return security.Authentication{}, err
		}
		password, ok := credential.(security.PasswordCredential)
		if !ok || password.Username() != "alice" || password.Password() != "secret" {
			return security.Authentication{}, security.ErrBadCredentials
		}
		return security.NewAuthentication(
			security.PrincipalFunc(func() string { return password.Username() }),
			security.WithAuthorities("ROLE_USER"),
			security.WithAuthenticated(true),
		), nil
	})

	auth, err := manager.Authenticate(context.Background(), security.NewPasswordCredential("alice", "secret"))
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if !auth.Authenticated() || !auth.HasAuthority("ROLE_USER") {
		t.Fatalf("auth = %#v, want ROLE_USER", auth)
	}
	if _, err := manager.Authenticate(context.Background(), security.NewBearerCredential("bad")); !errors.Is(err, security.ErrBadCredentials) {
		t.Fatalf("bad credential err = %v, want ErrBadCredentials", err)
	}
}

func TestAuthorizerRequiresAuthority(t *testing.T) {
	t.Parallel()

	auth := security.NewAuthentication(
		security.PrincipalFunc(func() string { return "alice" }),
		security.WithAuthorities("orders:read"),
		security.WithAuthenticated(true),
	)
	authorizer := security.RequireAuthority("orders:read")
	decision, err := authorizer.Authorize(context.Background(), security.AuthorizationRequest{
		Authentication: auth,
		Resource:       "orders",
		Action:         "read",
	})
	if err != nil {
		t.Fatalf("Authorize failed: %v", err)
	}
	if !decision.Granted() {
		t.Fatalf("decision = %#v, want granted", decision)
	}

	decision, err = security.RequireAuthority("orders:write").Authorize(context.Background(), security.AuthorizationRequest{
		Authentication: auth,
		Resource:       "orders",
		Action:         "write",
	})
	if err != nil {
		t.Fatalf("Authorize write failed: %v", err)
	}
	if decision.Granted() || decision.Reason() == "" {
		t.Fatalf("decision = %#v, want denied with reason", decision)
	}
}
