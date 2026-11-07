package compatibility

import "testing"

func TestReadyRequiresCompleteEvidence(t *testing.T) {
	a := Assessment{
		ModelRevision: "org/model@revision", ArtifactSHA256: "sha256:abc",
		EngineVersion: "llama.cpp@pinned", Backend: "cuda", RuntimeFingerprint: "gpu/t4-driver-x",
		Status: StatusSupported,
		Evidence: Evidence{ArtifactIntegrityVerified: true, EngineLoadSucceeded: true, HealthCheckPassed: true, InferenceSucceeded: true},
	}
	if !a.Ready() {
		t.Fatal("complete verified assessment should be ready")
	}
	a.Evidence.InferenceSucceeded = false
	if a.Ready() {
		t.Fatal("load without successful inference must not be ready")
	}
}

func TestUnverifiedIsNeverReady(t *testing.T) {
	a := Assessment{
		ModelRevision: "r1", ArtifactSHA256: "sha256:abc", EngineVersion: "v",
		Backend: "cpu", RuntimeFingerprint: "cpu", Status: StatusUnverified,
		Evidence: Evidence{true, true, true, true},
	}
	if a.Ready() {
		t.Fatal("unverified assessment must not be ready")
	}
}
