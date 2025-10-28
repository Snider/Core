// package config provides the public API for the config service.
package config

import (
	// Import the internal implementation with an alias.
	impl "github.com/Snider/Core/pkg/config"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/pkg/core"
)

// New is a public function that points to the real function in the implementation package.
var New = impl.New

// Register is a public function that points to the real function in the implementation package.
var Register = impl.Register

// Options is the public type for the Options service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Options = impl.Options

// Service is the public type for the Service service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Service = impl.Service

// Config is the public interface for the config service.
type Config = core.Config
