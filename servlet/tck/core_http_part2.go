package tck

import (
	"context"
	"net/http"
	"testing"

	"goark.dev/arkarta/servlet"
)

func runMappingElements(t *testing.T, driver Driver) {
	t.Helper()
	router := servlet.NewRouter()
	mustHandle(
		t,
		router,
		"/api/*",
		servlet.HandlerFunc(
			func(_ context.Context, req *servlet.Request, res servlet.Response) error {
				if req.Mapping().Type() != servlet.MappingPrefix {
					t.Fatalf("mapping type = %v, want prefix", req.Mapping().Type())
				}
				if req.ServletPath() != "/api" || req.PathInfo() != "/orders" {
					t.Fatalf(
						"mapping paths = %q/%q, want /api//orders",
						req.ServletPath(),
						req.PathInfo(),
					)
				}
				_, err := res.WriteString(req.Mapping().Pattern())
				return err
			},
		),
	)

	response := exchange(t, driver, router, NewRequest(http.MethodGet, "/api/orders"))

	if response.Status != http.StatusOK || string(response.Body) != "/api/*" {
		t.Fatalf("status/body = %d/%q, want 200//api/*", response.Status, response.Body)
	}
}

func markHandler(value string) servlet.Handler {
	return servlet.HandlerFunc(
		func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			_, err := res.WriteString(value)
			return err
		},
	)
}

func mustHandle(
	t *testing.T,
	router *servlet.Router,
	pattern string,
	handler servlet.Handler,
) {
	t.Helper()
	if err := router.Handle(pattern, handler); err != nil {
		t.Fatalf("Handle(%q) failed: %v", pattern, err)
	}
}
