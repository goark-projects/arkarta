package servlet

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseAcceptSortsByQualityAndSpecificity(t *testing.T) {
	t.Parallel()

	accepted := ParseAccept("application/xml;q=0.4, text/*;q=0.7, application/json; charset=utf-8, */*;q=0.1")
	got := mediaTexts(accepted)
	want := []string{"application/json; charset=utf-8", "text/*", "application/xml", "*/*"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("accepted = %#v, want %#v", got, want)
	}
	if accepted[0].Quality() != 1 || accepted[1].Quality() != 0.7 {
		t.Fatalf("quality = %v/%v, want 1/0.7", accepted[0].Quality(), accepted[1].Quality())
	}
	if charset, ok := accepted[0].Parameter("charset"); !ok || charset != "utf-8" {
		t.Fatalf("charset = %q/%v, want utf-8/true", charset, ok)
	}
}

func TestMediaTypeMatchesWildcardsAndSuffix(t *testing.T) {
	t.Parallel()

	text, _ := NewMediaType("text/*")
	if !text.Matches("text/plain; charset=utf-8") || text.Matches("application/json") {
		t.Fatal("text/* should match text/plain only")
	}
	json, _ := NewMediaType("application/*+json")
	if !json.Matches("application/problem+json") || json.Matches("application/json") {
		t.Fatal("application/*+json should match structured JSON suffix only")
	}
}

func TestNegotiateContentTypeUsesMostSpecificQuality(t *testing.T) {
	t.Parallel()

	accepted := ParseAccept("application/json;q=0, */*;q=0.8")
	if got, ok := NegotiateContentType(accepted, "application/json", "text/plain"); !ok || got != "text/plain" {
		t.Fatalf("negotiated = %q/%v, want text/plain/true", got, ok)
	}

	accepted = ParseAccept("text/*;q=0.5, application/json;q=0.9")
	if got, ok := NegotiateContentType(accepted, "text/plain", "application/json"); !ok || got != "application/json" {
		t.Fatalf("negotiated = %q/%v, want application/json/true", got, ok)
	}
}

func TestRequestNegotiateContentType(t *testing.T) {
	t.Parallel()

	httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	httpRequest.Header.Set("Accept", "application/xml;q=0.6, application/json")
	req, err := NewRequest(httpRequest)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	if got, ok := req.NegotiateContentType("application/xml", "application/json"); !ok || got != "application/json" {
		t.Fatalf("negotiated = %q/%v, want application/json/true", got, ok)
	}
}

func mediaTexts(values []MediaType) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result
}
