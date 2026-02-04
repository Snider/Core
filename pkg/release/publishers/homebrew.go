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
	"github.com/host-uk/core/pkg/io"
)

//go:embed templates/homebrew/*.tmpl
var homebrewTemplates embed.FS

// HomebrewConfig holds Homebrew-specific configuration.
type HomebrewConfig struct {
	// Tap is the Homebrew tap repository (e.g., "host-uk/homebrew-tap").
	Tap string
	// Formula is the formula name (defaults to project name).
	Formula string
	// Official config for generating files for official repo PRs.
	Official *OfficialConfig
}

// OfficialConfig holds configuration for generating files for official repo PRs.
type OfficialConfig struct {
	// Enabled determines whether to generate files for official repos.
	Enabled bool
	// Output is the directory to write generated files.
	Output string
}

// HomebrewPublisher publishes releases to Homebrew.
type HomebrewPublisher struct{}

// NewHomebrewPublisher creates a new Homebrew publisher.
func NewHomebrewPublisher() *HomebrewPublisher {
	return &HomebrewPublisher{}
}

// Name returns the publisher's identifier.
func (p *HomebrewPublisher) Name() string {
	return "homebrew"
}

// Publish publishes the release to Homebrew.
func (p *HomebrewPublisher) Publish(ctx context.Context, release *Release, pubCfg PublisherConfig, relCfg ReleaseConfig, dryRun bool) error {
	// Parse config
	cfg := p.parseConfig(pubCfg, relCfg)

	// Validate configuration
	if cfg.Tap == "" && (cfg.Official == nil || !cfg.Official.Enabled) {
		return fmt.Errorf("homebrew.Publish: tap is required (set publish.homebrew.tap in config)")
	}

	// Get repository and project info
	repo := ""
	if relCfg != nil {
		repo = relCfg.GetRepository()
	}
	if repo == "" {
		detectedRepo, err := detectRepository(release.ProjectDir)
		if err != nil {
			return fmt.Errorf("homebrew.Publish: could not determine repository: %w", err)
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

	formulaName := cfg.Formula
	if formulaName == "" {
		formulaName = projectName
	}

	// Strip leading 'v' from version
	version := strings.TrimPrefix(release.Version, "v")

	// Build checksums map from artifacts
	checksums := buildChecksumMap(release.Artifacts)

	// Template data
	data := homebrewTemplateData{
		FormulaClass: toFormulaClass(formulaName),
		Description:  fmt.Sprintf("%s CLI", projectName),
		Repository:   repo,
		Version:      version,
		License:      "MIT",
		BinaryName:   projectName,
		Checksums:    checksums,
	}

	if dryRun {
		return p.dryRunPublish(release.FS, data, cfg)
	}

	return p.executePublish(ctx, release.ProjectDir, data, cfg, release)
}

// homebrewTemplateData holds data for Homebrew templates.
type homebrewTemplateData struct {
	FormulaClass string
	Description  string
	Repository   string
	Version      string
	License      string
	BinaryName   string
	Checksums    ChecksumMap
}

// ChecksumMap holds checksums for different platform/arch combinations.
type ChecksumMap struct {
	DarwinAmd64  string
	DarwinArm64  string
	LinuxAmd64   string
	LinuxArm64   string
	WindowsAmd64 string
	WindowsArm64 string
}

// parseConfig extracts Homebrew-specific configuration.
func (p *HomebrewPublisher) parseConfig(pubCfg PublisherConfig, relCfg ReleaseConfig) HomebrewConfig {
	cfg := HomebrewConfig{
		Tap:     "",
		Formula: "",
	}

	if ext, ok := pubCfg.Extended.(map[string]any); ok {
		if tap, ok := ext["tap"].(string); ok && tap != "" {
			cfg.Tap = tap
		}
		if formula, ok := ext["formula"].(string); ok && formula != "" {
			cfg.Formula = formula
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

// dryRunPublish shows what would be done.
func (p *HomebrewPublisher) dryRunPublish(m io.Medium, data homebrewTemplateData, cfg HomebrewConfig) error {
	fmt.Println()
	fmt.Println("=== DRY RUN: Homebrew Publish ===")
	fmt.Println()
	fmt.Printf("Formula:    %s\n", data.FormulaClass)
	fmt.Printf("Version:    %s\n", data.Version)
	fmt.Printf("Tap:        %s\n", cfg.Tap)
	fmt.Printf("Repository: %s\n", data.Repository)
	fmt.Println()

	// Generate and show formula
	formula, err := p.renderTemplate(m, "templates/homebrew/formula.rb.tmpl", data)
	if err != nil {
		return fmt.Errorf("homebrew.dryRunPublish: %w", err)
	}
	fmt.Println("Generated formula.rb:")
	fmt.Println("---")
	fmt.Println(formula)
	fmt.Println("---")
	fmt.Println()

	if cfg.Tap != "" {
		fmt.Printf("Would commit to tap: %s\n", cfg.Tap)
	}
	if cfg.Official != nil && cfg.Official.Enabled {
		output := cfg.Official.Output
		if output == "" {
			output = "dist/homebrew"
		}
		fmt.Printf("Would write files for official PR to: %s\n", output)
	}
	fmt.Println()
	fmt.Println("=== END DRY RUN ===")

	return nil
}

// executePublish creates the formula and commits to tap.
func (p *HomebrewPublisher) executePublish(ctx context.Context, projectDir string, data homebrewTemplateData, cfg HomebrewConfig, release *Release) error {
	// Generate formula
	formula, err := p.renderTemplate(release.FS, "templates/homebrew/formula.rb.tmpl", data)
	if err != nil {
		return fmt.Errorf("homebrew.Publish: failed to render formula: %w", err)
	}

	// If official config is enabled, write to output directory
	if cfg.Official != nil && cfg.Official.Enabled {
		output := cfg.Official.Output
		if output == "" {
			output = filepath.Join(projectDir, "dist", "homebrew")
		} else if !filepath.IsAbs(output) {
			output = filepath.Join(projectDir, output)
		}

		if err := release.FS.EnsureDir(output); err != nil {
			return fmt.Errorf("homebrew.Publish: failed to create output directory: %w", err)
		}

		formulaPath := filepath.Join(output, fmt.Sprintf("%s.rb", strings.ToLower(data.FormulaClass)))
		if err := release.FS.Write(formulaPath, formula); err != nil {
			return fmt.Errorf("homebrew.Publish: failed to write formula: %w", err)
		}
		fmt.Printf("Wrote Homebrew formula for official PR: %s\n", formulaPath)
	}

	// If tap is configured, commit to it
	if cfg.Tap != "" {
		if err := p.commitToTap(ctx, cfg.Tap, data, formula); err != nil {
			return err
		}
	}

	return nil
}

// commitToTap commits the formula to the tap repository.
func (p *HomebrewPublisher) commitToTap(ctx context.Context, tap string, data homebrewTemplateData, formula string) error {
	// Clone tap repo to temp directory
	tmpDir, err := os.MkdirTemp("", "homebrew-tap-*")
	if err != nil {
		return fmt.Errorf("homebrew.Publish: failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Clone the tap
	fmt.Printf("Cloning tap %s...\n", tap)
	cmd := exec.CommandContext(ctx, "gh", "repo", "clone", tap, tmpDir, "--", "--depth=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("homebrew.Publish: failed to clone tap: %w", err)
	}

	// Ensure Formula directory exists
	formulaDir := filepath.Join(tmpDir, "Formula")
	if err := os.MkdirAll(formulaDir, 0755); err != nil {
		return fmt.Errorf("homebrew.Publish: failed to create Formula directory: %w", err)
	}

	// Write formula
	formulaPath := filepath.Join(formulaDir, fmt.Sprintf("%s.rb", strings.ToLower(data.FormulaClass)))
	if err := os.WriteFile(formulaPath, []byte(formula), 0644); err != nil {
		return fmt.Errorf("homebrew.Publish: failed to write formula: %w", err)
	}

	// Git add, commit, push
	commitMsg := fmt.Sprintf("Update %s to %s", data.FormulaClass, data.Version)

	cmd = exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("homebrew.Publish: git add failed: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", commitMsg)
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("homebrew.Publish: git commit failed: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "push")
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("homebrew.Publish: git push failed: %w", err)
	}

	fmt.Printf("Updated Homebrew tap: %s\n", tap)
	return nil
}

// renderTemplate renders an embedded template with the given data.
func (p *HomebrewPublisher) renderTemplate(m io.Medium, name string, data homebrewTemplateData) (string, error) {
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
		content, err = homebrewTemplates.ReadFile(name)
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

// toFormulaClass converts a package name to a Ruby class name.
func toFormulaClass(name string) string {
	// Convert kebab-case to PascalCase
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// buildChecksumMap extracts checksums from artifacts into a structured map.
func buildChecksumMap(artifacts []build.Artifact) ChecksumMap {
	checksums := ChecksumMap{}

	for _, a := range artifacts {
		// Parse artifact name to determine platform
		name := filepath.Base(a.Path)
		checksum := a.Checksum

		switch {
		case strings.Contains(name, "darwin-amd64"):
			checksums.DarwinAmd64 = checksum
		case strings.Contains(name, "darwin-arm64"):
			checksums.DarwinArm64 = checksum
		case strings.Contains(name, "linux-amd64"):
			checksums.LinuxAmd64 = checksum
		case strings.Contains(name, "linux-arm64"):
			checksums.LinuxArm64 = checksum
		case strings.Contains(name, "windows-amd64"):
			checksums.WindowsAmd64 = checksum
		case strings.Contains(name, "windows-arm64"):
			checksums.WindowsArm64 = checksum
		}
	}

	return checksums
}
