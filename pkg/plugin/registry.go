package plugin

import (
	"encoding/json"
	"path/filepath"
	"sort"

	core "github.com/host-uk/core/pkg/framework/core"
	"github.com/host-uk/core/pkg/io"
)

const registryFilename = "registry.json"

// Registry manages installed plugins.
// Plugin metadata is stored in a registry.json file under the base path.
type Registry struct {
	medium   io.Medium
	basePath string // e.g., ~/.core/plugins/
	plugins  map[string]*PluginConfig
}

// NewRegistry creates a new plugin registry.
func NewRegistry(m io.Medium, basePath string) *Registry {
	return &Registry{
		medium:   m,
		basePath: basePath,
		plugins:  make(map[string]*PluginConfig),
	}
}

// List returns all installed plugins sorted by name.
func (r *Registry) List() []*PluginConfig {
	result := make([]*PluginConfig, 0, len(r.plugins))
	for _, cfg := range r.plugins {
		result = append(result, cfg)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Get returns a plugin by name.
// The second return value indicates whether the plugin was found.
func (r *Registry) Get(name string) (*PluginConfig, bool) {
	cfg, ok := r.plugins[name]
	return cfg, ok
}

// Add registers a plugin in the registry.
func (r *Registry) Add(cfg *PluginConfig) error {
	if cfg.Name == "" {
		return core.E("plugin.Registry.Add", "plugin name is required", nil)
	}
	r.plugins[cfg.Name] = cfg
	return nil
}

// Remove unregisters a plugin from the registry.
func (r *Registry) Remove(name string) error {
	if _, ok := r.plugins[name]; !ok {
		return core.E("plugin.Registry.Remove", "plugin not found: "+name, nil)
	}
	delete(r.plugins, name)
	return nil
}

// registryPath returns the full path to the registry file.
func (r *Registry) registryPath() string {
	return filepath.Join(r.basePath, registryFilename)
}

// Load reads the plugin registry from disk.
// If the registry file does not exist, the registry starts empty.
func (r *Registry) Load() error {
	path := r.registryPath()

	if !r.medium.IsFile(path) {
		// No registry file yet; start with empty registry
		r.plugins = make(map[string]*PluginConfig)
		return nil
	}

	content, err := r.medium.Read(path)
	if err != nil {
		return core.E("plugin.Registry.Load", "failed to read registry", err)
	}

	var plugins map[string]*PluginConfig
	if err := json.Unmarshal([]byte(content), &plugins); err != nil {
		return core.E("plugin.Registry.Load", "failed to parse registry", err)
	}

	if plugins == nil {
		plugins = make(map[string]*PluginConfig)
	}
	r.plugins = plugins
	return nil
}

// Save writes the plugin registry to disk.
func (r *Registry) Save() error {
	if err := r.medium.EnsureDir(r.basePath); err != nil {
		return core.E("plugin.Registry.Save", "failed to create plugin directory", err)
	}

	data, err := json.MarshalIndent(r.plugins, "", "  ")
	if err != nil {
		return core.E("plugin.Registry.Save", "failed to marshal registry", err)
	}

	if err := r.medium.Write(r.registryPath(), string(data)); err != nil {
		return core.E("plugin.Registry.Save", "failed to write registry", err)
	}

	return nil
}
