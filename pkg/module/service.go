package module

import (
	"context"

	"github.com/Snider/Core/pkg/core"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Options holds configuration for the module service.
type Options struct {
	AppsDir string // Directory to scan for .itw3.json files
}

// Service wraps Registry for Wails service registration.
type Service struct {
	*core.ServiceRuntime[Options]
	registry *Registry
	config   Options
}

// NewService creates a new module service.
func NewService(opts Options) (*Service, error) {
	reg := NewRegistry()
	if opts.AppsDir != "" {
		reg.SetAppsDir(opts.AppsDir)
	}
	return &Service{
		registry: reg,
		config:   opts,
	}, nil
}

// ServiceName returns the canonical name.
func (s *Service) ServiceName() string {
	return "github.com/Snider/Core/module"
}

// ServiceStartup is called by Wails on app start.
func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	// Load any apps from the apps directory
	return s.registry.LoadApps(ctx)
}

// Registry returns the underlying registry for direct access.
func (s *Service) Registry() *Registry {
	return s.registry
}

// --- Wails-bound methods (exposed to frontend) ---

// RegisterModule registers a module from JSON config string.
func (s *Service) RegisterModule(jsonConfig string) error {
	return s.registry.RegisterFromJSON([]byte(jsonConfig))
}

// UnregisterModule removes a module by code.
func (s *Service) UnregisterModule(code string) {
	s.registry.Unregister(code)
}

// SetContext changes the active UI context.
func (s *Service) SetContext(ctx string) {
	s.registry.SetContext(Context(ctx))
}

// GetContext returns the current context.
func (s *Service) GetContext() string {
	return string(s.registry.GetContext())
}

// GetModules returns all registered modules.
func (s *Service) GetModules() []Config {
	return s.registry.GetModules()
}

// GetMenus returns menus for the current context.
func (s *Service) GetMenus() []MenuItem {
	return s.registry.GetMenus()
}

// GetRoutes returns routes for the current context.
func (s *Service) GetRoutes() []Route {
	return s.registry.GetRoutes()
}

// GetUIConfig returns complete UI config for the current context.
func (s *Service) GetUIConfig() UIConfig {
	return s.registry.GetUIConfig()
}

// GetAvailableContexts returns all available contexts.
func (s *Service) GetAvailableContexts() []string {
	return []string{
		string(ContextDefault),
		string(ContextDeveloper),
		string(ContextRetail),
		string(ContextMiner),
	}
}

// GetModule returns a specific module by code.
func (s *Service) GetModule(code string) (Config, bool) {
	m, ok := s.registry.Get(code)
	if !ok {
		return Config{}, false
	}
	return m.Config, true
}
