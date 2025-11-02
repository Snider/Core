package config

import (
	"github.com/Snider/Core/pkg/config/internal"
	"github.com/Snider/Core/pkg/core"
)

// Options holds configuration for the config service.
type Options = internal.Options

// Service provides access to the application's configuration.
type Service = internal.Service

// New is the constructor for static dependency injection.
func New() (*Service, error) {
	return internal.New()
}

// Register is the constructor for dynamic dependency injection.
func Register(c *core.Core) (any, error) {
	return internal.Register(c)
}
