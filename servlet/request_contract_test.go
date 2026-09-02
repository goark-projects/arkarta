package servlet

import (
	"context"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestRequestUsesTransportNeutralInput(t *testing.T) {
	header := NewHeader()
	header.Set("Content-Type", "text/plain")
	req, err := NewRequestFromInput(&RequestInput{
		Context:       context.Background(),
		Method:        "POST",
		Protocol:      "HTTP/1.1",
		Scheme:        "https",
		Host:          "example.test:8443",
		RequestURI:    "/app/orders?state=open",
		Path:          "/app/orders",
		QueryString:   "state=open",
		ContextPath:   "/app",
		Header:        header,
		Body:          io.NopCloser(strings.NewReader("payload")),
		ContentLength: 7,
		RemoteAddr:    "192.0.2.10:50123",
		LocalAddr:     "192.0.2.20:8443",
	})
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if req.Method() != "POST" || req.Protocol() != "HTTP/1.1" || req.Path() != "/orders" || req.ContextPath() != "/app" {
		t.Fatalf("request metadata = %s/%s/%s", req.Method(), req.Protocol(), req.Path())
	}
	if req.Header().Get("Content-Type") != "text/plain" || req.LocalAddr() != "192.0.2.20:8443" {
		t.Fatalf("request transport data was not preserved")
	}
}

func TestRequestInputParsesQueryAndFormParameters(t *testing.T) {
	header := NewHeader()
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	req, err := NewRequestFromInput(&RequestInput{
		Method:      "POST",
		QueryString: "q=query",
		Header:      header,
		Body:        io.NopCloser(strings.NewReader("q=form&body=ok")),
	})
	if err != nil {
		t.Fatalf("NewRequestFromInput failed: %v", err)
	}

	values, ok, err := req.ParameterValues("q")
	if err != nil {
		t.Fatalf("ParameterValues failed: %v", err)
	}
	if want := []string{"query", "form"}; !ok || !reflect.DeepEqual(values, want) {
		t.Fatalf("q values = %#v/%v, want %#v/true", values, ok, want)
	}
	if body, ok, err := req.Parameter("body"); err != nil || !ok || body != "ok" {
		t.Fatalf("body = %q/%v/%v, want ok/true/nil", body, ok, err)
	}
}

func TestNewRequestRejectsNilInput(t *testing.T) {
	if _, err := NewRequestFromInput(nil); err != ErrNilRequestInput {
		t.Fatalf("NewRequestFromInput error = %v, want ErrNilRequestInput", err)
	}
}
