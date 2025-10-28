package config_test

import (
	"testing"

	"github.com/Snider/Core/config"
	"github.com/Snider/Core/pkg/core"
)

func TestInterfaceCompliance(t *testing.T) {
	var _ config.Config = (*config.Service)(nil)
}

func TestRegister(t *testing.T) {
	if config.Register == nil {
		t.Fatal("config.Register factory is nil")
	}
}

// TestGet_NonExistentKey validates that getting a non-existent key returns an error.
func TestGet_NonExistentKey(t *testing.T) {
	coreImpl, err := core.New(core.WithService(config.Register))
	if err != nil {
		t.Fatalf("core.New() failed: %v", err)
	}

	var value string
	err = coreImpl.Config().Get("nonexistent.key", &value)
	if err == nil {
		t.Fatal("expected an error when getting a nonexistent key, but got nil")
	}
}

// TestSetAndGet verifies that a value can be set and then retrieved correctly.
func TestSetAndGet(t *testing.T) {
	coreImpl, err := core.New(core.WithService(config.Register))
	if err != nil {
		t.Fatalf("core.New() failed: %v", err)
	}

	cfg := coreImpl.Config()

	// 1. Set a value for an existing key
	key := "language"
	expectedValue := "fr"
	err = cfg.Set(key, expectedValue)
	if err != nil {
		t.Fatalf("Set(%q, %q) failed: %v", key, expectedValue, err)
	}

	// 2. Get the value back
	var actualValue string
	err = cfg.Get(key, &actualValue)
	if err != nil {
		t.Fatalf("Get(%q) failed: %v", key, err)
	}

	// 3. Compare the values
	if actualValue != expectedValue {
		t.Errorf("Get(%q) returned %q, want %q", key, actualValue, expectedValue)
	}
}
