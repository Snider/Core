// package config provides the public API for the Config service.
package config

import (
	// Import the internal implementation with an alias.
	impl "github.com/Snider/Core/pkg/config"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/pkg/core"
)

// Service is the public type for the Config service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type Service = impl.Service

// Options is the public type for the Config service options.
type Options = impl.Options

// New is the public constructor for the Config service. It is a variable
// that points to the real constructor in the implementation package.
var New = impl.New

// Register is the public factory for the dynamic core.WithService pattern.
var Register = impl.Register

// Config is the public interface for the Config service.
type Config = core.Config
