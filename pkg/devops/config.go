package devops

import (
	"os"
	"path/filepath"

	"github.com/host-uk/core/pkg/io"
	"gopkg.in/yaml.v3"
)

// Config holds global devops configuration from ~/.core/config.yaml.
type Config struct {
	Version int          `yaml:"version"`
	Images  ImagesConfig `yaml:"images"`
}

// ImagesConfig holds image source configuration.
type ImagesConfig struct {
	Source   string         `yaml:"source"` // auto, github, registry, cdn
	GitHub   GitHubConfig   `yaml:"github,omitempty"`
	Registry RegistryConfig `yaml:"registry,omitempty"`
	CDN      CDNConfig      `yaml:"cdn,omitempty"`
}

// GitHubConfig holds GitHub Releases configuration.
type GitHubConfig struct {
	Repo string `yaml:"repo"` // owner/repo format
}

// RegistryConfig holds container registry configuration.
type RegistryConfig struct {
	Image string `yaml:"image"` // e.g., ghcr.io/host-uk/core-devops
}

// CDNConfig holds CDN/S3 configuration.
type CDNConfig struct {
	URL string `yaml:"url"` // base URL for downloads
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

// LoadConfig loads configuration from ~/.core/config.yaml.
// Returns default config if file doesn't exist.
func LoadConfig() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	content, err := io.Local.Read(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(content), cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
