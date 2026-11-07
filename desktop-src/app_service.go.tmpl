package main

import (
	"context"
	"fmt"
	"time"

	appconfig "github.com/SoufianoDev/OhMyFreeAI/internal/application/config"
	domainconfig "github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
)

type DashboardSnapshot struct {
	AppVersion    string                `json:"appVersion"`
	DatabaseState string                `json:"databaseState"`
	RuntimeState  string                `json:"runtimeState"`
	Config        domainconfig.Config   `json:"config"`
}

// AppService is the narrow Wails boundary. Domain and application packages do
// not import Wails types.
type AppService struct {
	configuration appconfig.Service
}

func NewAppService(configuration appconfig.Service) *AppService {
	return &AppService{configuration: configuration}
}

// GetSnapshot returns only verified local application state. It does not
// claim that a remote runtime, gateway or model is running.
func (s *AppService) GetSnapshot() (DashboardSnapshot, error) {
	if s == nil {
		return DashboardSnapshot{}, fmt.Errorf("application service is unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := s.configuration.Load(ctx)
	if err != nil {
		return DashboardSnapshot{}, fmt.Errorf("load application snapshot: %w", err)
	}
	return DashboardSnapshot{
		AppVersion:    "0.1.0",
		DatabaseState: "ready",
		RuntimeState:  "not_connected",
		Config:        cfg,
	}, nil
}

// SaveConfig persists validated application settings. Secret values are not
// accepted here; the configuration model contains secret references only.
func (s *AppService) SaveConfig(cfg domainconfig.Config) error {
	if s == nil {
		return fmt.Errorf("application service is unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.configuration.Save(ctx, cfg); err != nil {
		return fmt.Errorf("save application configuration: %w", err)
	}
	return nil
}
