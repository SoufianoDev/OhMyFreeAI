package runtime

import (
	"testing"
	"time"
)

func TestPolicyAssessmentRequiresCleanHTTPSReference(t *testing.T) {
	base := PolicyAssessment{
		Provider: "colab", Decision: PolicyApproved, Scope: "remote inference",
		SourceURL:      "https://research.google.com/colaboratory/faq.html",
		PolicyRevision: "2026-11-07", ReviewedAt: time.Now(),
	}
	if !base.AllowsRemoteInference() {
		t.Fatal("valid sourced policy review should pass")
	}
	invalid := []string{
		"http://research.google.com/colaboratory/faq.html",
		"https://user:secret@research.google.com/colaboratory/faq.html",
		"https://research.google.com/colaboratory/faq.html?token=secret",
		"https://",
	}
	for _, source := range invalid {
		t.Run(source, func(t *testing.T) {
			assessment := base
			assessment.SourceURL = source
			if assessment.AllowsRemoteInference() {
				t.Fatalf("unsafe policy source accepted: %q", source)
			}
		})
	}
}
