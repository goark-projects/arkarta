package servlet

import (
	"context"
	"io"
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
		RequestURI:    "/orders?state=open",
		Path:          "/orders",
		QueryString:   "state=open",
		Header:        header,
		Body:          io.NopCloser(strings.NewReader("payload")),
		ContentLength: 7,
		RemoteAddr:    "192.0.2.10:50123",
		LocalAddr:     "192.0.2.20:8443",
	})
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if req.Method() != "POST" || req.Protocol() != "HTTP/1.1" || req.Path() != "/orders" {
		t.Fatalf("request metadata = %s/%s/%s", req.Method(), req.Protocol(), req.Path())
	}
	if req.Header().Get("Content-Type") != "text/plain" || req.LocalAddr() != "192.0.2.20:8443" {
		t.Fatalf("request transport data was not preserved")
	}
}

func TestNewRequestRejectsNilInput(t *testing.T) {
	if _, err := NewRequestFromInput(nil); err != ErrNilRequestInput {
		t.Fatalf("NewRequestFromInput error = %v, want ErrNilRequestInput", err)
	}
}
