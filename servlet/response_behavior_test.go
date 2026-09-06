package servlet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestResponseHelpersSetHeaders(t *testing.T) {
	t.Parallel()

	res := newTestResponse()
	if err := SetContentType(res, "application/json"); err != nil {
		t.Fatalf("SetContentType failed: %v", err)
	}
	if err := SetCharacterEncoding(res, "utf-8"); err != nil {
		t.Fatalf("SetCharacterEncoding failed: %v", err)
	}
	if err := SetContentLength(res, 12); err != nil {
		t.Fatalf("SetContentLength failed: %v", err)
	}
	if err := AddCookie(res, &Cookie{Name: "sid", Value: "abc", HTTPOnly: true}); err != nil {
		t.Fatalf("AddCookie failed: %v", err)
	}

	if ContentType(res) != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q, want application/json charset", ContentType(res))
	}
	if CharacterEncoding(res) != "utf-8" {
		t.Fatalf("charset = %q, want utf-8", CharacterEncoding(res))
	}
	if res.Header().Get("Content-Length") != "12" {
		t.Fatalf("content length = %q, want 12", res.Header().Get("Content-Length"))
	}
	if got := res.Header().Get("Set-Cookie"); got != "sid=abc; HttpOnly" {
		t.Fatalf("set-cookie = %q, want sid cookie", got)
	}
}

func TestRedirectResetsResponseAndSetsLocation(t *testing.T) {
	t.Parallel()

	res := newTestResponse()
	res.Header().Set("X-Old", "true")
	res.SetStatus(http.StatusCreated)

	if err := Redirect(res, "/login", http.StatusSeeOther); err != nil {
		t.Fatalf("Redirect failed: %v", err)
	}
	if res.Status() != http.StatusSeeOther {
		t.Fatalf("status = %d, want see other", res.Status())
	}
	if res.Header().Get("Location") != "/login" {
		t.Fatalf("location = %q, want /login", res.Header().Get("Location"))
	}
	if res.Header().Get("X-Old") != "" {
		t.Fatalf("old header leaked: %q", res.Header().Get("X-Old"))
	}
}

func TestSendErrorWritesSafeBody(t *testing.T) {
	t.Parallel()

	res := newTestResponse()
	if err := SendError(res, http.StatusForbidden, "denied"); err != nil {
		t.Fatalf("SendError failed: %v", err)
	}
	if res.Status() != http.StatusForbidden {
		t.Fatalf("status = %d, want forbidden", res.Status())
	}
	if res.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Fatalf("content type = %q, want text/plain utf-8", res.Header().Get("Content-Type"))
	}
	if res.body.String() != "denied\n" {
		t.Fatalf("body = %q, want denied newline", res.body.String())
	}
}

func TestResponseHelpersRejectCommittedResponse(t *testing.T) {
	t.Parallel()

	res := newTestResponse()
	if _, err := res.WriteString("committed"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	if err := Redirect(res, "/login", http.StatusFound); !errors.Is(err, ErrResponseCommitted) {
		t.Fatalf("Redirect err = %v, want ErrResponseCommitted", err)
	}
	if err := AddCookie(res, &Cookie{Name: "sid", Value: "abc"}); !errors.Is(err, ErrResponseCommitted) {
		t.Fatalf("AddCookie err = %v, want ErrResponseCommitted", err)
	}
}

func TestResponseTypedHeadersAndLocale(t *testing.T) {
	t.Parallel()

	res := newTestResponse()
	if err := SetHeader(res, "X-Mode", "set"); err != nil {
		t.Fatalf("SetHeader failed: %v", err)
	}
	if err := AddHeader(res, "X-Mode", "add"); err != nil {
		t.Fatalf("AddHeader failed: %v", err)
	}
	instant := time.Date(2026, time.August, 26, 10, 0, 0, 0, time.UTC)
	if err := SetDateHeader(res, "Last-Modified", instant); err != nil {
		t.Fatalf("SetDateHeader failed: %v", err)
	}
	if err := SetIntHeader(res, "X-Count", 7); err != nil {
		t.Fatalf("SetIntHeader failed: %v", err)
	}
	locale, _ := NewLocale("zh-cn")
	if err := SetLocale(res, locale); err != nil {
		t.Fatalf("SetLocale failed: %v", err)
	}

	values := HeaderValues(res, "X-Mode")
	if len(values) != 2 || values[0] != "set" || values[1] != "add" {
		t.Fatalf("X-Mode values = %#v", values)
	}
	if !ContainsHeader(res, "last-modified") {
		t.Fatal("Last-Modified header should exist")
	}
	if res.Header().Get("X-Count") != "7" {
		t.Fatalf("X-Count = %q, want 7", res.Header().Get("X-Count"))
	}
	gotLocale, ok := ResponseLocale(res)
	if !ok || gotLocale.Tag() != "zh-CN" {
		t.Fatalf("response locale = %s/%v, want zh-CN/true", gotLocale.Tag(), ok)
	}
}

func TestHeaderUsesCaseInsensitiveNamesAndPreservesValues(t *testing.T) {
	header := NewHeader()
	header.Add("content-type", "application/json")
	header.Add("X-Trace-ID", "trace-1")
	header.Add("x-trace-id", "trace-2")

	if got := header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	if got := header.Values("X-TRACE-ID"); !reflect.DeepEqual(got, []string{"trace-1", "trace-2"}) {
		t.Fatalf("X-Trace-ID values = %#v", got)
	}
}

func TestHeaderVisitStopsWhenVisitorReturnsFalse(t *testing.T) {
	header := NewHeader()
	header.Set("X-First", "1")
	header.Set("X-Second", "2")

	visited := 0
	header.Visit(func(string, string) bool {
		visited++
		return false
	})
	if visited != 1 {
		t.Fatalf("visited = %d, want 1", visited)
	}
}

func TestCloneHeaderIsIndependent(t *testing.T) {
	source := NewHeader()
	source.Add("Set-Cookie", "first=1")
	source.Add("Set-Cookie", "second=2")

	cloned := CloneHeader(source)
	cloned.Set("Set-Cookie", "replacement=3")

	if got := source.Values("Set-Cookie"); !reflect.DeepEqual(got, []string{"first=1", "second=2"}) {
		t.Fatalf("source values = %#v", got)
	}
	if got := cloned.Values("Set-Cookie"); !reflect.DeepEqual(got, []string{"replacement=3"}) {
		t.Fatalf("cloned values = %#v", got)
	}
}

func TestHeaderNamesReturnsStableCanonicalNames(t *testing.T) {
	header := NewHeader()
	header.Set("x-zeta", "z")
	header.Set("x-alpha", "a")
	header.Add("X-Alpha", "b")

	names := HeaderNames(header)
	sort.Strings(names)
	if !reflect.DeepEqual(names, []string{"X-Alpha", "X-Zeta"}) {
		t.Fatalf("names = %#v", names)
	}
}

func TestRouterUsesServletMappingPriority(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	mustHandle(t, router, "/", appendName("default"))
	mustHandle(t, router, "*.json", appendName("extension"))
	mustHandle(t, router, "/api/*", appendName("prefix"))
	mustHandle(t, router, "/api/users/me", appendName("exact"))

	tests := []struct {
		path string
		want string
	}{
		{path: "/api/users/me", want: "exact"},
		{path: "/api/orders", want: "prefix"},
		{path: "/report.json", want: "extension"},
		{path: "/assets/style.css", want: "default"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			req, err := NewRequest(httptest.NewRequest(http.MethodGet, tt.path, nil))
			if err != nil {
				t.Fatalf("NewRequest failed: %v", err)
			}
			if err := router.Serve(context.Background(), req, nil); err != nil {
				t.Fatalf("Serve failed: %v", err)
			}
			value, ok := req.Attribute("handler")
			if !ok || value != tt.want {
				t.Fatalf("handler = %v/%v, want %s/true", value, ok, tt.want)
			}
		})
	}
}

func TestRouterAppliesServletMappingElements(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	mustHandle(t, router, "/orders/*", HandlerFunc(func(_ context.Context, req *Request, _ Response) error {
		if req.Mapping().Type() != MappingPrefix {
			t.Fatalf("mapping type = %v, want prefix", req.Mapping().Type())
		}
		if req.Mapping().Pattern() != "/orders/*" {
			t.Fatalf("mapping pattern = %q, want /orders/*", req.Mapping().Pattern())
		}
		if req.ServletPath() != "/orders" {
			t.Fatalf("servlet path = %q, want /orders", req.ServletPath())
		}
		if req.PathInfo() != "/42" {
			t.Fatalf("path info = %q, want /42", req.PathInfo())
		}
		return nil
	}))

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/orders/42", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if err := router.Serve(context.Background(), req, nil); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}
}

func TestRouterReturnsNotFoundWithoutDefaultMapping(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/missing", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	err = router.Serve(context.Background(), req, nil)
	var statusErr StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("err = %v, want StatusError", err)
	}
	if statusErr.StatusCode() != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", statusErr.StatusCode())
	}
}

func TestRouterRejectsDuplicateMapping(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	mustHandle(t, router, "/users", appendName("first"))
	err := router.Handle("/users", appendName("second"))
	if !errors.Is(err, ErrDuplicateMapping) {
		t.Fatalf("err = %v, want ErrDuplicateMapping", err)
	}
}

func mustHandle(t *testing.T, router *Router, pattern string, handler Handler) {
	t.Helper()
	if err := router.Handle(pattern, handler); err != nil {
		t.Fatalf("Handle(%q) failed: %v", pattern, err)
	}
}

func appendName(name string) Handler {
	return HandlerFunc(func(_ context.Context, req *Request, _ Response) error {
		req.SetAttribute("handler", name)
		return nil
	})
}
