package config

import (
	"errors"
	"fmt"
	"net"
)

const CurrentSchemaVersion = 1

var ErrInvalidConfig = errors.New("invalid application configuration")

type Config struct {
	SchemaVersion int           `json:"schema_version"`
	Gateway       GatewayConfig `json:"gateway"`
	Engine        EngineConfig  `json:"engine"`
	Runtime       RuntimeConfig `json:"runtime"`
	Storage       StorageConfig `json:"storage"`
	UI            UIConfig      `json:"ui"`
}

type GatewayConfig struct {
	Enabled       bool   `json:"enabled"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	AuthSecretRef string `json:"auth_secret_ref,omitempty"`
}

type EngineConfig struct {
	ServerPort int `json:"server_port"`
}

type RuntimeConfig struct {
	Provider string `json:"provider"`
}

type StorageConfig struct {
	ArtifactDirectory string `json:"artifact_directory,omitempty"`
}

type UIConfig struct {
	Language string `json:"language"`
	Theme    string `json:"theme"`
}

func Defaults() Config {
	return Config{
		SchemaVersion: CurrentSchemaVersion,
		Gateway:       GatewayConfig{Enabled: false, Host: "127.0.0.1", Port: 18081},
		Engine:        EngineConfig{ServerPort: 8081},
		Runtime:       RuntimeConfig{Provider: "colab"},
		UI:            UIConfig{Language: "ar", Theme: "system"},
	}
}

func (c Config) Validate() error {
	if c.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("%w: unsupported schema version %d", ErrInvalidConfig, c.SchemaVersion)
	}
	ip := net.ParseIP(c.Gateway.Host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("%w: Gateway host must be a loopback IP address", ErrInvalidConfig)
	}
	if !validPort(c.Gateway.Port) || !validPort(c.Engine.ServerPort) {
		return fmt.Errorf("%w: ports must be between 1 and 65535", ErrInvalidConfig)
	}
	if c.Gateway.Enabled && c.Gateway.AuthSecretRef == "" {
		return fmt.Errorf("%w: enabled Gateway requires an OS secret reference", ErrInvalidConfig)
	}
	if c.Runtime.Provider != "" && c.Runtime.Provider != "colab" {
		return fmt.Errorf("%w: unsupported runtime provider %q", ErrInvalidConfig, c.Runtime.Provider)
	}
	if c.UI.Language != "ar" && c.UI.Language != "en" {
		return fmt.Errorf("%w: UI language must be ar or en", ErrInvalidConfig)
	}
	if c.UI.Theme != "light" && c.UI.Theme != "dark" && c.UI.Theme != "system" {
		return fmt.Errorf("%w: UI theme must be light, dark or system", ErrInvalidConfig)
	}
	return nil
}

func validPort(port int) bool { return port >= 1 && port <= 65535 }
