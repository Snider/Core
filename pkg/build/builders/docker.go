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

// DockerBuilder builds Docker images.
type DockerBuilder struct{}

// NewDockerBuilder creates a new Docker builder.
func NewDockerBuilder() *DockerBuilder {
	return &DockerBuilder{}
}

// Name returns the builder's identifier.
func (b *DockerBuilder) Name() string {
	return "docker"
}

// Detect checks if a Dockerfile exists in the directory.
func (b *DockerBuilder) Detect(dir string) (bool, error) {
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		return true, nil
	}
	return false, nil
}

// Build builds Docker images for the specified targets.
func (b *DockerBuilder) Build(ctx context.Context, cfg *build.Config, targets []build.Target) ([]build.Artifact, error) {
	// Validate docker CLI is available
	if err := b.validateDockerCli(); err != nil {
		return nil, err
	}

	// Ensure buildx is available
	if err := b.ensureBuildx(ctx); err != nil {
		return nil, err
	}

	// Determine Dockerfile path
	dockerfile := cfg.Dockerfile
	if dockerfile == "" {
		dockerfile = filepath.Join(cfg.ProjectDir, "Dockerfile")
	}

	// Validate Dockerfile exists
	if _, err := os.Stat(dockerfile); err != nil {
		return nil, fmt.Errorf("docker.Build: Dockerfile not found: %s", dockerfile)
	}

	// Determine image name
	imageName := cfg.Image
	if imageName == "" {
		imageName = cfg.Name
	}
	if imageName == "" {
		imageName = filepath.Base(cfg.ProjectDir)
	}

	// Build platform string from targets
	var platforms []string
	for _, t := range targets {
		platforms = append(platforms, fmt.Sprintf("%s/%s", t.OS, t.Arch))
	}

	// If no targets specified, use current platform
	if len(platforms) == 0 {
		platforms = []string{"linux/amd64"}
	}

	// Determine registry
	registry := cfg.Registry
	if registry == "" {
		registry = "ghcr.io"
	}

	// Determine tags
	tags := cfg.Tags
	if len(tags) == 0 {
		tags = []string{"latest"}
		if cfg.Version != "" {
			tags = append(tags, cfg.Version)
		}
	}

	// Build full image references
	var imageRefs []string
	for _, tag := range tags {
		// Expand version template
		expandedTag := strings.ReplaceAll(tag, "{{.Version}}", cfg.Version)
		expandedTag = strings.ReplaceAll(expandedTag, "{{Version}}", cfg.Version)

		if registry != "" {
			imageRefs = append(imageRefs, fmt.Sprintf("%s/%s:%s", registry, imageName, expandedTag))
		} else {
			imageRefs = append(imageRefs, fmt.Sprintf("%s:%s", imageName, expandedTag))
		}
	}

	// Build the docker buildx command
	args := []string{"buildx", "build"}

	// Multi-platform support
	args = append(args, "--platform", strings.Join(platforms, ","))

	// Add all tags
	for _, ref := range imageRefs {
		args = append(args, "-t", ref)
	}

	// Dockerfile path
	args = append(args, "-f", dockerfile)

	// Build arguments
	for k, v := range cfg.BuildArgs {
		expandedValue := strings.ReplaceAll(v, "{{.Version}}", cfg.Version)
		expandedValue = strings.ReplaceAll(expandedValue, "{{Version}}", cfg.Version)
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, expandedValue))
	}

	// Always add VERSION build arg if version is set
	if cfg.Version != "" {
		args = append(args, "--build-arg", fmt.Sprintf("VERSION=%s", cfg.Version))
	}

	// Output to local docker images or push
	if cfg.Push {
		args = append(args, "--push")
	} else {
		// For multi-platform builds without push, we need to load or output somewhere
		if len(platforms) == 1 {
			args = append(args, "--load")
		} else {
			// Multi-platform builds can't use --load, output to tarball
			outputPath := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s.tar", imageName))
			args = append(args, "--output", fmt.Sprintf("type=oci,dest=%s", outputPath))
		}
	}

	// Build context (project directory)
	args = append(args, cfg.ProjectDir)

	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("docker.Build: failed to create output directory: %w", err)
	}

	// Execute build
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = cfg.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Building Docker image: %s\n", imageName)
	fmt.Printf("  Platforms: %s\n", strings.Join(platforms, ", "))
	fmt.Printf("  Tags: %s\n", strings.Join(imageRefs, ", "))

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("docker.Build: buildx build failed: %w", err)
	}

	// Create artifacts for each platform
	var artifacts []build.Artifact
	for _, t := range targets {
		artifacts = append(artifacts, build.Artifact{
			Path: imageRefs[0], // Primary image reference
			OS:   t.OS,
			Arch: t.Arch,
		})
	}

	return artifacts, nil
}

// validateDockerCli checks if the docker CLI is available.
func (b *DockerBuilder) validateDockerCli() error {
	cmd := exec.Command("docker", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker: docker CLI not found. Install it from https://docs.docker.com/get-docker/")
	}
	return nil
}

// ensureBuildx ensures docker buildx is available and has a builder.
func (b *DockerBuilder) ensureBuildx(ctx context.Context) error {
	// Check if buildx is available
	cmd := exec.CommandContext(ctx, "docker", "buildx", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker: buildx is not available. Install it from https://docs.docker.com/buildx/working-with-buildx/")
	}

	// Check if we have a builder, create one if not
	cmd = exec.CommandContext(ctx, "docker", "buildx", "inspect", "--bootstrap")
	if err := cmd.Run(); err != nil {
		// Try to create a builder
		cmd = exec.CommandContext(ctx, "docker", "buildx", "create", "--use", "--bootstrap")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("docker: failed to create buildx builder: %w", err)
		}
	}

	return nil
}
