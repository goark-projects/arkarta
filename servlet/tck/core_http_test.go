package tck_test

import (
	"net/http"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/nethttp"
	"goark.dev/arkarta/servlet/session"
	"goark.dev/arkarta/servlet/tck"
)

func TestRunCoreHTTPWithNetHTTPAdapter(t *testing.T) {
	tck.RunCoreHTTP(t, func(handler servlet.Handler) http.Handler {
		return nethttp.Handler(handler)
	})
}

func TestRunSessionManagerWithMemoryManager(t *testing.T) {
	tck.RunSessionManager(t, func() session.Manager {
		return session.NewMemoryManager()
	})
}
