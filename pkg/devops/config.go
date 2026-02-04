package devops

import (
	"os"
	"path/filepath"

	"github.com/host-uk/core/pkg/config"
	"github.com/host-uk/core/pkg/io"
)

// Config holds global devops configuration from ~/.core/config.yaml.
type Config struct {
	Version int          `yaml:"version" mapstructure:"version"`
	Images  ImagesConfig `yaml:"images" mapstructure:"images"`
}

// ImagesConfig holds image source configuration.
type ImagesConfig struct {
	Source   string         `yaml:"source" mapstructure:"source"` // auto, github, registry, cdn
	GitHub   GitHubConfig   `yaml:"github,omitempty" mapstructure:"github,omitempty"`
	Registry RegistryConfig `yaml:"registry,omitempty" mapstructure:"registry,omitempty"`
	CDN      CDNConfig      `yaml:"cdn,omitempty" mapstructure:"cdn,omitempty"`
}

// GitHubConfig holds GitHub Releases configuration.
type GitHubConfig struct {
	Repo string `yaml:"repo" mapstructure:"repo"` // owner/repo format
}

// RegistryConfig holds container registry configuration.
type RegistryConfig struct {
	Image string `yaml:"image" mapstructure:"image"` // e.g., ghcr.io/host-uk/core-devops
}

// CDNConfig holds CDN/S3 configuration.
type CDNConfig struct {
	URL string `yaml:"url" mapstructure:"url"` // base URL for downloads
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Images: ImagesConfig{
			Source: "auto",
			GitHub: GitHubConfig{
				Repo: "host-uk/core-images",
			},
			Registry: RegistryConfig{
				Image: "ghcr.io/host-uk/core-devops",
			},
		},
	}
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".core", "config.yaml"), nil
}

// LoadConfig loads configuration from ~/.core/config.yaml using the provided medium.
// Returns default config if file doesn't exist.
func LoadConfig(m io.Medium) (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	cfg := DefaultConfig()

	if !m.IsFile(configPath) {
		return cfg, nil
	}

	// Use centralized config service
	c, err := config.New(config.WithMedium(m), config.WithPath(configPath))
	if err != nil {
		return nil, err
	}

	if err := c.Get("", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
