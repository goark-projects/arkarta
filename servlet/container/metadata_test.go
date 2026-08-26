package container

import "testing"

func TestMetadataUsesSnapshots(t *testing.T) {
	t.Parallel()

	profiles := []Profile{ProfileCore}
	limits := map[string]string{"maxHeaderBytes": "1048576"}
	metadata := NewMetadata("arkarta-nethttp", "1.0.0", profiles, limits)

	profiles[0] = ProfileSession
	limits["maxHeaderBytes"] = "1"

	if !metadata.Supports(ProfileCore) {
		t.Fatal("metadata should keep original core profile")
	}
	if metadata.Limits()["maxHeaderBytes"] != "1048576" {
		t.Fatalf("limit changed to %s", metadata.Limits()["maxHeaderBytes"])
	}
}
