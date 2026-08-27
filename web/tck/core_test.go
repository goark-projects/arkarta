package tck_test

import (
	"net/http"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/nethttp"
	"goark.dev/arkarta/web/tck"
)

func TestRunJSONValidationWithNetHTTPAdapter(t *testing.T) {
	tck.RunJSONValidation(t, func(handler servlet.Handler) http.Handler {
		return nethttp.Handler(handler)
	})
}

func TestRunRoutingBindingWithNetHTTPAdapter(t *testing.T) {
	tck.RunRoutingBinding(t, func(handler servlet.Handler) http.Handler {
		return nethttp.Handler(handler)
	})
}
