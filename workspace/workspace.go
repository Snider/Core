// package workspace provides the public API for the workspace service.
package workspace

import (
	// Import the internal implementation with an alias.
	impl "github.com/Snider/Core/pkg/workspace"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/pkg/core"
)

// Options is the public type for the Options service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Options = impl.Options

// Workspace is the public type for the Workspace service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Workspace = impl.Workspace

// Service is the public type for the Service service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Service = impl.Service

// New is a public function that points to the real function in the implementation package.
var New = impl.New

// Register is a public function that points to the real function in the implementation package.
var Register = impl.Register

// Workspace is the public interface for the workspace service.
type Workspace = core.Workspace
