package nethttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"goark.dev/arkarta/servlet"
)

func TestResponseResetBeforeCommit(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.Header().Set("X-Trace", "1")
	response.SetStatus(http.StatusAccepted)

	if err := response.Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
	if got := response.Header().Get("X-Trace"); got != "" {
		t.Fatalf("header after reset = %q", got)
	}
	if response.Status() != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Status(), http.StatusOK)
	}
}

func TestResponseResetAfterCommitFails(t *testing.T) {
	t.Parallel()

	response := NewResponse(httptest.NewRecorder())
	if _, err := response.WriteString("body"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	if err := response.Reset(); !errors.Is(err, servlet.ErrResponseCommitted) {
		t.Fatalf("Reset err = %v, want ErrResponseCommitted", err)
	}
}

func TestResponseNormalizesInvalidStatus(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	response := NewResponse(recorder)
	response.SetStatus(42)
	if _, err := response.WriteString("body"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestResponseBufferControl(t *testing.T) {
	t.Parallel()

	response := NewResponse(httptest.NewRecorder())
	if err := servlet.SetBufferSize(response, 8192); err != nil {
		t.Fatalf("SetBufferSize failed: %v", err)
	}
	if servlet.BufferSize(response) != 8192 {
		t.Fatalf("buffer size = %d, want 8192", servlet.BufferSize(response))
	}
	if err := servlet.ResetBuffer(response); err != nil {
		t.Fatalf("ResetBuffer failed: %v", err)
	}
}
