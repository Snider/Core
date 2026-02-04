package config

import (
	"context"

	core "github.com/host-uk/core/pkg/framework/core"
	"github.com/host-uk/core/pkg/io"
)

// Service wraps Config as a framework service with lifecycle support.
type Service struct {
	*core.ServiceRuntime[ServiceOptions]
	config *Config
}

// ServiceOptions holds configuration for the config service.
type ServiceOptions struct {
	// Path overrides the default config file path.
	Path string
	// Medium overrides the default storage medium.
	Medium io.Medium
}

// NewConfigService creates a new config service factory for the Core framework.
// Register it with core.WithService(config.NewConfigService).
func NewConfigService(c *core.Core) (any, error) {
	svc := &Service{
		ServiceRuntime: core.NewServiceRuntime(c, ServiceOptions{}),
	}
	return svc, nil
}

// OnStartup loads the configuration file during application startup.
func (s *Service) OnStartup(_ context.Context) error {
	opts := s.Opts()

	var configOpts []Option
	if opts.Path != "" {
		configOpts = append(configOpts, WithPath(opts.Path))
	}
	if opts.Medium != nil {
		configOpts = append(configOpts, WithMedium(opts.Medium))
	}

	cfg, err := New(configOpts...)
	if err != nil {
		return err
	}

	s.config = cfg
	return nil
}

// Get retrieves a configuration value by key.
func (s *Service) Get(key string, out any) error {
	if s.config == nil {
		return core.E("config.Service.Get", "config not loaded", nil)
	}
	return s.config.Get(key, out)
}

// Set stores a configuration value by key.
func (s *Service) Set(key string, v any) error {
	if s.config == nil {
		return core.E("config.Service.Set", "config not loaded", nil)
	}
	return s.config.Set(key, v)
}

// LoadFile merges a configuration file into the central configuration.
func (s *Service) LoadFile(m io.Medium, path string) error {
	if s.config == nil {
		return core.E("config.Service.LoadFile", "config not loaded", nil)
	}
	return s.config.LoadFile(m, path)
}

// Ensure Service implements core.Config and Startable at compile time.
var (
	_ core.Config    = (*Service)(nil)
	_ core.Startable = (*Service)(nil)
)
