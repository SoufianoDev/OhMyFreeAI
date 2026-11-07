package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/SoufianoDev/OhMyFreeAI/internal/application/ports"
	domainconfig "github.com/SoufianoDev/OhMyFreeAI/internal/domain/config"
)

var ErrStoredConfigSchemaMismatch = errors.New("stored configuration schema version mismatch")

var _ ports.ConfigRepository = (*Store)(nil)

// Load returns the stored configuration, or application defaults when the
// settings row has not been created yet.
func (s *Store) Load(ctx context.Context) (domainconfig.Config, error) {
	if s == nil || s.db == nil {
		return domainconfig.Config{}, ErrStoreClosed
	}

	var storedVersion int
	var encoded string
	err := s.db.QueryRowContext(ctx,
		"SELECT schema_version, config_json FROM app_settings WHERE singleton_id = 1",
	).Scan(&storedVersion, &encoded)
	if errors.Is(err, sql.ErrNoRows) {
		return domainconfig.Defaults(), nil
	}
	if err != nil {
		return domainconfig.Config{}, fmt.Errorf("read application configuration: %w", err)
	}

	var cfg domainconfig.Config
	if err := json.Unmarshal([]byte(encoded), &cfg); err != nil {
		return domainconfig.Config{}, fmt.Errorf("decode stored application configuration: %w", err)
	}
	if cfg.SchemaVersion != storedVersion {
		return domainconfig.Config{}, fmt.Errorf("%w: row=%d document=%d", ErrStoredConfigSchemaMismatch, storedVersion, cfg.SchemaVersion)
	}
	return cfg, nil
}

// Save validates and atomically replaces the single persisted configuration.
// Secret values are never stored; only validated OS secret references exist in
// the configuration model.
func (s *Store) Save(ctx context.Context, cfg domainconfig.Config) error {
	if s == nil || s.db == nil {
		return ErrStoreClosed
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	encoded, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode application configuration: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO app_settings(singleton_id, schema_version, config_json, updated_at) VALUES (1, ?, ?, ?) ON CONFLICT(singleton_id) DO UPDATE SET schema_version=excluded.schema_version, config_json=excluded.config_json, updated_at=excluded.updated_at",
		cfg.SchemaVersion, string(encoded), time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("save application configuration: %w", err)
	}
	return nil
}
