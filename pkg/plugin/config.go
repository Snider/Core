package plugin

// PluginConfig holds configuration for a single installed plugin.
type PluginConfig struct {
	Name        string `json:"name" yaml:"name"`
	Version     string `json:"version" yaml:"version"`
	Source      string `json:"source" yaml:"source"` // e.g., "github:org/repo"
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	InstalledAt string `json:"installed_at" yaml:"installed_at"` // RFC 3339 timestamp
}
