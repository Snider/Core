// Package build provides project type detection and cross-compilation for the Core build system.
// This file handles configuration loading from .core/build.yaml files.
package build

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigFileName is the name of the build configuration file.
const ConfigFileName = "build.yaml"

// ConfigDir is the directory where build configuration is stored.
const ConfigDir = ".core"

// BuildConfig holds the complete build configuration loaded from .core/build.yaml.
// This is distinct from Config which holds runtime build parameters.
type BuildConfig struct {
	// Version is the config file format version.
	Version int `yaml:"version"`
	// Project contains project metadata.
	Project Project `yaml:"project"`
	// Build contains build settings.
	Build Build `yaml:"build"`
	// Targets defines the build targets.
	Targets []TargetConfig `yaml:"targets"`
}

// Project holds project metadata.
type Project struct {
	// Name is the project name.
	Name string `yaml:"name"`
	// Description is a brief description of the project.
	Description string `yaml:"description"`
	// Main is the path to the main package (e.g., ./cmd/core).
	Main string `yaml:"main"`
	// Binary is the output binary name.
	Binary string `yaml:"binary"`
}

// Build holds build-time settings.
type Build struct {
	// CGO enables CGO for the build.
	CGO bool `yaml:"cgo"`
	// Flags are additional build flags (e.g., ["-trimpath"]).
	Flags []string `yaml:"flags"`
	// LDFlags are linker flags (e.g., ["-s", "-w"]).
	LDFlags []string `yaml:"ldflags"`
	// Env are additional environment variables.
	Env []string `yaml:"env"`
}

// TargetConfig defines a build target in the config file.
// This is separate from Target to allow for additional config-specific fields.
type TargetConfig struct {
	// OS is the target operating system (e.g., "linux", "darwin", "windows").
	OS string `yaml:"os"`
	// Arch is the target architecture (e.g., "amd64", "arm64").
	Arch string `yaml:"arch"`
}

// LoadConfig loads build configuration from the .core/build.yaml file in the given directory.
// If the config file does not exist, it returns DefaultConfig().
// Returns an error if the file exists but cannot be parsed.
func LoadConfig(dir string) (*BuildConfig, error) {
	configPath := filepath.Join(dir, ConfigDir, ConfigFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("build.LoadConfig: failed to read config file: %w", err)
	}

	var cfg BuildConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("build.LoadConfig: failed to parse config file: %w", err)
	}

	// Apply defaults for any missing fields
	applyDefaults(&cfg)

	return &cfg, nil
}

// DefaultConfig returns sensible defaults for Go projects.
func DefaultConfig() *BuildConfig {
	return &BuildConfig{
		Version: 1,
		Project: Project{
			Name:   "",
			Main:   ".",
			Binary: "",
		},
		Build: Build{
			CGO:     false,
			Flags:   []string{"-trimpath"},
			LDFlags: []string{"-s", "-w"},
			Env:     []string{},
		},
		Targets: []TargetConfig{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
			{OS: "darwin", Arch: "amd64"},
			{OS: "darwin", Arch: "arm64"},
			{OS: "windows", Arch: "amd64"},
		},
	}
}

// applyDefaults fills in default values for any empty fields in the config.
func applyDefaults(cfg *BuildConfig) {
	defaults := DefaultConfig()

	if cfg.Version == 0 {
		cfg.Version = defaults.Version
	}

	if cfg.Project.Main == "" {
		cfg.Project.Main = defaults.Project.Main
	}

	if cfg.Build.Flags == nil {
		cfg.Build.Flags = defaults.Build.Flags
	}

	if cfg.Build.LDFlags == nil {
		cfg.Build.LDFlags = defaults.Build.LDFlags
	}

	if cfg.Build.Env == nil {
		cfg.Build.Env = defaults.Build.Env
	}

	if len(cfg.Targets) == 0 {
		cfg.Targets = defaults.Targets
	}
}

// ConfigPath returns the path to the build config file for a given directory.
func ConfigPath(dir string) string {
	return filepath.Join(dir, ConfigDir, ConfigFileName)
}

// ConfigExists checks if a build config file exists in the given directory.
func ConfigExists(dir string) bool {
	return fileExists(ConfigPath(dir))
}

// ToTargets converts TargetConfig slice to Target slice for use with builders.
func (cfg *BuildConfig) ToTargets() []Target {
	targets := make([]Target, len(cfg.Targets))
	for i, t := range cfg.Targets {
		targets[i] = Target{OS: t.OS, Arch: t.Arch}
	}
	return targets
}
