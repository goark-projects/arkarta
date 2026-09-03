package servlet

import (
	"context"
	"errors"
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

func TestRequestInputHonorsConfiguredFormBodyLimit(t *testing.T) {
	header := NewHeader()
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	req, err := NewRequestFromInput(&RequestInput{
		Method: "POST",
		Header: header,
		Body:   io.NopCloser(strings.NewReader("field=value")),
	}, WithMaxFormBodySize(4))
	if err != nil {
		t.Fatalf("NewRequestFromInput failed: %v", err)
	}
	if err := req.ParseParameters(); !errors.Is(err, ErrFormBodyTooLarge) {
		t.Fatalf("ParseParameters err = %v, want ErrFormBodyTooLarge", err)
	}
}

func TestRequestInputAllowsUnlimitedFormBody(t *testing.T) {
	header := NewHeader()
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	req, err := NewRequestFromInput(&RequestInput{
		Method: "POST",
		Header: header,
		Body:   io.NopCloser(strings.NewReader("name=goark")),
	}, WithMaxFormBodySize(-1))
	if err != nil {
		t.Fatalf("NewRequestFromInput failed: %v", err)
	}
	if req.maxFormBodySize != -1 {
		t.Fatalf("form body limit = %d, want -1", req.maxFormBodySize)
	}
	value, ok, err := req.Parameter("name")
	if err != nil || !ok || value != "goark" {
		t.Fatalf("name = %q/%v/%v, want goark/true/nil", value, ok, err)
	}
}

func TestRequestSupportsTransportNeutralFilterOverrides(t *testing.T) {
	req, err := NewRequestFromInput(&RequestInput{
		Method:        "POST",
		Scheme:        "http",
		Host:          "internal:8080",
		Body:          io.NopCloser(strings.NewReader("before")),
		ContentLength: 6,
		RemoteAddr:    "10.0.0.1:5000",
	})
	if err != nil {
		t.Fatalf("NewRequestFromInput failed: %v", err)
	}
	req.SetMethod("DELETE")
	req.SetScheme("https")
	req.SetHost("api.example.com")
	req.SetRemoteAddr("192.0.2.10")
	req.SetBody(io.NopCloser(strings.NewReader("after")), 5)

	if req.Method() != "DELETE" || req.Scheme() != "https" || req.Host() != "api.example.com" {
		t.Fatalf("request override = %s/%s/%s", req.Method(), req.Scheme(), req.Host())
	}
	if req.RemoteAddr() != "192.0.2.10" || req.ContentLength() != 5 {
		t.Fatalf("network/body metadata = %s/%d", req.RemoteAddr(), req.ContentLength())
	}
	body, err := io.ReadAll(req.Body())
	if err != nil || string(body) != "after" {
		t.Fatalf("body = %q/%v, want after/nil", body, err)
	}
}

func TestNewRequestRejectsNilInput(t *testing.T) {
	if _, err := NewRequestFromInput(nil); err != ErrNilRequestInput {
		t.Fatalf("NewRequestFromInput error = %v, want ErrNilRequestInput", err)
	}
}
