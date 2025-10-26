// package docs provides the public API for the Docs service.
package docs

import (
	// Import the internal implementation with an alias.
	"github.com/Snider/Core/pkg/docs"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/pkg/core"
)

// Service is the public type for the Docs service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Service = docs.Service

// New is the public constructor for the Docs service. It is a variable
// that points to the real constructor in the implementation package.
var New = docs.New

// Docs is the public interface for the Docs service.
type Docs = core.Docs
