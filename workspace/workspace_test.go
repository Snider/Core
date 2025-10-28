package workspace_test

import (
	"testing"

	"github.com/Snider/Core/pkg/core"
	"github.com/Snider/Core/workspace"
)

// TestNew ensures that the public constructor New is available.
func TestNew(t *testing.T) {
	if workspace.New == nil {
		t.Fatal("workspace.New constructor is nil")
	}
	// Note: This is a basic check. Some services may require a core instance
	// or other arguments. This test can be expanded as needed.
}

// TestRegister ensures that the public factory Register is available.
func TestRegister(t *testing.T) {
	if workspace.Register == nil {
		t.Fatal("workspace.Register factory is nil")
	}
}

// TestInterfaceCompliance ensures that the public Service type correctly
// implements the public Workspace interface. This is a compile-time check.
func TestInterfaceCompliance(t *testing.T) {
	// This is a compile-time check. If it compiles, the test passes.
	var _ core.Workspace = (*workspace.Service)(nil)
}
