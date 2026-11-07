package ports

import (
	"context"

	domainruntime "github.com/SoufianoDev/OhMyFreeAI/internal/domain/runtime"
)

type RuntimePolicy interface {
	Assessment(context.Context) (domainruntime.PolicyAssessment, error)
}

type RuntimeConnector interface {
	Connect(context.Context) (domainruntime.Session, error)
}
