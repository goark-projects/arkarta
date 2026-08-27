package servlet

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

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
