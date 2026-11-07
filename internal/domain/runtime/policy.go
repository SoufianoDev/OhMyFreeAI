package runtime

import (
	"net/url"
	"time"
)

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
		p.PolicyRevision != "" && validPolicySource(p.SourceURL) && !p.ReviewedAt.IsZero()
}

func validPolicySource(source string) bool {
	u, err := url.Parse(source)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" &&
		u.User == nil && u.RawQuery == "" && u.Fragment == ""
}
