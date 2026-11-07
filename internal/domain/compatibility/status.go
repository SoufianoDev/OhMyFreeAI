package compatibility

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
	EngineLoadSucceeded      bool
	HealthCheckPassed        bool
	InferenceSucceeded       bool
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
	return a.ModelRevision != "" && a.ArtifactSHA256 != "" && a.EngineVersion != "" &&
		a.Backend != "" && a.RuntimeFingerprint != "" &&
		(a.Status == StatusSupported || a.Status == StatusSupportedWithLimits) &&
		a.Evidence.ArtifactIntegrityVerified && a.Evidence.EngineLoadSucceeded &&
		a.Evidence.HealthCheckPassed && a.Evidence.InferenceSucceeded
}
