package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/SoufianoDev/OhMyFreeAI/internal/application/ports"
	domainconfig "github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
)

var ErrRepositoryRequired = errors.New("configuration repository is required")

type Service struct {
	Repository ports.ConfigRepository
}

func (s Service) Load(ctx context.Context) (domainconfig.Config, error) {
	if s.Repository == nil {
		return domainconfig.Config{}, ErrRepositoryRequired
	}
	cfg, err := s.Repository.Load(ctx)
	if err != nil {
		return domainconfig.Config{}, fmt.Errorf("load configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return domainconfig.Config{}, err
	}
	return cfg, nil
}

func (s Service) Save(ctx context.Context, cfg domainconfig.Config) error {
	if s.Repository == nil {
		return ErrRepositoryRequired
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := s.Repository.Save(ctx, cfg); err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}
	return nil
}
