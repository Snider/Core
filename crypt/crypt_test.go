package crypt_test

import (
	"testing"

	"github.com/Snider/Core/crypt"
	"github.com/Snider/Core/pkg/core"
)

// TestNew ensures that the public constructor New is available.
func TestNew(t *testing.T) {
	if crypt.New == nil {
		t.Fatal("crypt.New constructor is nil")
	}
	// Note: This is a basic check. Some services may require a core instance
	// or other arguments. This test can be expanded as needed.
}

// TestRegister ensures that the public factory Register is available.
func TestRegister(t *testing.T) {
	if crypt.Register == nil {
		t.Fatal("crypt.Register factory is nil")
	}
}

// TestInterfaceCompliance ensures that the public Service type correctly
// implements the public Crypt interface. This is a compile-time check.
func TestInterfaceCompliance(t *testing.T) {
	// This is a compile-time check. If it compiles, the test passes.
	var _ core.Crypt = (*crypt.Service)(nil)
}
