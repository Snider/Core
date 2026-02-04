// Package publishers provides release publishing implementations.
package publishers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DockerConfig holds configuration for the Docker publisher.
type DockerConfig struct {
	// Registry is the container registry (default: ghcr.io).
	Registry string `yaml:"registry"`
	// Image is the image name in owner/repo format.
	Image string `yaml:"image"`
	// Dockerfile is the path to the Dockerfile (default: Dockerfile).
	Dockerfile string `yaml:"dockerfile"`
	// Platforms are the target platforms (linux/amd64, linux/arm64).
	Platforms []string `yaml:"platforms"`
	// Tags are additional tags to apply (supports {{.Version}} template).
	Tags []string `yaml:"tags"`
	// BuildArgs are additional build arguments.
	BuildArgs map[string]string `yaml:"build_args"`
}

// DockerPublisher builds and publishes Docker images.
type DockerPublisher struct{}

// NewDockerPublisher creates a new Docker publisher.
func NewDockerPublisher() *DockerPublisher {
	return &DockerPublisher{}
}

// Name returns the publisher's identifier.
func (p *DockerPublisher) Name() string {
	return "docker"
}

// Publish builds and pushes Docker images.
func (p *DockerPublisher) Publish(ctx context.Context, release *Release, pubCfg PublisherConfig, relCfg ReleaseConfig, dryRun bool) error {
	// Validate docker CLI is available
	if err := validateDockerCli(); err != nil {
		return err
	}

	// Parse Docker-specific config from publisher config
	dockerCfg := p.parseConfig(pubCfg, relCfg, release.ProjectDir)

	// Validate Dockerfile exists
	if !release.FS.Exists(dockerCfg.Dockerfile) {
		return fmt.Errorf("docker.Publish: Dockerfile not found: %s", dockerCfg.Dockerfile)
	}

	if dryRun {
		return p.dryRunPublish(release, dockerCfg)
	}

	return p.executePublish(ctx, release, dockerCfg)
}

// parseConfig extracts Docker-specific configuration.
func (p *DockerPublisher) parseConfig(pubCfg PublisherConfig, relCfg ReleaseConfig, projectDir string) DockerConfig {
	cfg := DockerConfig{
		Registry:   "ghcr.io",
		Image:      "",
		Dockerfile: filepath.Join(projectDir, "Dockerfile"),
		Platforms:  []string{"linux/amd64", "linux/arm64"},
		Tags:       []string{"latest", "{{.Version}}"},
		BuildArgs:  make(map[string]string),
	}

	// Try to get image from repository config
	if relCfg != nil && relCfg.GetRepository() != "" {
		cfg.Image = relCfg.GetRepository()
	}

	// Override from extended config if present
	if ext, ok := pubCfg.Extended.(map[string]any); ok {
		if registry, ok := ext["registry"].(string); ok && registry != "" {
			cfg.Registry = registry
		}
		if image, ok := ext["image"].(string); ok && image != "" {
			cfg.Image = image
		}
		if dockerfile, ok := ext["dockerfile"].(string); ok && dockerfile != "" {
			if filepath.IsAbs(dockerfile) {
				cfg.Dockerfile = dockerfile
			} else {
				cfg.Dockerfile = filepath.Join(projectDir, dockerfile)
			}
		}
		if platforms, ok := ext["platforms"].([]any); ok && len(platforms) > 0 {
			cfg.Platforms = make([]string, 0, len(platforms))
			for _, plat := range platforms {
				if s, ok := plat.(string); ok {
					cfg.Platforms = append(cfg.Platforms, s)
				}
			}
		}
		if tags, ok := ext["tags"].([]any); ok && len(tags) > 0 {
			cfg.Tags = make([]string, 0, len(tags))
			for _, tag := range tags {
				if s, ok := tag.(string); ok {
					cfg.Tags = append(cfg.Tags, s)
				}
			}
		}
		if buildArgs, ok := ext["build_args"].(map[string]any); ok {
			for k, v := range buildArgs {
				if s, ok := v.(string); ok {
					cfg.BuildArgs[k] = s
				}
			}
		}
	}

	return cfg
}

// dryRunPublish shows what would be done without actually building.
func (p *DockerPublisher) dryRunPublish(release *Release, cfg DockerConfig) error {
	fmt.Println()
	fmt.Println("=== DRY RUN: Docker Build & Push ===")
	fmt.Println()
	fmt.Printf("Version:       %s\n", release.Version)
	fmt.Printf("Registry:      %s\n", cfg.Registry)
	fmt.Printf("Image:         %s\n", cfg.Image)
	fmt.Printf("Dockerfile:    %s\n", cfg.Dockerfile)
	fmt.Printf("Platforms:     %s\n", strings.Join(cfg.Platforms, ", "))
	fmt.Println()

	// Resolve tags
	tags := p.resolveTags(cfg.Tags, release.Version)
	fmt.Println("Tags to be applied:")
	for _, tag := range tags {
		fullTag := p.buildFullTag(cfg.Registry, cfg.Image, tag)
		fmt.Printf("  - %s\n", fullTag)
	}
	fmt.Println()

	fmt.Println("Would execute command:")
	args := p.buildBuildxArgs(cfg, tags, release.Version)
	fmt.Printf("  docker %s\n", strings.Join(args, " "))

	if len(cfg.BuildArgs) > 0 {
		fmt.Println()
		fmt.Println("Build arguments:")
		for k, v := range cfg.BuildArgs {
			fmt.Printf("  %s=%s\n", k, v)
		}
	}

	fmt.Println()
	fmt.Println("=== END DRY RUN ===")

	return nil
}

// executePublish builds and pushes Docker images.
func (p *DockerPublisher) executePublish(ctx context.Context, release *Release, cfg DockerConfig) error {
	// Ensure buildx is available and builder is set up
	if err := p.ensureBuildx(ctx); err != nil {
		return err
	}

	// Resolve tags
	tags := p.resolveTags(cfg.Tags, release.Version)

	// Build the docker buildx command
	args := p.buildBuildxArgs(cfg, tags, release.Version)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = release.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Building and pushing Docker image: %s\n", cfg.Image)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker.Publish: buildx build failed: %w", err)
	}

	return nil
}

// resolveTags expands template variables in tags.
func (p *DockerPublisher) resolveTags(tags []string, version string) []string {
	resolved := make([]string, 0, len(tags))
	for _, tag := range tags {
		// Replace {{.Version}} with actual version
		resolvedTag := strings.ReplaceAll(tag, "{{.Version}}", version)
		// Also support simpler {{Version}} syntax
		resolvedTag = strings.ReplaceAll(resolvedTag, "{{Version}}", version)
		resolved = append(resolved, resolvedTag)
	}
	return resolved
}

// buildFullTag builds the full image tag including registry.
func (p *DockerPublisher) buildFullTag(registry, image, tag string) string {
	if registry != "" {
		return fmt.Sprintf("%s/%s:%s", registry, image, tag)
	}
	return fmt.Sprintf("%s:%s", image, tag)
}

// buildBuildxArgs builds the arguments for docker buildx build command.
func (p *DockerPublisher) buildBuildxArgs(cfg DockerConfig, tags []string, version string) []string {
	args := []string{"buildx", "build"}

	// Multi-platform support
	if len(cfg.Platforms) > 0 {
		args = append(args, "--platform", strings.Join(cfg.Platforms, ","))
	}

	// Add all tags
	for _, tag := range tags {
		fullTag := p.buildFullTag(cfg.Registry, cfg.Image, tag)
		args = append(args, "-t", fullTag)
	}

	// Dockerfile path
	dockerfilePath := cfg.Dockerfile
	args = append(args, "-f", dockerfilePath)

	// Build arguments
	for k, v := range cfg.BuildArgs {
		// Expand version in build args
		expandedValue := strings.ReplaceAll(v, "{{.Version}}", version)
		expandedValue = strings.ReplaceAll(expandedValue, "{{Version}}", version)
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, expandedValue))
	}

	// Always add VERSION build arg
	args = append(args, "--build-arg", fmt.Sprintf("VERSION=%s", version))

	// Push the image
	args = append(args, "--push")

	// Build context (current directory)
	args = append(args, ".")

	return args
}

// ensureBuildx ensures docker buildx is available and has a builder.
func (p *DockerPublisher) ensureBuildx(ctx context.Context) error {
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

// validateDockerCli checks if the docker CLI is available.
func validateDockerCli() error {
	cmd := exec.Command("docker", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker: docker CLI not found. Install it from https://docs.docker.com/get-docker/")
	}
	return nil
}
