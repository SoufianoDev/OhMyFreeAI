package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/SoufianoDev/OhMyFreeAI/internal/application/ports"
	domainruntime "github.com/SoufianoDev/OhMyFreeAI/internal/domain/runtime"
)

var ErrPolicyNotApproved = errors.New("remote inference policy has not been approved")

type ConnectService struct {
	Policy    ports.RuntimePolicy
	Connector ports.RuntimeConnector
}

func (s ConnectService) Connect(ctx context.Context) (domainruntime.Session, error) {
	if s.Policy == nil || s.Connector == nil {
		return domainruntime.Session{}, errors.New("runtime policy and connector ports are required")
	}
	assessment, err := s.Policy.Assessment(ctx)
	if err != nil {
		return domainruntime.Session{}, fmt.Errorf("read runtime policy assessment: %w", err)
	}
	if !assessment.AllowsRemoteInference() {
		return domainruntime.Session{}, ErrPolicyNotApproved
	}
	return s.Connector.Connect(ctx)
}
