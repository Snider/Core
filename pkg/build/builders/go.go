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

// GoBuilder implements the Builder interface for Go projects.
type GoBuilder struct{}

// NewGoBuilder creates a new GoBuilder instance.
func NewGoBuilder() *GoBuilder {
	return &GoBuilder{}
}

// Name returns the builder's identifier.
func (b *GoBuilder) Name() string {
	return "go"
}

// Detect checks if this builder can handle the project in the given directory.
// Uses IsGoProject from the build package which checks for go.mod or wails.json.
func (b *GoBuilder) Detect(dir string) (bool, error) {
	return build.IsGoProject(dir), nil
}

// Build compiles the Go project for the specified targets.
// It sets GOOS, GOARCH, and CGO_ENABLED environment variables,
// applies ldflags and trimpath, and runs go build.
func (b *GoBuilder) Build(ctx context.Context, cfg *build.Config, targets []build.Target) ([]build.Artifact, error) {
	if cfg == nil {
		return nil, fmt.Errorf("builders.GoBuilder.Build: config is nil")
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("builders.GoBuilder.Build: no targets specified")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("builders.GoBuilder.Build: failed to create output directory: %w", err)
	}

	var artifacts []build.Artifact

	for _, target := range targets {
		artifact, err := b.buildTarget(ctx, cfg, target)
		if err != nil {
			return artifacts, fmt.Errorf("builders.GoBuilder.Build: failed to build %s: %w", target.String(), err)
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

// buildTarget compiles for a single target platform.
func (b *GoBuilder) buildTarget(ctx context.Context, cfg *build.Config, target build.Target) (build.Artifact, error) {
	// Determine output binary name
	binaryName := cfg.Name
	if binaryName == "" {
		binaryName = filepath.Base(cfg.ProjectDir)
	}

	// Add .exe extension for Windows
	if target.OS == "windows" && !strings.HasSuffix(binaryName, ".exe") {
		binaryName += ".exe"
	}

	// Create platform-specific output path: output/os_arch/binary
	platformDir := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_%s", target.OS, target.Arch))
	if err := os.MkdirAll(platformDir, 0755); err != nil {
		return build.Artifact{}, fmt.Errorf("failed to create platform directory: %w", err)
	}

	outputPath := filepath.Join(platformDir, binaryName)

	// Build the go build arguments
	args := []string{"build"}

	// Add trimpath flag
	args = append(args, "-trimpath")

	// Add ldflags if specified
	if len(cfg.LDFlags) > 0 {
		ldflags := strings.Join(cfg.LDFlags, " ")
		args = append(args, "-ldflags", ldflags)
	}

	// Add output path
	args = append(args, "-o", outputPath)

	// Add the project directory as the build target (current directory)
	args = append(args, ".")

	// Create the command
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = cfg.ProjectDir

	// Set up environment
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", target.OS))
	env = append(env, fmt.Sprintf("GOARCH=%s", target.Arch))
	env = append(env, "CGO_ENABLED=0") // CGO disabled by default for cross-compilation
	cmd.Env = env

	// Capture output for error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		return build.Artifact{}, fmt.Errorf("go build failed: %w\nOutput: %s", err, string(output))
	}

	return build.Artifact{
		Path: outputPath,
		OS:   target.OS,
		Arch: target.Arch,
	}, nil
}

// Ensure GoBuilder implements the Builder interface.
var _ build.Builder = (*GoBuilder)(nil)
