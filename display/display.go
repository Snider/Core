// package display provides the public API for the Display service.
package display

import (
	// Import the internal implementation with an alias.
	"github.com/Snider/Core/pkg/display"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/pkg/core"
)

// Service is the public type for the Display service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Service = display.Service

// New is the public constructor for the Display service. It is a variable
// that points to the real constructor in the implementation package.
var New = display.New

// Display is the public interface for the Display service.
type Display = core.Display
