// Package sdk provides OpenAPI SDK generation and diff capabilities.
package sdk

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/host-uk/core/pkg/sdk/generators"
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
	version    string
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

// SetVersion sets the SDK version for generation.
// This updates both the internal version field and the config's Package.Version.
func (s *SDK) SetVersion(version string) {
	s.version = version
	if s.config != nil {
		s.config.Package.Version = version
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
	// Generate for each language
	for _, lang := range s.config.Languages {
		if err := s.GenerateLanguage(ctx, lang); err != nil {
			return err
		}
	}

	return nil
}

// GenerateLanguage generates SDK for a specific language.
func (s *SDK) GenerateLanguage(ctx context.Context, lang string) error {
	specPath, err := s.DetectSpec()
	if err != nil {
		return err
	}

	registry := generators.NewRegistry()
	registry.Register(generators.NewTypeScriptGenerator())
	registry.Register(generators.NewPythonGenerator())
	registry.Register(generators.NewGoGenerator())
	registry.Register(generators.NewPHPGenerator())

	gen, ok := registry.Get(lang)
	if !ok {
		return fmt.Errorf("sdk.GenerateLanguage: unknown language: %s", lang)
	}

	if !gen.Available() {
		fmt.Printf("Warning: %s generator not available. Install with: %s\n", lang, gen.Install())
		fmt.Printf("Falling back to Docker...\n")
	}

	outputDir := filepath.Join(s.projectDir, s.config.Output, lang)
	opts := generators.Options{
		SpecPath:    specPath,
		OutputDir:   outputDir,
		PackageName: s.config.Package.Name,
		Version:     s.config.Package.Version,
	}

	fmt.Printf("Generating %s SDK...\n", lang)
	if err := gen.Generate(ctx, opts); err != nil {
		return fmt.Errorf("sdk.GenerateLanguage: %s generation failed: %w", lang, err)
	}
	fmt.Printf("Generated %s SDK at %s\n", lang, outputDir)

	return nil
}
