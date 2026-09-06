package servlet

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
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

func TestRequestParameterNamesAreSorted(t *testing.T) {
	t.Parallel()

	httpRequest := httptest.NewRequest(http.MethodPost, "/submit?b=1&a=query", strings.NewReader("c=3&a=form"))
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req, err := NewRequest(httpRequest)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	names, err := req.ParameterNames()
	if err != nil {
		t.Fatalf("ParameterNames failed: %v", err)
	}
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("names = %#v, want %#v", names, want)
	}
}

func TestRequestCookiesReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	httpRequest := httptest.NewRequest(http.MethodGet, "/profile", nil)
	httpRequest.AddCookie(&http.Cookie{Name: "sid", Value: "abc"})
	httpRequest.AddCookie(&http.Cookie{Name: "theme", Value: "dark"})
	req, err := NewRequest(httpRequest)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	cookies := req.Cookies()
	if len(cookies) != 2 || cookies[0].Name != "sid" || cookies[1].Name != "theme" {
		t.Fatalf("cookies = %#v", cookies)
	}
	cookies[0].Value = "mutated"
	sid, err := req.Cookie("sid")
	if err != nil {
		t.Fatalf("Cookie failed: %v", err)
	}
	if sid.Value != "abc" {
		t.Fatalf("sid = %q, want abc", sid.Value)
	}
}

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

func TestRequestHTTPMetadata(t *testing.T) {
	t.Parallel()

	httpRequest := httptest.NewRequest(http.MethodPost, "https://example.com:8443/app/orders", nil)
	httpRequest = httpRequest.WithContext(context.WithValue(
		httpRequest.Context(),
		http.LocalAddrContextKey,
		&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9443},
	))
	httpRequest.RemoteAddr = "192.0.2.10:53111"
	httpRequest.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpRequest.Header.Set("Accept-Language", "en-US;q=0.8, zh-CN;q=0.9")
	httpRequest.Header.Set("X-Count", "42")
	modified := time.Date(2026, time.August, 26, 9, 30, 0, 0, time.UTC)
	httpRequest.Header.Set("If-Modified-Since", modified.Format(http.TimeFormat))
	httpRequest.Trailer = http.Header{"X-Trailer": nil}

	req, err := NewRequest(httpRequest, WithRequestContextPath("/app"), WithRequestConnectionID("conn-1"))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	if req.ContentType() != "application/json" || req.CharacterEncoding() != "utf-8" {
		t.Fatalf("content metadata = %q/%q", req.ContentType(), req.CharacterEncoding())
	}
	if req.ServerName() != "example.com" || req.ServerPort() != 8443 {
		t.Fatalf("server = %s:%d, want example.com:8443", req.ServerName(), req.ServerPort())
	}
	if req.RemoteHost() != "192.0.2.10" || req.RemotePort() != 53111 {
		t.Fatalf("remote = %s:%d", req.RemoteHost(), req.RemotePort())
	}
	if req.LocalName() != "127.0.0.1" || req.LocalPort() != 9443 {
		t.Fatalf("local = %s:%d", req.LocalName(), req.LocalPort())
	}
	count, ok, err := req.IntHeader("X-Count")
	if err != nil || !ok || count != 42 {
		t.Fatalf("int header = %d/%v/%v, want 42/true/nil", count, ok, err)
	}
	date, ok, err := req.DateHeader("If-Modified-Since")
	if err != nil || !ok || !date.Equal(modified) {
		t.Fatalf("date header = %s/%v/%v, want modified", date, ok, err)
	}
	locale, ok := req.Locale()
	if !ok || locale.Tag() != "zh-CN" {
		t.Fatalf("locale = %s/%v, want zh-CN/true", locale.Tag(), ok)
	}
	if req.TrailerFieldsReady() {
		t.Fatal("trailer must not be ready while declared values are nil")
	}
	httpRequest.Trailer.Set("X-Trailer", "done")
	if !req.TrailerFieldsReady() || req.Trailer().Get("X-Trailer") != "done" {
		t.Fatalf("trailer = %v ready=%v", req.Trailer(), req.TrailerFieldsReady())
	}
	connection := req.ConnectionInfo()
	if connection.ID() != "conn-1" || connection.Protocol() != "HTTP/1.1" || !connection.Secure() {
		t.Fatalf("connection = %#v", connection)
	}
}
