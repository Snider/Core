// package i18n provides the public API for the I18n service.
package i18n

import (
	// Import the internal implementation with an alias.
	impl "github.com/Snider/Core/pkg/i18n"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/pkg/core"
)

// Service is the public type for the I18n service.
type Service = impl.Service

// New is the public factory for the core.WithService pattern.
var New = impl.New

// Register is the public factory for the dynamic core.WithService pattern.
var Register = impl.Register

// I18n is the public interface for the I18n service.
type I18n = core.I18n
