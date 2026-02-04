// Package plugin provides a plugin system for the core CLI.
//
// Plugins extend the CLI with additional commands and functionality.
// They are distributed as GitHub repositories and managed via a local registry.
//
// Plugin lifecycle:
//   - Install: Download from GitHub, validate manifest, register
//   - Init: Parse manifest and prepare plugin
//   - Start: Activate plugin functionality
//   - Stop: Deactivate and clean up
//   - Remove: Unregister and delete files
package plugin

import "context"

// Plugin is the interface that all plugins must implement.
type Plugin interface {
	// Name returns the plugin's unique identifier.
	Name() string

	// Version returns the plugin's semantic version.
	Version() string

	// Init prepares the plugin for use.
	Init(ctx context.Context) error

	// Start activates the plugin.
	Start(ctx context.Context) error

	// Stop deactivates the plugin and releases resources.
	Stop(ctx context.Context) error
}

// BasePlugin provides a default implementation of Plugin.
// Embed this in concrete plugin types to inherit default behaviour.
type BasePlugin struct {
	PluginName    string
	PluginVersion string
}

// Name returns the plugin name.
func (p *BasePlugin) Name() string { return p.PluginName }

// Version returns the plugin version.
func (p *BasePlugin) Version() string { return p.PluginVersion }

// Init is a no-op default implementation.
func (p *BasePlugin) Init(_ context.Context) error { return nil }

// Start is a no-op default implementation.
func (p *BasePlugin) Start(_ context.Context) error { return nil }

// Stop is a no-op default implementation.
func (p *BasePlugin) Stop(_ context.Context) error { return nil }
