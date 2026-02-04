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

	"github.com/host-uk/core/pkg/build"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/io"
)

//go:embed templates/chocolatey/*.tmpl templates/chocolatey/tools/*.tmpl
var chocolateyTemplates embed.FS

// ChocolateyConfig holds Chocolatey-specific configuration.
type ChocolateyConfig struct {
	// Package is the Chocolatey package name.
	Package string
	// Push determines whether to push to Chocolatey (false = generate only).
	Push bool
	// Official config for generating files for official repo PRs.
	Official *OfficialConfig
}

// ChocolateyPublisher publishes releases to Chocolatey.
type ChocolateyPublisher struct{}

// NewChocolateyPublisher creates a new Chocolatey publisher.
func NewChocolateyPublisher() *ChocolateyPublisher {
	return &ChocolateyPublisher{}
}

// Name returns the publisher's identifier.
func (p *ChocolateyPublisher) Name() string {
	return "chocolatey"
}

// Publish publishes the release to Chocolatey.
func (p *ChocolateyPublisher) Publish(ctx context.Context, release *Release, pubCfg PublisherConfig, relCfg ReleaseConfig, dryRun bool) error {
	cfg := p.parseConfig(pubCfg, relCfg)

	repo := ""
	if relCfg != nil {
		repo = relCfg.GetRepository()
	}
	if repo == "" {
		detectedRepo, err := detectRepository(release.ProjectDir)
		if err != nil {
			return fmt.Errorf("chocolatey.Publish: could not determine repository: %w", err)
		}
		repo = detectedRepo
	}

	projectName := ""
	if relCfg != nil {
		projectName = relCfg.GetProjectName()
	}
	if projectName == "" {
		parts := strings.Split(repo, "/")
		projectName = parts[len(parts)-1]
	}

	packageName := cfg.Package
	if packageName == "" {
		packageName = projectName
	}

	version := strings.TrimPrefix(release.Version, "v")
	checksums := buildChecksumMap(release.Artifacts)

	// Extract authors from repository
	authors := strings.Split(repo, "/")[0]

	data := chocolateyTemplateData{
		PackageName: packageName,
		Title:       fmt.Sprintf("%s CLI", i18n.Title(projectName)),
		Description: fmt.Sprintf("%s CLI", projectName),
		Repository:  repo,
		Version:     version,
		License:     "MIT",
		BinaryName:  projectName,
		Authors:     authors,
		Tags:        fmt.Sprintf("cli %s", projectName),
		Checksums:   checksums,
	}

	if dryRun {
		return p.dryRunPublish(release.FS, data, cfg)
	}

	return p.executePublish(ctx, release.ProjectDir, data, cfg, release)
}

type chocolateyTemplateData struct {
	PackageName string
	Title       string
	Description string
	Repository  string
	Version     string
	License     string
	BinaryName  string
	Authors     string
	Tags        string
	Checksums   ChecksumMap
}

func (p *ChocolateyPublisher) parseConfig(pubCfg PublisherConfig, relCfg ReleaseConfig) ChocolateyConfig {
	cfg := ChocolateyConfig{
		Push: false, // Default to generate only
	}

	if ext, ok := pubCfg.Extended.(map[string]any); ok {
		if pkg, ok := ext["package"].(string); ok && pkg != "" {
			cfg.Package = pkg
		}
		if push, ok := ext["push"].(bool); ok {
			cfg.Push = push
		}
		if official, ok := ext["official"].(map[string]any); ok {
			cfg.Official = &OfficialConfig{}
			if enabled, ok := official["enabled"].(bool); ok {
				cfg.Official.Enabled = enabled
			}
			if output, ok := official["output"].(string); ok {
				cfg.Official.Output = output
			}
		}
	}

	return cfg
}

func (p *ChocolateyPublisher) dryRunPublish(m io.Medium, data chocolateyTemplateData, cfg ChocolateyConfig) error {
	fmt.Println()
	fmt.Println("=== DRY RUN: Chocolatey Publish ===")
	fmt.Println()
	fmt.Printf("Package:    %s\n", data.PackageName)
	fmt.Printf("Version:    %s\n", data.Version)
	fmt.Printf("Push:       %t\n", cfg.Push)
	fmt.Printf("Repository: %s\n", data.Repository)
	fmt.Println()

	nuspec, err := p.renderTemplate(m, "templates/chocolatey/package.nuspec.tmpl", data)
	if err != nil {
		return fmt.Errorf("chocolatey.dryRunPublish: %w", err)
	}
	fmt.Println("Generated package.nuspec:")
	fmt.Println("---")
	fmt.Println(nuspec)
	fmt.Println("---")
	fmt.Println()

	install, err := p.renderTemplate(m, "templates/chocolatey/tools/chocolateyinstall.ps1.tmpl", data)
	if err != nil {
		return fmt.Errorf("chocolatey.dryRunPublish: %w", err)
	}
	fmt.Println("Generated chocolateyinstall.ps1:")
	fmt.Println("---")
	fmt.Println(install)
	fmt.Println("---")
	fmt.Println()

	if cfg.Push {
		fmt.Println("Would push to Chocolatey community repo")
	} else {
		fmt.Println("Would generate package files only (push=false)")
	}
	fmt.Println()
	fmt.Println("=== END DRY RUN ===")

	return nil
}

func (p *ChocolateyPublisher) executePublish(ctx context.Context, projectDir string, data chocolateyTemplateData, cfg ChocolateyConfig, release *Release) error {
	nuspec, err := p.renderTemplate(release.FS, "templates/chocolatey/package.nuspec.tmpl", data)
	if err != nil {
		return fmt.Errorf("chocolatey.Publish: failed to render nuspec: %w", err)
	}

	install, err := p.renderTemplate(release.FS, "templates/chocolatey/tools/chocolateyinstall.ps1.tmpl", data)
	if err != nil {
		return fmt.Errorf("chocolatey.Publish: failed to render install script: %w", err)
	}

	// Create package directory
	output := filepath.Join(projectDir, "dist", "chocolatey")
	if cfg.Official != nil && cfg.Official.Enabled && cfg.Official.Output != "" {
		output = cfg.Official.Output
		if !filepath.IsAbs(output) {
			output = filepath.Join(projectDir, output)
		}
	}

	toolsDir := filepath.Join(output, "tools")
	if err := release.FS.EnsureDir(toolsDir); err != nil {
		return fmt.Errorf("chocolatey.Publish: failed to create output directory: %w", err)
	}

	// Write files
	nuspecPath := filepath.Join(output, fmt.Sprintf("%s.nuspec", data.PackageName))
	if err := release.FS.Write(nuspecPath, nuspec); err != nil {
		return fmt.Errorf("chocolatey.Publish: failed to write nuspec: %w", err)
	}

	installPath := filepath.Join(toolsDir, "chocolateyinstall.ps1")
	if err := release.FS.Write(installPath, install); err != nil {
		return fmt.Errorf("chocolatey.Publish: failed to write install script: %w", err)
	}

	fmt.Printf("Wrote Chocolatey package files: %s\n", output)

	// Push to Chocolatey if configured
	if cfg.Push {
		if err := p.pushToChocolatey(ctx, output, data); err != nil {
			return err
		}
	}

	return nil
}

func (p *ChocolateyPublisher) pushToChocolatey(ctx context.Context, packageDir string, data chocolateyTemplateData) error {
	// Check for CHOCOLATEY_API_KEY
	apiKey := os.Getenv("CHOCOLATEY_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("chocolatey.Publish: CHOCOLATEY_API_KEY environment variable is required for push")
	}

	// Pack the package
	nupkgPath := filepath.Join(packageDir, fmt.Sprintf("%s.%s.nupkg", data.PackageName, data.Version))

	cmd := exec.CommandContext(ctx, "choco", "pack", filepath.Join(packageDir, fmt.Sprintf("%s.nuspec", data.PackageName)), "-OutputDirectory", packageDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("chocolatey.Publish: choco pack failed: %w", err)
	}

	// Push the package
	cmd = exec.CommandContext(ctx, "choco", "push", nupkgPath, "--source", "https://push.chocolatey.org/", "--api-key", apiKey)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("chocolatey.Publish: choco push failed: %w", err)
	}

	fmt.Printf("Published to Chocolatey: https://community.chocolatey.org/packages/%s\n", data.PackageName)
	return nil
}

func (p *ChocolateyPublisher) renderTemplate(m io.Medium, name string, data chocolateyTemplateData) (string, error) {
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
		content, err = chocolateyTemplates.ReadFile(name)
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

// Ensure build package is used
var _ = build.Artifact{}
