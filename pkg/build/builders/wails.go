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
func (b *WailsBuilder) Detect(dir string) (bool, error) {
	return build.IsWailsProject(dir), nil
}

// Build compiles the Wails project for the specified targets.
// It installs frontend dependencies, builds the frontend, then runs wails3 build.
func (b *WailsBuilder) Build(ctx context.Context, cfg *build.Config, targets []build.Target) ([]build.Artifact, error) {
	if cfg == nil {
		return nil, fmt.Errorf("builders.WailsBuilder.Build: config is nil")
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("builders.WailsBuilder.Build: no targets specified")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("builders.WailsBuilder.Build: failed to create output directory: %w", err)
	}

	// Find frontend directory (typically "frontend")
	frontendDir := filepath.Join(cfg.ProjectDir, "frontend")
	hasFrontend := dirExists(frontendDir)

	if hasFrontend {
		// Detect package manager
		pkgManager := detectPackageManager(frontendDir)

		// Install frontend dependencies if node_modules is missing
		nodeModules := filepath.Join(frontendDir, "node_modules")
		if !dirExists(nodeModules) {
			if err := b.installFrontendDeps(ctx, frontendDir, pkgManager); err != nil {
				return nil, fmt.Errorf("builders.WailsBuilder.Build: failed to install frontend dependencies: %w", err)
			}
		}

		// Build frontend
		if err := b.buildFrontend(ctx, frontendDir, pkgManager); err != nil {
			return nil, fmt.Errorf("builders.WailsBuilder.Build: failed to build frontend: %w", err)
		}
	}

	var artifacts []build.Artifact

	for _, target := range targets {
		artifact, err := b.buildTarget(ctx, cfg, target)
		if err != nil {
			return artifacts, fmt.Errorf("builders.WailsBuilder.Build: failed to build %s: %w", target.String(), err)
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

// installFrontendDeps installs frontend dependencies using the detected package manager.
func (b *WailsBuilder) installFrontendDeps(ctx context.Context, frontendDir, pkgManager string) error {
	var cmd *exec.Cmd

	switch pkgManager {
	case "bun":
		cmd = exec.CommandContext(ctx, "bun", "install")
	case "pnpm":
		cmd = exec.CommandContext(ctx, "pnpm", "install")
	case "yarn":
		cmd = exec.CommandContext(ctx, "yarn", "install")
	default:
		cmd = exec.CommandContext(ctx, "npm", "install")
	}

	cmd.Dir = frontendDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s install failed: %w\nOutput: %s", pkgManager, err, string(output))
	}

	return nil
}

// buildFrontend runs the frontend build command using the detected package manager.
func (b *WailsBuilder) buildFrontend(ctx context.Context, frontendDir, pkgManager string) error {
	var cmd *exec.Cmd

	switch pkgManager {
	case "bun":
		cmd = exec.CommandContext(ctx, "bun", "run", "build")
	case "pnpm":
		cmd = exec.CommandContext(ctx, "pnpm", "run", "build")
	case "yarn":
		cmd = exec.CommandContext(ctx, "yarn", "run", "build")
	default:
		cmd = exec.CommandContext(ctx, "npm", "run", "build")
	}

	cmd.Dir = frontendDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s run build failed: %w\nOutput: %s", pkgManager, err, string(output))
	}

	return nil
}

// buildTarget compiles for a single target platform using wails3.
func (b *WailsBuilder) buildTarget(ctx context.Context, cfg *build.Config, target build.Target) (build.Artifact, error) {
	// Determine output binary name
	binaryName := cfg.Name
	if binaryName == "" {
		binaryName = filepath.Base(cfg.ProjectDir)
	}

	// Create platform-specific output path: output/os_arch/
	platformDir := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_%s", target.OS, target.Arch))
	if err := os.MkdirAll(platformDir, 0755); err != nil {
		return build.Artifact{}, fmt.Errorf("failed to create platform directory: %w", err)
	}

	// Build the wails3 build arguments
	args := []string{"build"}

	// Add output directory
	args = append(args, "-o", platformDir)

	// Create the command
	cmd := exec.CommandContext(ctx, "wails3", args...)
	cmd.Dir = cfg.ProjectDir

	// Set up environment for cross-compilation
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", target.OS))
	env = append(env, fmt.Sprintf("GOARCH=%s", target.Arch))
	cmd.Env = env

	// Capture output for error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		return build.Artifact{}, fmt.Errorf("wails3 build failed: %w\nOutput: %s", err, string(output))
	}

	// Find the built artifact - depends on platform
	artifactPath, err := b.findArtifact(platformDir, binaryName, target)
	if err != nil {
		return build.Artifact{}, fmt.Errorf("failed to find build artifact: %w", err)
	}

	return build.Artifact{
		Path: artifactPath,
		OS:   target.OS,
		Arch: target.Arch,
	}, nil
}

// findArtifact locates the built artifact based on the target platform.
func (b *WailsBuilder) findArtifact(platformDir, binaryName string, target build.Target) (string, error) {
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
		if fileOrDirExists(candidate) {
			return candidate, nil
		}
	}

	// If no specific candidate found, try to find any executable or package in the directory
	entries, err := os.ReadDir(platformDir)
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
func detectPackageManager(dir string) string {
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
		if fileExists(filepath.Join(dir, lf.file)) {
			return lf.manager
		}
	}

	// Default to npm if no lock file found
	return "npm"
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// fileOrDirExists checks if a file or directory exists.
func fileOrDirExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Ensure WailsBuilder implements the Builder interface.
var _ build.Builder = (*WailsBuilder)(nil)
