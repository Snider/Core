// Package plugin provides a plugin system for Core applications.
// Plugins can register HTTP handlers that get served alongside the main app.
package plugin

import (
	"context"
	"net/http"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Plugin defines the interface that all plugins must implement.
type Plugin interface {
	// Name returns the unique identifier for this plugin.
	// Used for routing: /api/{namespace}/{name}/...
	Name() string

	// Namespace returns the plugin's namespace (e.g., "core", "mining", "marketplace").
	// Plugins in the same namespace share configuration and can communicate.
	Namespace() string

	// ServeHTTP handles HTTP requests routed to this plugin.
	// The request path will have the /api/{namespace}/{name} prefix stripped.
	http.Handler

	// OnRegister is called when the plugin is registered with the router.
	OnRegister(ctx context.Context) error

	// OnUnregister is called when the plugin is being removed.
	OnUnregister(ctx context.Context) error
}

// PluginInfo contains metadata about a registered plugin.
type PluginInfo struct {
	Name        string
	Namespace   string
	Description string
	Version     string
	Author      string
	Routes      []string // List of sub-routes this plugin handles
}

// BasePlugin provides a default implementation of Plugin that can be embedded.
type BasePlugin struct {
	name        string
	namespace   string
	description string
	version     string
	handler     http.Handler
}

// NewBasePlugin creates a new BasePlugin with the given configuration.
func NewBasePlugin(namespace, name string, handler http.Handler) *BasePlugin {
	return &BasePlugin{
		name:      name,
		namespace: namespace,
		handler:   handler,
	}
}

func (p *BasePlugin) Name() string      { return p.name }
func (p *BasePlugin) Namespace() string { return p.namespace }

func (p *BasePlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.handler != nil {
		p.handler.ServeHTTP(w, r)
	} else {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}

func (p *BasePlugin) OnRegister(ctx context.Context) error   { return nil }
func (p *BasePlugin) OnUnregister(ctx context.Context) error { return nil }

// WithDescription sets the plugin description.
func (p *BasePlugin) WithDescription(desc string) *BasePlugin {
	p.description = desc
	return p
}

// WithVersion sets the plugin version.
func (p *BasePlugin) WithVersion(version string) *BasePlugin {
	p.version = version
	return p
}

// Info returns the plugin's metadata.
func (p *BasePlugin) Info() PluginInfo {
	return PluginInfo{
		Name:        p.name,
		Namespace:   p.namespace,
		Description: p.description,
		Version:     p.version,
	}
}

// ServiceOptions returns Wails service options for this plugin.
// This allows plugins to be registered directly as Wails services.
func ServiceOptionsForPlugin(p Plugin) application.ServiceOptions {
	return application.ServiceOptions{
		Route: "/api/" + p.Namespace() + "/" + p.Name(),
	}
}
