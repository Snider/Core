// Package core provides the main runtime and plugin system for Core applications.
package core

import (
	"context"

	"github.com/host-uk/core/pkg/plugin"
	"github.com/host-uk/core/pkg/plugin/builtin/system"
	"github.com/host-uk/core/pkg/runtime"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Runtime wraps the internal runtime and plugin system.
type Runtime struct {
	*runtime.Runtime
	Plugins *plugin.Router
}

// NewRuntime creates a new Core runtime with plugin support.
func NewRuntime() (*Runtime, error) {
	// Create the base runtime
	rt, err := runtime.New()
	if err != nil {
		return nil, err
	}

	// Create the plugin router
	plugins := plugin.NewRouter()

	// Register built-in plugins
	ctx := context.Background()
	if err := plugins.Register(ctx, system.New()); err != nil {
		return nil, err
	}

	return &Runtime{
		Runtime: rt,
		Plugins: plugins,
	}, nil
}

// RegisterPlugin adds a plugin to the router.
func (r *Runtime) RegisterPlugin(p plugin.Plugin) error {
	return r.Plugins.Register(context.Background(), p)
}

// PluginServices returns Wails services for the plugin system.
func (r *Runtime) PluginServices() []application.Service {
	return []application.Service{
		application.NewService(r.Plugins),
	}
}
