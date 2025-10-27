// package workspace provides the public API for the Workspace service.
package workspace

import (
	// Import the internal implementation with an alias.
	impl "github.com/Snider/Core/pkg/workspace"
)

// Service is the public type for the Workspace service.
type Service = impl.Service

// New is the public factory for the core.WithService pattern.
var New = impl.New

// Register is the public factory for the dynamic core.WithService pattern.
var Register = impl.Register

// Workspace is the public interface for the Workspace service.
type Workspace = impl.Workspace
