package runtime

import "time"

type PolicyDecision string

const (
	PolicyReviewRequired PolicyDecision = "review_required"
	PolicyApproved       PolicyDecision = "approved"
	PolicyBlocked        PolicyDecision = "blocked"
)

type PolicyAssessment struct {
	Provider       string
	Decision       PolicyDecision
	Scope          string
	SourceURL      string
	PolicyRevision string
	ReviewedAt     time.Time
}

// AllowsRemoteInference fails closed unless a scoped, sourced review is recorded.
func (p PolicyAssessment) AllowsRemoteInference() bool {
	return p.Provider != "" && p.Decision == PolicyApproved && p.Scope != "" &&
		p.SourceURL != "" && !p.ReviewedAt.IsZero()
}
