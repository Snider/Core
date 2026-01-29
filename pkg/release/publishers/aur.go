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
)

//go:embed templates/aur/*.tmpl
var aurTemplates embed.FS

// AURConfig holds AUR-specific configuration.
type AURConfig struct {
	// Package is the AUR package name.
	Package string
	// Maintainer is the package maintainer (e.g., "Name <email>").
	Maintainer string
	// Official config for generating files for official repo PRs.
	Official *OfficialConfig
}

// AURPublisher publishes releases to AUR.
type AURPublisher struct{}

// NewAURPublisher creates a new AUR publisher.
func NewAURPublisher() *AURPublisher {
	return &AURPublisher{}
}

// Name returns the publisher's identifier.
func (p *AURPublisher) Name() string {
	return "aur"
}

// Publish publishes the release to AUR.
func (p *AURPublisher) Publish(ctx context.Context, release *Release, pubCfg PublisherConfig, relCfg ReleaseConfig, dryRun bool) error {
	cfg := p.parseConfig(pubCfg, relCfg)

	if cfg.Maintainer == "" {
		return fmt.Errorf("aur.Publish: maintainer is required (set publish.aur.maintainer in config)")
	}

	repo := ""
	if relCfg != nil {
		repo = relCfg.GetRepository()
	}
	if repo == "" {
		detectedRepo, err := detectRepository(release.ProjectDir)
		if err != nil {
			return fmt.Errorf("aur.Publish: could not determine repository: %w", err)
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

	data := aurTemplateData{
		PackageName: packageName,
		Description: fmt.Sprintf("%s CLI", projectName),
		Repository:  repo,
		Version:     version,
		License:     "MIT",
		BinaryName:  projectName,
		Maintainer:  cfg.Maintainer,
		Checksums:   checksums,
	}

	if dryRun {
		return p.dryRunPublish(data, cfg)
	}

	return p.executePublish(ctx, release.ProjectDir, data, cfg)
}

type aurTemplateData struct {
	PackageName string
	Description string
	Repository  string
	Version     string
	License     string
	BinaryName  string
	Maintainer  string
	Checksums   ChecksumMap
}

func (p *AURPublisher) parseConfig(pubCfg PublisherConfig, relCfg ReleaseConfig) AURConfig {
	cfg := AURConfig{}

	if ext, ok := pubCfg.Extended.(map[string]any); ok {
		if pkg, ok := ext["package"].(string); ok && pkg != "" {
			cfg.Package = pkg
		}
		if maintainer, ok := ext["maintainer"].(string); ok && maintainer != "" {
			cfg.Maintainer = maintainer
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

func (p *AURPublisher) dryRunPublish(data aurTemplateData, cfg AURConfig) error {
	fmt.Println()
	fmt.Println("=== DRY RUN: AUR Publish ===")
	fmt.Println()
	fmt.Printf("Package:    %s-bin\n", data.PackageName)
	fmt.Printf("Version:    %s\n", data.Version)
	fmt.Printf("Maintainer: %s\n", data.Maintainer)
	fmt.Printf("Repository: %s\n", data.Repository)
	fmt.Println()

	pkgbuild, err := p.renderTemplate("templates/aur/PKGBUILD.tmpl", data)
	if err != nil {
		return fmt.Errorf("aur.dryRunPublish: %w", err)
	}
	fmt.Println("Generated PKGBUILD:")
	fmt.Println("---")
	fmt.Println(pkgbuild)
	fmt.Println("---")
	fmt.Println()

	srcinfo, err := p.renderTemplate("templates/aur/.SRCINFO.tmpl", data)
	if err != nil {
		return fmt.Errorf("aur.dryRunPublish: %w", err)
	}
	fmt.Println("Generated .SRCINFO:")
	fmt.Println("---")
	fmt.Println(srcinfo)
	fmt.Println("---")
	fmt.Println()

	fmt.Printf("Would push to AUR: ssh://aur@aur.archlinux.org/%s-bin.git\n", data.PackageName)
	fmt.Println()
	fmt.Println("=== END DRY RUN ===")

	return nil
}

func (p *AURPublisher) executePublish(ctx context.Context, projectDir string, data aurTemplateData, cfg AURConfig) error {
	pkgbuild, err := p.renderTemplate("templates/aur/PKGBUILD.tmpl", data)
	if err != nil {
		return fmt.Errorf("aur.Publish: failed to render PKGBUILD: %w", err)
	}

	srcinfo, err := p.renderTemplate("templates/aur/.SRCINFO.tmpl", data)
	if err != nil {
		return fmt.Errorf("aur.Publish: failed to render .SRCINFO: %w", err)
	}

	// If official config is enabled, write to output directory
	if cfg.Official != nil && cfg.Official.Enabled {
		output := cfg.Official.Output
		if output == "" {
			output = filepath.Join(projectDir, "dist", "aur")
		} else if !filepath.IsAbs(output) {
			output = filepath.Join(projectDir, output)
		}

		if err := os.MkdirAll(output, 0755); err != nil {
			return fmt.Errorf("aur.Publish: failed to create output directory: %w", err)
		}

		pkgbuildPath := filepath.Join(output, "PKGBUILD")
		if err := os.WriteFile(pkgbuildPath, []byte(pkgbuild), 0644); err != nil {
			return fmt.Errorf("aur.Publish: failed to write PKGBUILD: %w", err)
		}

		srcinfoPath := filepath.Join(output, ".SRCINFO")
		if err := os.WriteFile(srcinfoPath, []byte(srcinfo), 0644); err != nil {
			return fmt.Errorf("aur.Publish: failed to write .SRCINFO: %w", err)
		}
		fmt.Printf("Wrote AUR files: %s\n", output)
	}

	// Push to AUR if not in official-only mode
	if cfg.Official == nil || !cfg.Official.Enabled {
		if err := p.pushToAUR(ctx, data, pkgbuild, srcinfo); err != nil {
			return err
		}
	}

	return nil
}

func (p *AURPublisher) pushToAUR(ctx context.Context, data aurTemplateData, pkgbuild, srcinfo string) error {
	aurURL := fmt.Sprintf("ssh://aur@aur.archlinux.org/%s-bin.git", data.PackageName)

	tmpDir, err := os.MkdirTemp("", "aur-package-*")
	if err != nil {
		return fmt.Errorf("aur.Publish: failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Clone existing AUR repo (or initialize new one)
	fmt.Printf("Cloning AUR package %s-bin...\n", data.PackageName)
	cmd := exec.CommandContext(ctx, "git", "clone", aurURL, tmpDir)
	if err := cmd.Run(); err != nil {
		// If clone fails, init a new repo
		cmd = exec.CommandContext(ctx, "git", "init", tmpDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("aur.Publish: failed to initialize repo: %w", err)
		}
		cmd = exec.CommandContext(ctx, "git", "-C", tmpDir, "remote", "add", "origin", aurURL)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("aur.Publish: failed to add remote: %w", err)
		}
	}

	// Write files
	if err := os.WriteFile(filepath.Join(tmpDir, "PKGBUILD"), []byte(pkgbuild), 0644); err != nil {
		return fmt.Errorf("aur.Publish: failed to write PKGBUILD: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".SRCINFO"), []byte(srcinfo), 0644); err != nil {
		return fmt.Errorf("aur.Publish: failed to write .SRCINFO: %w", err)
	}

	commitMsg := fmt.Sprintf("Update to %s", data.Version)

	cmd = exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("aur.Publish: git add failed: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", commitMsg)
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("aur.Publish: git commit failed: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "push", "origin", "master")
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("aur.Publish: git push failed: %w", err)
	}

	fmt.Printf("Published to AUR: https://aur.archlinux.org/packages/%s-bin\n", data.PackageName)
	return nil
}

func (p *AURPublisher) renderTemplate(name string, data aurTemplateData) (string, error) {
	content, err := aurTemplates.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", name, err)
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
