// Package release provides release automation with changelog generation and publishing.
package release

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigFileName is the name of the release configuration file.
const ConfigFileName = "release.yaml"

// ConfigDir is the directory where release configuration is stored.
const ConfigDir = ".core"

// Config holds the complete release configuration loaded from .core/release.yaml.
type Config struct {
	// Version is the config file format version.
	Version int `yaml:"version"`
	// Project contains project metadata.
	Project ProjectConfig `yaml:"project"`
	// Build contains build settings for the release.
	Build BuildConfig `yaml:"build"`
	// Publishers defines where to publish the release.
	Publishers []PublisherConfig `yaml:"publishers"`
	// Changelog configures changelog generation.
	Changelog ChangelogConfig `yaml:"changelog"`

	// Internal fields (not serialized)
	projectDir string // Set by LoadConfig
	version    string // Set by CLI flag
}

// ProjectConfig holds project metadata for releases.
type ProjectConfig struct {
	// Name is the project name.
	Name string `yaml:"name"`
	// Repository is the GitHub repository in owner/repo format.
	Repository string `yaml:"repository"`
}

// BuildConfig holds build settings for releases.
type BuildConfig struct {
	// Targets defines the build targets.
	Targets []TargetConfig `yaml:"targets"`
}

// TargetConfig defines a build target.
type TargetConfig struct {
	// OS is the target operating system (e.g., "linux", "darwin", "windows").
	OS string `yaml:"os"`
	// Arch is the target architecture (e.g., "amd64", "arm64").
	Arch string `yaml:"arch"`
}

// PublisherConfig holds configuration for a publisher.
type PublisherConfig struct {
	// Type is the publisher type (e.g., "github").
	Type string `yaml:"type"`
	// Prerelease marks the release as a prerelease.
	Prerelease bool `yaml:"prerelease"`
	// Draft creates the release as a draft.
	Draft bool `yaml:"draft"`
}

// ChangelogConfig holds changelog generation settings.
type ChangelogConfig struct {
	// Include specifies commit types to include in the changelog.
	Include []string `yaml:"include"`
	// Exclude specifies commit types to exclude from the changelog.
	Exclude []string `yaml:"exclude"`
}

// LoadConfig loads release configuration from the .core/release.yaml file in the given directory.
// If the config file does not exist, it returns DefaultConfig().
// Returns an error if the file exists but cannot be parsed.
func LoadConfig(dir string) (*Config, error) {
	configPath := filepath.Join(dir, ConfigDir, ConfigFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			cfg.projectDir = dir
			return cfg, nil
		}
		return nil, fmt.Errorf("release.LoadConfig: failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("release.LoadConfig: failed to parse config file: %w", err)
	}

	// Apply defaults for any missing fields
	applyDefaults(&cfg)
	cfg.projectDir = dir

	return &cfg, nil
}

// DefaultConfig returns sensible defaults for release configuration.
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Project: ProjectConfig{
			Name:       "",
			Repository: "",
		},
		Build: BuildConfig{
			Targets: []TargetConfig{
				{OS: "linux", Arch: "amd64"},
				{OS: "linux", Arch: "arm64"},
				{OS: "darwin", Arch: "amd64"},
				{OS: "darwin", Arch: "arm64"},
				{OS: "windows", Arch: "amd64"},
			},
		},
		Publishers: []PublisherConfig{
			{
				Type:       "github",
				Prerelease: false,
				Draft:      false,
			},
		},
		Changelog: ChangelogConfig{
			Include: []string{"feat", "fix", "perf", "refactor"},
			Exclude: []string{"chore", "docs", "style", "test", "ci"},
		},
	}
}

// applyDefaults fills in default values for any empty fields in the config.
func applyDefaults(cfg *Config) {
	defaults := DefaultConfig()

	if cfg.Version == 0 {
		cfg.Version = defaults.Version
	}

	if len(cfg.Build.Targets) == 0 {
		cfg.Build.Targets = defaults.Build.Targets
	}

	if len(cfg.Publishers) == 0 {
		cfg.Publishers = defaults.Publishers
	}

	if len(cfg.Changelog.Include) == 0 && len(cfg.Changelog.Exclude) == 0 {
		cfg.Changelog.Include = defaults.Changelog.Include
		cfg.Changelog.Exclude = defaults.Changelog.Exclude
	}
}

// SetProjectDir sets the project directory on the config.
func (c *Config) SetProjectDir(dir string) {
	c.projectDir = dir
}

// SetVersion sets the version override on the config.
func (c *Config) SetVersion(version string) {
	c.version = version
}

// ConfigPath returns the path to the release config file for a given directory.
func ConfigPath(dir string) string {
	return filepath.Join(dir, ConfigDir, ConfigFileName)
}

// ConfigExists checks if a release config file exists in the given directory.
func ConfigExists(dir string) bool {
	_, err := os.Stat(ConfigPath(dir))
	return err == nil
}

// GetRepository returns the repository from the config.
func (c *Config) GetRepository() string {
	return c.Project.Repository
}

// GetProjectName returns the project name from the config.
func (c *Config) GetProjectName() string {
	return c.Project.Name
}

// WriteConfig writes the config to the .core/release.yaml file.
func WriteConfig(cfg *Config, dir string) error {
	configPath := ConfigPath(dir)

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("release.WriteConfig: failed to create directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("release.WriteConfig: failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("release.WriteConfig: failed to write config file: %w", err)
	}

	return nil
}
