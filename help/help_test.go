package help_test

import (
	"testing"

	"github.com/Snider/Core/help"
	"github.com/Snider/Core/pkg/core"
)

// TestNew ensures that the public constructor New is available.
func TestNew(t *testing.T) {
	if help.New == nil {
		t.Fatal("help.New constructor is nil")
	}
	// Note: This is a basic check. Some services may require a core instance
	// or other arguments. This test can be expanded as needed.
}

// TestRegister ensures that the public factory Register is available.
func TestRegister(t *testing.T) {
	if help.Register == nil {
		t.Fatal("help.Register factory is nil")
	}
}

// TestInterfaceCompliance ensures that the public Service type correctly
// implements the public Help interface. This is a compile-time check.
func TestInterfaceCompliance(t *testing.T) {
	// This is a compile-time check. If it compiles, the test passes.
	var _ core.Help = (*help.Service)(nil)
}
