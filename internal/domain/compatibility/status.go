package compatibility

import "encoding/hex"

// Status is an evidence-backed assessment result for an exact model/runtime tuple.
type Status string

const (
	StatusUnknown             Status = "unknown"
	StatusSupported           Status = "supported"
	StatusSupportedWithLimits Status = "supported_with_limits"
	StatusUnsupported         Status = "unsupported"
	StatusUnverified          Status = "unverified"
	StatusResourceIneligible  Status = "resource_ineligible"
	StatusPolicyBlocked       Status = "policy_blocked"
)

type Evidence struct {
	ArtifactIntegrityVerified bool
	EngineLoadSucceeded       bool
	HealthCheckPassed         bool
	InferenceSucceeded        bool
}

type Assessment struct {
	ModelRevision      string
	ArtifactSHA256     string
	EngineVersion      string
	Backend            string
	RuntimeFingerprint string
	Status             Status
	Evidence           Evidence
}

// Ready requires exact tuple identity and successful end-to-end execution evidence.
func (a Assessment) Ready() bool {
	return a.ModelRevision != "" && validSHA256(a.ArtifactSHA256) && a.EngineVersion != "" &&
		a.Backend != "" && a.RuntimeFingerprint != "" &&
		(a.Status == StatusSupported || a.Status == StatusSupportedWithLimits) &&
		a.Evidence.ArtifactIntegrityVerified && a.Evidence.EngineLoadSucceeded &&
		a.Evidence.HealthCheckPassed && a.Evidence.InferenceSucceeded
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32
}
