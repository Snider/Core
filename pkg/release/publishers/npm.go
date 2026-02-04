// Package publishers provides release publishing implementations.
package publishers

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/host-uk/core/pkg/io"
)

//go:embed templates/npm/*.tmpl
var npmTemplates embed.FS

// NpmConfig holds npm-specific configuration.
type NpmConfig struct {
	// Package is the npm package name (e.g., "@host-uk/core").
	Package string
	// Access is the npm access level: "public" or "restricted".
	Access string
}

// NpmPublisher publishes releases to npm using the binary wrapper pattern.
type NpmPublisher struct{}

// NewNpmPublisher creates a new npm publisher.
func NewNpmPublisher() *NpmPublisher {
	return &NpmPublisher{}
}

// Name returns the publisher's identifier.
func (p *NpmPublisher) Name() string {
	return "npm"
}

// Publish publishes the release to npm.
// It generates a binary wrapper package that downloads the correct platform binary on postinstall.
func (p *NpmPublisher) Publish(ctx context.Context, release *Release, pubCfg PublisherConfig, relCfg ReleaseConfig, dryRun bool) error {
	// Parse npm config
	npmCfg := p.parseConfig(pubCfg, relCfg)

	// Validate configuration
	if npmCfg.Package == "" {
		return fmt.Errorf("npm.Publish: package name is required (set publish.npm.package in config)")
	}

	// Get repository
	repo := ""
	if relCfg != nil {
		repo = relCfg.GetRepository()
	}
	if repo == "" {
		detectedRepo, err := detectRepository(release.ProjectDir)
		if err != nil {
			return fmt.Errorf("npm.Publish: could not determine repository: %w", err)
		}
		repo = detectedRepo
	}

	// Get project name (binary name)
	projectName := ""
	if relCfg != nil {
		projectName = relCfg.GetProjectName()
	}
	if projectName == "" {
		// Try to infer from package name
		parts := strings.Split(npmCfg.Package, "/")
		projectName = parts[len(parts)-1]
	}

	// Strip leading 'v' from version for npm
	version := strings.TrimPrefix(release.Version, "v")

	// Template data
	data := npmTemplateData{
		Package:     npmCfg.Package,
		Version:     version,
		Description: fmt.Sprintf("%s CLI", projectName),
		License:     "MIT",
		Repository:  repo,
		BinaryName:  projectName,
		ProjectName: projectName,
		Access:      npmCfg.Access,
	}

	if dryRun {
		return p.dryRunPublish(release.FS, data, &npmCfg)
	}

	return p.executePublish(ctx, release.FS, data, &npmCfg)
}

// parseConfig extracts npm-specific configuration from the publisher config.
func (p *NpmPublisher) parseConfig(pubCfg PublisherConfig, relCfg ReleaseConfig) NpmConfig {
	cfg := NpmConfig{
		Package: "",
		Access:  "public",
	}

	// Override from extended config if present
	if ext, ok := pubCfg.Extended.(map[string]any); ok {
		if pkg, ok := ext["package"].(string); ok && pkg != "" {
			cfg.Package = pkg
		}
		if access, ok := ext["access"].(string); ok && access != "" {
			cfg.Access = access
		}
	}

	return cfg
}

// npmTemplateData holds data for npm templates.
type npmTemplateData struct {
	Package     string
	Version     string
	Description string
	License     string
	Repository  string
	BinaryName  string
	ProjectName string
	Access      string
}

// dryRunPublish shows what would be done without actually publishing.
func (p *NpmPublisher) dryRunPublish(m io.Medium, data npmTemplateData, cfg *NpmConfig) error {
	fmt.Println()
	fmt.Println("=== DRY RUN: npm Publish ===")
	fmt.Println()
	fmt.Printf("Package:    %s\n", data.Package)
	fmt.Printf("Version:    %s\n", data.Version)
	fmt.Printf("Access:     %s\n", data.Access)
	fmt.Printf("Repository: %s\n", data.Repository)
	fmt.Printf("Binary:     %s\n", data.BinaryName)
	fmt.Println()

	// Generate and show package.json
	pkgJSON, err := p.renderTemplate(m, "templates/npm/package.json.tmpl", data)
	if err != nil {
		return fmt.Errorf("npm.dryRunPublish: %w", err)
	}
	fmt.Println("Generated package.json:")
	fmt.Println("---")
	fmt.Println(pkgJSON)
	fmt.Println("---")
	fmt.Println()

	fmt.Println("Would run: npm publish --access", data.Access)
	fmt.Println()
	fmt.Println("=== END DRY RUN ===")

	return nil
}

// executePublish actually creates and publishes the npm package.
func (p *NpmPublisher) executePublish(ctx context.Context, m io.Medium, data npmTemplateData, cfg *NpmConfig) error {
	// Check for NPM_TOKEN
	if os.Getenv("NPM_TOKEN") == "" {
		return fmt.Errorf("npm.Publish: NPM_TOKEN environment variable is required")
	}

	// Create temp directory for package
	tmpDir, err := os.MkdirTemp("", "npm-publish-*")
	if err != nil {
		return fmt.Errorf("npm.Publish: failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create bin directory
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("npm.Publish: failed to create bin directory: %w", err)
	}

	// Generate package.json
	pkgJSON, err := p.renderTemplate(m, "templates/npm/package.json.tmpl", data)
	if err != nil {
		return fmt.Errorf("npm.Publish: failed to render package.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		return fmt.Errorf("npm.Publish: failed to write package.json: %w", err)
	}

	// Generate install.js
	installJS, err := p.renderTemplate(m, "templates/npm/install.js.tmpl", data)
	if err != nil {
		return fmt.Errorf("npm.Publish: failed to render install.js: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "install.js"), []byte(installJS), 0644); err != nil {
		return fmt.Errorf("npm.Publish: failed to write install.js: %w", err)
	}

	// Generate run.js
	runJS, err := p.renderTemplate(m, "templates/npm/run.js.tmpl", data)
	if err != nil {
		return fmt.Errorf("npm.Publish: failed to render run.js: %w", err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "run.js"), []byte(runJS), 0755); err != nil {
		return fmt.Errorf("npm.Publish: failed to write run.js: %w", err)
	}

	// Create .npmrc with token
	npmrc := "//registry.npmjs.org/:_authToken=${NPM_TOKEN}\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".npmrc"), []byte(npmrc), 0600); err != nil {
		return fmt.Errorf("npm.Publish: failed to write .npmrc: %w", err)
	}

	// Run npm publish
	cmd := exec.CommandContext(ctx, "npm", "publish", "--access", data.Access)
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "NPM_TOKEN="+os.Getenv("NPM_TOKEN"))

	fmt.Printf("Publishing %s@%s to npm...\n", data.Package, data.Version)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm.Publish: npm publish failed: %w", err)
	}

	fmt.Printf("Published %s@%s to npm\n", data.Package, data.Version)
	fmt.Printf("  https://www.npmjs.com/package/%s\n", data.Package)

	return nil
}

// renderTemplate renders an embedded template with the given data.
func (p *NpmPublisher) renderTemplate(m io.Medium, name string, data npmTemplateData) (string, error) {
	var content []byte
	var err error

	// Try custom template from medium
	customPath := filepath.Join(".core", name)
	if m != nil && m.IsFile(customPath) {
		customContent, err := m.Read(customPath)
		if err == nil {
			content = []byte(customContent)
		}
	}

	// Fallback to embedded template
	if content == nil {
		content, err = npmTemplates.ReadFile(name)
		if err != nil {
			return "", fmt.Errorf("failed to read template %s: %w", name, err)
		}
	}

	tmpl, err := template.New(filepath.Base(name)).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}
