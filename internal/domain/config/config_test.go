package config

import (
	"errors"
	"testing"
)

func TestDefaultsAreValidAndLoopbackOnly(t *testing.T) {
	c := Defaults()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.Gateway.Host != "127.0.0.1" || c.Gateway.Port != 18081 || c.Gateway.Enabled {
		t.Fatalf("unexpected safe Gateway defaults: %+v", c.Gateway)
	}
}

func TestRejectsNonLoopbackGateway(t *testing.T) {
	c := Defaults()
	c.Gateway.Host = "0.0.0.0"
	if err := c.Validate(); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}

func TestEnabledGatewayRequiresSecretReference(t *testing.T) {
	c := Defaults()
	c.Gateway.Enabled = true
	if err := c.Validate(); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
	c.Gateway.AuthSecretRef = "os-keychain://ohmyfreeai/gateway"
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}
