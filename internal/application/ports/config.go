package ports

import (
	"context"

	"github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
)

type ConfigRepository interface {
	Load(context.Context) (config.Config, error)
	Save(context.Context, config.Config) error
}
