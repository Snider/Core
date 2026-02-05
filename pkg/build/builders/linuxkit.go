// Package builders provides build implementations for different project types.
package builders

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/build"
)

// LinuxKitBuilder builds LinuxKit images.
type LinuxKitBuilder struct{}

// NewLinuxKitBuilder creates a new LinuxKit builder.
func NewLinuxKitBuilder() *LinuxKitBuilder {
	return &LinuxKitBuilder{}
}

// Name returns the builder's identifier.
func (b *LinuxKitBuilder) Name() string {
	return "linuxkit"
}

// Detect checks if a linuxkit.yml or .yml config exists in the directory.
func (b *LinuxKitBuilder) Detect(dir string) (bool, error) {
	// Check for linuxkit.yml
	if _, err := os.Stat(filepath.Join(dir, "linuxkit.yml")); err == nil {
		return true, nil
	}
	// Check for .core/linuxkit/*.yml
	if matches, _ := filepath.Glob(filepath.Join(dir, ".core", "linuxkit", "*.yml")); len(matches) > 0 {
		return true, nil
	}
	return false, nil
}

// Build builds LinuxKit images for the specified targets.
func (b *LinuxKitBuilder) Build(ctx context.Context, cfg *build.Config, targets []build.Target) ([]build.Artifact, error) {
	// Validate linuxkit CLI is available
	if err := b.validateLinuxKitCli(); err != nil {
		return nil, err
	}

	// Determine config file path
	configPath := cfg.LinuxKitConfig
	if configPath == "" {
		// Auto-detect
		if _, err := os.Stat(filepath.Join(cfg.ProjectDir, "linuxkit.yml")); err == nil {
			configPath = filepath.Join(cfg.ProjectDir, "linuxkit.yml")
		} else {
			// Look in .core/linuxkit/
			matches, _ := filepath.Glob(filepath.Join(cfg.ProjectDir, ".core", "linuxkit", "*.yml"))
			if len(matches) > 0 {
				configPath = matches[0]
			}
		}
	}

	if configPath == "" {
		return nil, fmt.Errorf("linuxkit.Build: no LinuxKit config file found. Specify with --config or create linuxkit.yml")
	}

	// Validate config file exists
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("linuxkit.Build: config file not found: %s", configPath)
	}

	// Determine output formats
	formats := cfg.Formats
	if len(formats) == 0 {
		formats = []string{"qcow2-bios"} // Default to QEMU-compatible format
	}

	// Create output directory
	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join(cfg.ProjectDir, "dist")
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("linuxkit.Build: failed to create output directory: %w", err)
	}

	// Determine base name from config file or project name
	baseName := cfg.Name
	if baseName == "" {
		baseName = strings.TrimSuffix(filepath.Base(configPath), ".yml")
	}

	// If no targets, default to linux/amd64
	if len(targets) == 0 {
		targets = []build.Target{{OS: "linux", Arch: "amd64"}}
	}

	var artifacts []build.Artifact

	// Build for each target and format
	for _, target := range targets {
		// LinuxKit only supports Linux
		if target.OS != "linux" {
			fmt.Printf("Skipping %s/%s (LinuxKit only supports Linux)\n", target.OS, target.Arch)
			continue
		}

		for _, format := range formats {
			outputName := fmt.Sprintf("%s-%s", baseName, target.Arch)

			args := b.buildLinuxKitArgs(configPath, format, outputName, outputDir, target.Arch)

			cmd := exec.CommandContext(ctx, "linuxkit", args...)
			cmd.Dir = cfg.ProjectDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			fmt.Printf("Building LinuxKit image: %s (%s, %s)\n", outputName, format, target.Arch)

			if err := cmd.Run(); err != nil {
				return nil, fmt.Errorf("linuxkit.Build: build failed for %s/%s: %w", target.Arch, format, err)
			}

			// Determine the actual output file path
			artifactPath := b.getArtifactPath(outputDir, outputName, format)

			// Verify the artifact was created
			if _, err := os.Stat(artifactPath); err != nil {
				// Try alternate naming conventions
				artifactPath = b.findArtifact(outputDir, outputName, format)
				if artifactPath == "" {
					return nil, fmt.Errorf("linuxkit.Build: artifact not found after build: expected %s", b.getArtifactPath(outputDir, outputName, format))
				}
			}

			artifacts = append(artifacts, build.Artifact{
				Path: artifactPath,
				OS:   target.OS,
				Arch: target.Arch,
			})
		}
	}

	return artifacts, nil
}

// buildLinuxKitArgs builds the arguments for linuxkit build command.
func (b *LinuxKitBuilder) buildLinuxKitArgs(configPath, format, outputName, outputDir, arch string) []string {
	args := []string{"build"}

	// Output format
	args = append(args, "--format", format)

	// Output name
	args = append(args, "--name", outputName)

	// Output directory
	args = append(args, "--dir", outputDir)

	// Architecture (if not amd64)
	if arch != "amd64" {
		args = append(args, "--arch", arch)
	}

	// Config file
	args = append(args, configPath)

	return args
}

// getArtifactPath returns the expected path of the built artifact.
func (b *LinuxKitBuilder) getArtifactPath(outputDir, outputName, format string) string {
	ext := b.getFormatExtension(format)
	return filepath.Join(outputDir, outputName+ext)
}

// findArtifact searches for the built artifact with various naming conventions.
func (b *LinuxKitBuilder) findArtifact(outputDir, outputName, format string) string {
	// LinuxKit can create files with different suffixes
	extensions := []string{
		b.getFormatExtension(format),
		"-bios" + b.getFormatExtension(format),
		"-efi" + b.getFormatExtension(format),
	}

	for _, ext := range extensions {
		path := filepath.Join(outputDir, outputName+ext)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Try to find any file matching the output name
	matches, _ := filepath.Glob(filepath.Join(outputDir, outputName+"*"))
	for _, match := range matches {
		// Return first match that looks like an image
		ext := filepath.Ext(match)
		if ext == ".iso" || ext == ".qcow2" || ext == ".raw" || ext == ".vmdk" || ext == ".vhd" {
			return match
		}
	}

	return ""
}

// getFormatExtension returns the file extension for a LinuxKit output format.
func (b *LinuxKitBuilder) getFormatExtension(format string) string {
	switch format {
	case "iso", "iso-bios", "iso-efi":
		return ".iso"
	case "raw", "raw-bios", "raw-efi":
		return ".raw"
	case "qcow2", "qcow2-bios", "qcow2-efi":
		return ".qcow2"
	case "vmdk":
		return ".vmdk"
	case "vhd":
		return ".vhd"
	case "gcp":
		return ".img.tar.gz"
	case "aws":
		return ".raw"
	default:
		return "." + strings.TrimSuffix(format, "-bios")
	}
}

// validateLinuxKitCli checks if the linuxkit CLI is available.
func (b *LinuxKitBuilder) validateLinuxKitCli() error {
	// Check PATH first
	if _, err := exec.LookPath("linuxkit"); err == nil {
		return nil
	}

	// Check common locations
	paths := []string{
		"/usr/local/bin/linuxkit",
		"/opt/homebrew/bin/linuxkit",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return nil
		}
	}

	return fmt.Errorf("linuxkit: linuxkit CLI not found. Install with: brew install linuxkit (macOS) or see https://github.com/linuxkit/linuxkit")
}
