package servlet

import (
	"reflect"
	"sort"
	"testing"
)

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
