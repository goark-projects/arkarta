package servlet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorPageRegistryHandlesStatusMapping(t *testing.T) {
	t.Parallel()

	registry := NewErrorPageRegistry()
	if err := registry.RegisterStatus(http.StatusNotFound, HandlerFunc(func(_ context.Context, req *Request, res Response) error {
		status, _ := req.Attribute(AttributeErrorStatusCode)
		if status != http.StatusNotFound {
			t.Fatalf("status attr = %v, want 404", status)
		}
		_, err := res.WriteString("not-found-page")
		return err
	})); err != nil {
		t.Fatalf("RegisterStatus failed: %v", err)
	}

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/missing", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	response := newTestResponse()

	handled, err := registry.Handle(context.Background(), req, response, http.StatusNotFound, nil)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if !handled {
		t.Fatal("status error page should be handled")
	}
	if response.Status() != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Status())
	}
	if response.body.String() != "not-found-page" {
		t.Fatalf("body = %q, want not-found-page", response.body.String())
	}
}

func TestErrorPageRegistryPrefersErrorTypeMapping(t *testing.T) {
	t.Parallel()

	registry := NewErrorPageRegistry()
	if err := registry.RegisterStatus(http.StatusInternalServerError, HandlerFunc(func(_ context.Context, _ *Request, res Response) error {
		_, err := res.WriteString("status-page")
		return err
	})); err != nil {
		t.Fatalf("RegisterStatus failed: %v", err)
	}
	if err := RegisterErrorType[*typedFailure](registry, HandlerFunc(func(_ context.Context, req *Request, res Response) error {
		errValue, _ := req.Attribute(AttributeErrorException)
		var failure *typedFailure
		if !errors.As(errValue.(error), &failure) || failure.code != "E_TYPED" {
			t.Fatalf("error attr = %v, want typed failure", errValue)
		}
		_, err := res.WriteString("type-page")
		return err
	})); err != nil {
		t.Fatalf("RegisterErrorType failed: %v", err)
	}

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/boom", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	response := newTestResponse()
	cause := errors.Join(&typedFailure{code: "E_TYPED"}, errors.New("wrapper"))

	handled, err := registry.Handle(context.Background(), req, response, http.StatusInternalServerError, cause)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if !handled || response.body.String() != "type-page" {
		t.Fatalf("handled/body = %v/%q, want true/type-page", handled, response.body.String())
	}
}

func TestErrorPageRegistryPrefersMostSpecificErrorTypeMapping(t *testing.T) {
	t.Parallel()

	registry := NewErrorPageRegistry()
	if err := RegisterErrorType[baseFailure](registry, HandlerFunc(func(_ context.Context, _ *Request, res Response) error {
		_, err := res.WriteString("base-page")
		return err
	})); err != nil {
		t.Fatalf("RegisterErrorType base failed: %v", err)
	}
	if err := RegisterErrorType[*specificFailure](registry, HandlerFunc(func(_ context.Context, _ *Request, res Response) error {
		_, err := res.WriteString("specific-page")
		return err
	})); err != nil {
		t.Fatalf("RegisterErrorType specific failed: %v", err)
	}

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/boom", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	response := newTestResponse()
	handled, err := registry.Handle(context.Background(), req, response, http.StatusInternalServerError, &specificFailure{})
	if err != nil || !handled {
		t.Fatalf("Handle handled/err = %v/%v, want true/nil", handled, err)
	}
	if response.body.String() != "specific-page" {
		t.Fatalf("body = %q, want specific-page", response.body.String())
	}
}

func TestErrorPageRegistryUsesDefaultAndPreventsLoop(t *testing.T) {
	t.Parallel()

	registry := NewErrorPageRegistry()
	if err := registry.RegisterDefault(HandlerFunc(func(ctx context.Context, req *Request, res Response) error {
		if handled, err := registry.Handle(ctx, req, res, http.StatusInternalServerError, errors.New("loop")); handled || !errors.Is(err, ErrErrorPageLoop) {
			t.Fatalf("loop handled/err = %v/%v, want false/ErrErrorPageLoop", handled, err)
		}
		_, err := res.WriteString("default-page")
		return err
	})); err != nil {
		t.Fatalf("RegisterDefault failed: %v", err)
	}
	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/boom", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	response := newTestResponse()

	handled, err := registry.Handle(context.Background(), req, response, http.StatusTeapot, nil)
	if err != nil || !handled {
		t.Fatalf("Handle default handled/err = %v/%v, want true/nil", handled, err)
	}
	if response.body.String() != "default-page" {
		t.Fatalf("body = %q, want default-page", response.body.String())
	}
}

func TestErrorPageRegistrySkipsCommittedResponse(t *testing.T) {
	t.Parallel()

	registry := NewErrorPageRegistry()
	if err := registry.RegisterStatus(http.StatusInternalServerError, HandlerFunc(func(context.Context, *Request, Response) error {
		t.Fatal("committed response should not dispatch error page")
		return nil
	})); err != nil {
		t.Fatalf("RegisterStatus failed: %v", err)
	}
	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/boom", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	response := newTestResponse()
	if _, err := response.WriteString("committed"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	handled, err := registry.Handle(context.Background(), req, response, http.StatusInternalServerError, errors.New("boom"))
	if !errors.Is(err, ErrResponseCommitted) {
		t.Fatalf("err = %v, want ErrResponseCommitted", err)
	}
	if handled {
		t.Fatal("committed response must not be handled")
	}
}

type typedFailure struct {
	code string
}

func (e *typedFailure) Error() string {
	return e.code
}

type baseFailure interface {
	error
	Base()
}

type specificFailure struct{}

func (e *specificFailure) Error() string {
	return "specific"
}

func (e *specificFailure) Base() {
}
