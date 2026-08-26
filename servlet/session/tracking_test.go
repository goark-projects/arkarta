package session

import (
	"errors"
	"reflect"
	"testing"
)

func TestTrackingPolicyModes(t *testing.T) {
	t.Parallel()

	policy, err := NewTrackingPolicy(TrackingURL, TrackingCookie)
	if err != nil {
		t.Fatalf("NewTrackingPolicy failed: %v", err)
	}
	if !policy.Allows(TrackingCookie) || !policy.Allows(TrackingURL) || policy.Allows(TrackingSSL) {
		t.Fatalf("policy allows unexpected modes")
	}
	want := []TrackingMode{TrackingCookie, TrackingURL}
	if !reflect.DeepEqual(policy.Modes(), want) {
		t.Fatalf("modes = %#v, want %#v", policy.Modes(), want)
	}
	if _, err := NewTrackingPolicy(TrackingMode("BAD")); !errors.Is(err, ErrInvalidTrackingMode) {
		t.Fatalf("invalid mode err = %v, want ErrInvalidTrackingMode", err)
	}
}
