package nethttp

import (
	"net/http"
	"net/textproto"
)

type headerAdapter struct {
	values http.Header
}

func (h *headerAdapter) Get(name string) string {
	return h.values.Get(name)
}

func (h *headerAdapter) Values(name string) []string {
	return h.values.Values(name)
}

func (h *headerAdapter) Has(name string) bool {
	_, ok := h.values[textproto.CanonicalMIMEHeaderKey(name)]
	return ok
}

func (h *headerAdapter) Set(name, value string) {
	h.values.Set(name, value)
}

func (h *headerAdapter) Add(name, value string) {
	h.values.Add(name, value)
}

func (h *headerAdapter) Delete(name string) {
	h.values.Del(name)
}

func (h *headerAdapter) Visit(visitor func(name, value string) bool) {
	if visitor == nil {
		return
	}
	for name, values := range h.values {
		for _, value := range values {
			if !visitor(name, value) {
				return
			}
		}
	}
}
