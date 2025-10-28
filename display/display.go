// package display provides the public API for the display service.
package display

import (
	// Import the internal implementation with an alias.
	impl "github.com/Snider/Core/pkg/display"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/pkg/core"
)

// Options is the public type for the Options service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Options = impl.Options

// Service is the public type for the Service service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Service = impl.Service

// WindowOption is the public type for the WindowOption service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type WindowOption = impl.WindowOption

// New is a public function that points to the real function in the implementation package.
var New = impl.New

// Register is a public function that points to the real function in the implementation package.
var Register = impl.Register

// Display is the public interface for the display service.
type Display = core.Display
