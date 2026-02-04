package plugin

import (
	"encoding/json"

	core "github.com/host-uk/core/pkg/framework/core"
	"github.com/host-uk/core/pkg/io"
)

// Manifest represents a plugin.json manifest file.
// Each plugin repository must contain a plugin.json at its root.
type Manifest struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Author       string   `json:"author"`
	Entrypoint   string   `json:"entrypoint"`
	Dependencies []string `json:"dependencies,omitempty"`
	MinVersion   string   `json:"min_version,omitempty"`
}

// LoadManifest reads and parses a plugin.json file from the given path.
func LoadManifest(m io.Medium, path string) (*Manifest, error) {
	content, err := m.Read(path)
	if err != nil {
		return nil, core.E("plugin.LoadManifest", "failed to read manifest", err)
	}

	var manifest Manifest
	if err := json.Unmarshal([]byte(content), &manifest); err != nil {
		return nil, core.E("plugin.LoadManifest", "failed to parse manifest JSON", err)
	}

	return &manifest, nil
}

// Validate checks the manifest for required fields.
// Returns an error if name, version, or entrypoint are missing.
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return core.E("plugin.Manifest.Validate", "name is required", nil)
	}
	if m.Version == "" {
		return core.E("plugin.Manifest.Validate", "version is required", nil)
	}
	if m.Entrypoint == "" {
		return core.E("plugin.Manifest.Validate", "entrypoint is required", nil)
	}
	return nil
}
