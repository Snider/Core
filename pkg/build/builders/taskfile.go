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

// TaskfileBuilder builds projects using Taskfile (https://taskfile.dev/).
// This is a generic builder that can handle any project type that has a Taskfile.
type TaskfileBuilder struct{}

// NewTaskfileBuilder creates a new Taskfile builder.
func NewTaskfileBuilder() *TaskfileBuilder {
	return &TaskfileBuilder{}
}

// Name returns the builder's identifier.
func (b *TaskfileBuilder) Name() string {
	return "taskfile"
}

// Detect checks if a Taskfile exists in the directory.
func (b *TaskfileBuilder) Detect(dir string) (bool, error) {
	// Check for Taskfile.yml, Taskfile.yaml, or Taskfile
	taskfiles := []string{
		"Taskfile.yml",
		"Taskfile.yaml",
		"Taskfile",
		"taskfile.yml",
		"taskfile.yaml",
	}

	for _, tf := range taskfiles {
		if _, err := os.Stat(filepath.Join(dir, tf)); err == nil {
			return true, nil
		}
	}
	return false, nil
}

// Build runs the Taskfile build task for each target platform.
func (b *TaskfileBuilder) Build(ctx context.Context, cfg *build.Config, targets []build.Target) ([]build.Artifact, error) {
	// Validate task CLI is available
	if err := b.validateTaskCli(); err != nil {
		return nil, err
	}

	// Create output directory
	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join(cfg.ProjectDir, "dist")
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("taskfile.Build: failed to create output directory: %w", err)
	}

	var artifacts []build.Artifact

	// If no targets specified, just run the build task once
	if len(targets) == 0 {
		if err := b.runTask(ctx, cfg, "", ""); err != nil {
			return nil, err
		}

		// Try to find artifacts in output directory
		found := b.findArtifacts(outputDir)
		artifacts = append(artifacts, found...)
	} else {
		// Run build task for each target
		for _, target := range targets {
			if err := b.runTask(ctx, cfg, target.OS, target.Arch); err != nil {
				return nil, err
			}

			// Try to find artifacts for this target
			found := b.findArtifactsForTarget(outputDir, target)
			artifacts = append(artifacts, found...)
		}
	}

	return artifacts, nil
}

// runTask executes the Taskfile build task.
func (b *TaskfileBuilder) runTask(ctx context.Context, cfg *build.Config, goos, goarch string) error {
	// Build task command
	args := []string{"build"}

	// Pass variables if targets are specified
	if goos != "" {
		args = append(args, fmt.Sprintf("GOOS=%s", goos))
	}
	if goarch != "" {
		args = append(args, fmt.Sprintf("GOARCH=%s", goarch))
	}
	if cfg.OutputDir != "" {
		args = append(args, fmt.Sprintf("OUTPUT_DIR=%s", cfg.OutputDir))
	}
	if cfg.Name != "" {
		args = append(args, fmt.Sprintf("NAME=%s", cfg.Name))
	}
	if cfg.Version != "" {
		args = append(args, fmt.Sprintf("VERSION=%s", cfg.Version))
	}

	cmd := exec.CommandContext(ctx, "task", args...)
	cmd.Dir = cfg.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment variables
	cmd.Env = os.Environ()
	if goos != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", goos))
	}
	if goarch != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", goarch))
	}
	if cfg.OutputDir != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("OUTPUT_DIR=%s", cfg.OutputDir))
	}
	if cfg.Name != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("NAME=%s", cfg.Name))
	}
	if cfg.Version != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("VERSION=%s", cfg.Version))
	}

	if goos != "" && goarch != "" {
		fmt.Printf("Running task build for %s/%s\n", goos, goarch)
	} else {
		fmt.Println("Running task build")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("taskfile.Build: task build failed: %w", err)
	}

	return nil
}

// findArtifacts searches for built artifacts in the output directory.
func (b *TaskfileBuilder) findArtifacts(outputDir string) []build.Artifact {
	var artifacts []build.Artifact

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return artifacts
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip common non-artifact files
		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == "CHECKSUMS.txt" {
			continue
		}

		artifacts = append(artifacts, build.Artifact{
			Path: filepath.Join(outputDir, name),
			OS:   "",
			Arch: "",
		})
	}

	return artifacts
}

// findArtifactsForTarget searches for built artifacts for a specific target.
func (b *TaskfileBuilder) findArtifactsForTarget(outputDir string, target build.Target) []build.Artifact {
	var artifacts []build.Artifact

	// Look for files matching the target pattern
	patterns := []string{
		fmt.Sprintf("*-%s-%s*", target.OS, target.Arch),
		fmt.Sprintf("*_%s_%s*", target.OS, target.Arch),
		fmt.Sprintf("*-%s*", target.Arch),
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(outputDir, pattern))
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}

			artifacts = append(artifacts, build.Artifact{
				Path: match,
				OS:   target.OS,
				Arch: target.Arch,
			})
		}

		if len(artifacts) > 0 {
			break // Found matches, stop looking
		}
	}

	return artifacts
}

// validateTaskCli checks if the task CLI is available.
func (b *TaskfileBuilder) validateTaskCli() error {
	// Check PATH first
	if _, err := exec.LookPath("task"); err == nil {
		return nil
	}

	// Check common locations
	paths := []string{
		"/usr/local/bin/task",
		"/opt/homebrew/bin/task",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return nil
		}
	}

	return fmt.Errorf("taskfile: task CLI not found. Install with: brew install go-task (macOS), go install github.com/go-task/task/v3/cmd/task@latest, or see https://taskfile.dev/installation/")
}
