package config

import (
	"context"
	"errors"
	"testing"

	domainconfig "github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
)

type configRepoStub struct {
	saved domainconfig.Config
	saves int
}

func (r *configRepoStub) Load(context.Context) (domainconfig.Config, error) {
	return domainconfig.Defaults(), nil
}

func (r *configRepoStub) Save(_ context.Context, cfg domainconfig.Config) error {
	r.saved = cfg
	r.saves++
	return nil
}

func TestSaveValidatesBeforePersistence(t *testing.T) {
	repo := &configRepoStub{}
	service := Service{Repository: repo}
	cfg := domainconfig.Defaults()
	cfg.Gateway.Host = "0.0.0.0"
	if err := service.Save(context.Background(), cfg); !errors.Is(err, domainconfig.ErrInvalidConfig) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if repo.saves != 0 {
		t.Fatal("invalid config must not reach repository")
	}
}

func TestSaveValidConfig(t *testing.T) {
	repo := &configRepoStub{}
	if err := (Service{Repository: repo}).Save(context.Background(), domainconfig.Defaults()); err != nil {
		t.Fatal(err)
	}
	if repo.saves != 1 {
		t.Fatalf("save calls = %d", repo.saves)
	}
}
