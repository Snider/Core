// Package builders provides build implementations for different project types.
package builders

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/build"
	"github.com/host-uk/core/pkg/io"
)

// WailsBuilder implements the Builder interface for Wails v3 projects.
type WailsBuilder struct{}

// NewWailsBuilder creates a new WailsBuilder instance.
func NewWailsBuilder() *WailsBuilder {
	return &WailsBuilder{}
}

// Name returns the builder's identifier.
func (b *WailsBuilder) Name() string {
	return "wails"
}

// Detect checks if this builder can handle the project in the given directory.
// Uses IsWailsProject from the build package which checks for wails.json.
func (b *WailsBuilder) Detect(fs io.Medium, dir string) (bool, error) {
	return build.IsWailsProject(fs, dir), nil
}

// Build compiles the Wails project for the specified targets.
// It detects the Wails version and chooses the appropriate build strategy:
// - Wails v3: Delegates to Taskfile (error if missing)
// - Wails v2: Uses 'wails build' command
func (b *WailsBuilder) Build(ctx context.Context, cfg *build.Config, targets []build.Target) ([]build.Artifact, error) {
	if cfg == nil {
		return nil, fmt.Errorf("builders.WailsBuilder.Build: config is nil")
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("builders.WailsBuilder.Build: no targets specified")
	}

	// Detect Wails version
	isV3 := b.isWailsV3(cfg.FS, cfg.ProjectDir)

	if isV3 {
		// Wails v3 strategy: Delegate to Taskfile
		taskBuilder := NewTaskfileBuilder()
		if detected, _ := taskBuilder.Detect(cfg.FS, cfg.ProjectDir); detected {
			return taskBuilder.Build(ctx, cfg, targets)
		}
		return nil, fmt.Errorf("wails v3 projects require a Taskfile for building")
	}

	// Wails v2 strategy: Use 'wails build'
	// Ensure output directory exists
	if err := cfg.FS.EnsureDir(cfg.OutputDir); err != nil {
		return nil, fmt.Errorf("builders.WailsBuilder.Build: failed to create output directory: %w", err)
	}

	// Note: Wails v2 handles frontend installation/building automatically via wails.json config

	var artifacts []build.Artifact

	for _, target := range targets {
		artifact, err := b.buildV2Target(ctx, cfg, target)
		if err != nil {
			return artifacts, fmt.Errorf("builders.WailsBuilder.Build: failed to build %s: %w", target.String(), err)
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

// isWailsV3 checks if the project uses Wails v3 by inspecting go.mod.
func (b *WailsBuilder) isWailsV3(fs io.Medium, dir string) bool {
	goModPath := filepath.Join(dir, "go.mod")
	content, err := fs.Read(goModPath)
	if err != nil {
		return false
	}
	return strings.Contains(content, "github.com/wailsapp/wails/v3")
}

// buildV2Target compiles for a single target platform using wails (v2).
func (b *WailsBuilder) buildV2Target(ctx context.Context, cfg *build.Config, target build.Target) (build.Artifact, error) {
	// Determine output binary name
	binaryName := cfg.Name
	if binaryName == "" {
		binaryName = filepath.Base(cfg.ProjectDir)
	}

	// Build the wails build arguments
	args := []string{"build"}

	// Platform
	args = append(args, "-platform", fmt.Sprintf("%s/%s", target.OS, target.Arch))

	// Output (Wails v2 uses -o for the binary name, relative to build/bin usually, but we want to control it)
	// Actually, Wails v2 is opinionated about output dir (build/bin).
	// We might need to copy artifacts after build if we want them in cfg.OutputDir.
	// For now, let's try to let Wails do its thing and find the artifact.

	// Create the command
	cmd := exec.CommandContext(ctx, "wails", args...)
	cmd.Dir = cfg.ProjectDir

	// Capture output for error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		return build.Artifact{}, fmt.Errorf("wails build failed: %w\nOutput: %s", err, string(output))
	}

	// Wails v2 typically outputs to build/bin
	// We need to move/copy it to our desired output dir

	// Construct the source path where Wails v2 puts the binary
	wailsOutputDir := filepath.Join(cfg.ProjectDir, "build", "bin")

	// Find the artifact in Wails output dir
	sourcePath, err := b.findArtifact(cfg.FS, wailsOutputDir, binaryName, target)
	if err != nil {
		return build.Artifact{}, fmt.Errorf("failed to find Wails v2 build artifact: %w", err)
	}

	// Move/Copy to our output dir
	// Create platform specific dir in our output
	platformDir := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_%s", target.OS, target.Arch))
	if err := cfg.FS.EnsureDir(platformDir); err != nil {
		return build.Artifact{}, fmt.Errorf("failed to create output dir: %w", err)
	}

	destPath := filepath.Join(platformDir, filepath.Base(sourcePath))

	// Simple copy using the medium
	content, err := cfg.FS.Read(sourcePath)
	if err != nil {
		return build.Artifact{}, err
	}
	if err := cfg.FS.Write(destPath, content); err != nil {
		return build.Artifact{}, err
	}

	return build.Artifact{
		Path: destPath,
		OS:   target.OS,
		Arch: target.Arch,
	}, nil
}

// findArtifact locates the built artifact based on the target platform.
func (b *WailsBuilder) findArtifact(fs io.Medium, platformDir, binaryName string, target build.Target) (string, error) {
	var candidates []string

	switch target.OS {
	case "windows":
		// Look for NSIS installer first, then plain exe
		candidates = []string{
			filepath.Join(platformDir, binaryName+"-installer.exe"),
			filepath.Join(platformDir, binaryName+".exe"),
			filepath.Join(platformDir, binaryName+"-amd64-installer.exe"),
		}
	case "darwin":
		// Look for .dmg, then .app bundle, then plain binary
		candidates = []string{
			filepath.Join(platformDir, binaryName+".dmg"),
			filepath.Join(platformDir, binaryName+".app"),
			filepath.Join(platformDir, binaryName),
		}
	default:
		// Linux and others: look for plain binary
		candidates = []string{
			filepath.Join(platformDir, binaryName),
		}
	}

	// Try each candidate
	for _, candidate := range candidates {
		if fs.Exists(candidate) {
			return candidate, nil
		}
	}

	// If no specific candidate found, try to find any executable or package in the directory
	entries, err := fs.List(platformDir)
	if err != nil {
		return "", fmt.Errorf("failed to read platform directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		// Skip common non-artifact files
		if strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".json") {
			continue
		}

		path := filepath.Join(platformDir, name)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// On Unix, check if it's executable; on Windows, check for .exe
		if target.OS == "windows" {
			if strings.HasSuffix(name, ".exe") {
				return path, nil
			}
		} else if info.Mode()&0111 != 0 || entry.IsDir() {
			// Executable file or directory (.app bundle)
			return path, nil
		}
	}

	return "", fmt.Errorf("no artifact found in %s", platformDir)
}

// detectPackageManager detects the frontend package manager based on lock files.
// Returns "bun", "pnpm", "yarn", or "npm" (default).
func detectPackageManager(fs io.Medium, dir string) string {
	// Check in priority order: bun, pnpm, yarn, npm
	lockFiles := []struct {
		file    string
		manager string
	}{
		{"bun.lockb", "bun"},
		{"pnpm-lock.yaml", "pnpm"},
		{"yarn.lock", "yarn"},
		{"package-lock.json", "npm"},
	}

	for _, lf := range lockFiles {
		if fs.IsFile(filepath.Join(dir, lf.file)) {
			return lf.manager
		}
	}

	// Default to npm if no lock file found
	return "npm"
}

// Ensure WailsBuilder implements the Builder interface.
var _ build.Builder = (*WailsBuilder)(nil)
