package servlet

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestRequestPathElementsWithContextPath(t *testing.T) {
	t.Parallel()

	req, err := NewRequest(
		httptest.NewRequest(http.MethodGet, "https://example.com/app/orders/42?expand=items", nil),
		WithRequestContextPath("/app"),
	)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	if req.ContextPath() != "/app" {
		t.Fatalf("context path = %q, want /app", req.ContextPath())
	}
	if req.Path() != "/orders/42" {
		t.Fatalf("path = %q, want /orders/42", req.Path())
	}
	if req.RequestURI() != "/app/orders/42" {
		t.Fatalf("request uri = %q, want /app/orders/42", req.RequestURI())
	}
	if req.QueryString() != "expand=items" {
		t.Fatalf("query string = %q, want expand=items", req.QueryString())
	}
	if req.RequestURL() != "https://example.com/app/orders/42" {
		t.Fatalf("request url = %q, want https://example.com/app/orders/42", req.RequestURL())
	}
}

func TestRequestParametersMergeQueryAndFormBody(t *testing.T) {
	t.Parallel()

	body := strings.NewReader("a=body&b=3")
	httpRequest := httptest.NewRequest(http.MethodPost, "/submit?a=1&a=2", body)
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req, err := NewRequest(httpRequest)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	first, ok, err := req.Parameter("a")
	if err != nil {
		t.Fatalf("Parameter failed: %v", err)
	}
	if !ok || first != "1" {
		t.Fatalf("a first = %q/%v, want 1/true", first, ok)
	}
	values, ok, err := req.ParameterValues("a")
	if err != nil {
		t.Fatalf("ParameterValues failed: %v", err)
	}
	wantValues := []string{"1", "2", "body"}
	if !ok || !reflect.DeepEqual(values, wantValues) {
		t.Fatalf("a values = %#v/%v, want %#v/true", values, ok, wantValues)
	}
	params, err := req.Parameters()
	if err != nil {
		t.Fatalf("Parameters failed: %v", err)
	}
	want := url.Values{"a": {"1", "2", "body"}, "b": {"3"}}
	if !reflect.DeepEqual(params, want) {
		t.Fatalf("params = %#v, want %#v", params, want)
	}
	params.Set("a", "mutated")
	values, _, _ = req.ParameterValues("a")
	if reflect.DeepEqual(values, []string{"mutated"}) {
		t.Fatal("Parameters must return a defensive copy")
	}
}
