// Package sdk provides OpenAPI SDK generation and diff capabilities.
package sdk

import (
	"context"
	"fmt"
)

// Config holds SDK generation configuration from .core/release.yaml.
type Config struct {
	// Spec is the path to the OpenAPI spec file (auto-detected if empty).
	Spec string `yaml:"spec,omitempty"`
	// Languages to generate SDKs for.
	Languages []string `yaml:"languages,omitempty"`
	// Output directory (default: sdk/).
	Output string `yaml:"output,omitempty"`
	// Package naming configuration.
	Package PackageConfig `yaml:"package,omitempty"`
	// Diff configuration for breaking change detection.
	Diff DiffConfig `yaml:"diff,omitempty"`
	// Publish configuration for monorepo publishing.
	Publish PublishConfig `yaml:"publish,omitempty"`
}

// PackageConfig holds package naming configuration.
type PackageConfig struct {
	// Name is the base package name.
	Name string `yaml:"name,omitempty"`
	// Version is the SDK version (supports templates like {{.Version}}).
	Version string `yaml:"version,omitempty"`
}

// DiffConfig holds breaking change detection configuration.
type DiffConfig struct {
	// Enabled determines whether to run diff checks.
	Enabled bool `yaml:"enabled,omitempty"`
	// FailOnBreaking fails the release if breaking changes are detected.
	FailOnBreaking bool `yaml:"fail_on_breaking,omitempty"`
}

// PublishConfig holds monorepo publishing configuration.
type PublishConfig struct {
	// Repo is the SDK monorepo (e.g., "myorg/sdks").
	Repo string `yaml:"repo,omitempty"`
	// Path is the subdirectory for this SDK (e.g., "packages/myapi").
	Path string `yaml:"path,omitempty"`
}

// SDK orchestrates OpenAPI SDK generation.
type SDK struct {
	config     *Config
	projectDir string
}

// New creates a new SDK instance.
func New(projectDir string, config *Config) *SDK {
	if config == nil {
		config = DefaultConfig()
	}
	return &SDK{
		config:     config,
		projectDir: projectDir,
	}
}

// DefaultConfig returns sensible defaults for SDK configuration.
func DefaultConfig() *Config {
	return &Config{
		Languages: []string{"typescript", "python", "go", "php"},
		Output:    "sdk",
		Diff: DiffConfig{
			Enabled:        true,
			FailOnBreaking: false,
		},
	}
}

// Generate generates SDKs for all configured languages.
func (s *SDK) Generate(ctx context.Context) error {
	return fmt.Errorf("sdk.Generate: not implemented")
}

// GenerateLanguage generates SDK for a specific language.
func (s *SDK) GenerateLanguage(ctx context.Context, lang string) error {
	return fmt.Errorf("sdk.GenerateLanguage: not implemented")
}
