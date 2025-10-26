// package crypt provides the public API for the Crypt service.
package crypt

import (
	// Import the internal implementation with an alias.
	impl "github.com/Snider/Core/pkg/crypt"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/pkg/core"
)

// Service is the public type for the Crypt service.
type Service = impl.Service

// New is the public constructor for the Crypt service.
var New = impl.New

// Crypt is the public interface for the Crypt service.
type Crypt = core.Crypt
