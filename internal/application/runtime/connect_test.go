package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	domainruntime "github.com/SoufianoDev/OhMyFreeAI/internal/domain/runtime"
)

type policyStub struct {
	assessment domainruntime.PolicyAssessment
}

func (p policyStub) Assessment(context.Context) (domainruntime.PolicyAssessment, error) {
	return p.assessment, nil
}

type connectorStub struct{ calls int }

func (c *connectorStub) Connect(context.Context) (domainruntime.Session, error) {
	c.calls++
	return domainruntime.NewSession("test", 1, time.Now()), nil
}

func TestConnectFailsClosedWithoutPolicyApproval(t *testing.T) {
	connector := &connectorStub{}
	service := ConnectService{Policy: policyStub{assessment: domainruntime.PolicyAssessment{Provider: "colab", Decision: domainruntime.PolicyReviewRequired}}, Connector: connector}
	if _, err := service.Connect(context.Background()); !errors.Is(err, ErrPolicyNotApproved) {
		t.Fatalf("expected policy gate error, got %v", err)
	}
	if connector.calls != 0 {
		t.Fatal("connector must not run before policy approval")
	}
}

func TestConnectRequiresScopedSourcedApproval(t *testing.T) {
	connector := &connectorStub{}
	assessment := domainruntime.PolicyAssessment{Provider: "provider", Decision: domainruntime.PolicyApproved, Scope: "interactive inference flow", SourceURL: "https://example.invalid/policy", PolicyRevision: "2026-11-07", ReviewedAt: time.Now()}
	service := ConnectService{Policy: policyStub{assessment: assessment}, Connector: connector}
	if _, err := service.Connect(context.Background()); err != nil {
		t.Fatal(err)
	}
	if connector.calls != 1 {
		t.Fatalf("connector calls = %d", connector.calls)
	}
}
