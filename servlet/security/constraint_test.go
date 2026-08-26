package security

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
)

func TestConstraintAuthorizesRolesAndTransport(t *testing.T) {
	t.Parallel()

	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "https://example.com/admin", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	SetPrincipal(req, PrincipalFunc(func() string { return "alice" }), "BASIC", "admin")
	constraint := NewConstraint(
		WithRoles("admin"),
		WithTransportGuarantee(TransportConfidential),
	)

	if err := constraint.Authorize(context.Background(), req); err != nil {
		t.Fatalf("Authorize failed: %v", err)
	}
	if RemoteUser(req) != "alice" || AuthType(req) != "BASIC" || !UserInRole(req, "admin") {
		t.Fatalf("security context user=%q auth=%q role=%v", RemoteUser(req), AuthType(req), UserInRole(req, "admin"))
	}
}

func TestConstraintRejectsMissingAuthAndInsecureTransport(t *testing.T) {
	t.Parallel()

	insecure, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "http://example.com/admin", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	constraint := NewConstraint(WithRoles("admin"), WithTransportGuarantee(TransportConfidential))
	if err := constraint.Authorize(context.Background(), insecure); !statusIs(err, http.StatusForbidden) {
		t.Fatalf("insecure err = %v, want 403", err)
	}

	secure, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "https://example.com/admin", nil))
	if err != nil {
		t.Fatalf("NewRequest secure failed: %v", err)
	}
	constraint = NewConstraint(WithRoles("admin"))
	if err := constraint.Authorize(context.Background(), secure); !statusIs(err, http.StatusUnauthorized) {
		t.Fatalf("missing auth err = %v, want 401", err)
	}
}

func TestFilterShortCircuitsDeniedRequest(t *testing.T) {
	t.Parallel()

	req, err := servlet.NewRequest(httptest.NewRequest(http.MethodGet, "https://example.com/admin", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	filter := NewFilter(NewConstraint(WithEmptyRoleSemantic(EmptyRoleDeny)))
	chainCalled := false
	err = filter.Filter(context.Background(), req, nil, servlet.ChainFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		chainCalled = true
		return nil
	}))
	if !statusIs(err, http.StatusForbidden) {
		t.Fatalf("filter err = %v, want 403", err)
	}
	if chainCalled {
		t.Fatal("chain should not be called for denied request")
	}
}

func statusIs(err error, status int) bool {
	var statusErr servlet.StatusError
	return errors.As(err, &statusErr) && statusErr.StatusCode() == status
}
