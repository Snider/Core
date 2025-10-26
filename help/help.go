// package help provides the public API for the Help service.
package help

import (
	// Import the internal implementation with an alias.
	impl "github.com/Snider/Core/pkg/help"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/pkg/core"
)

// Service is the public type for the Help service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Service = impl.Service

// New is the public constructor for the Help service.
var New = impl.New

// Help is the public interface for the Help service.
type Help = core.Help
