// Package release provides release automation with changelog generation and publishing.
// It orchestrates the build system, changelog generation, and publishing to targets
// like GitHub Releases.
package release

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/host-uk/core/pkg/build"
	"github.com/host-uk/core/pkg/build/builders"
	"github.com/host-uk/core/pkg/release/publishers"
)

// Release represents a release with its version, artifacts, and changelog.
type Release struct {
	// Version is the semantic version string (e.g., "v1.2.3").
	Version string
	// Artifacts are the built release artifacts (archives with checksums).
	Artifacts []build.Artifact
	// Changelog is the generated markdown changelog.
	Changelog string
	// ProjectDir is the root directory of the project.
	ProjectDir string
}

// Run executes the release process: determine version, build artifacts,
// generate changelog, and publish to configured targets.
// If dryRun is true, it will show what would be done without actually publishing.
func Run(ctx context.Context, cfg *Config, dryRun bool) (*Release, error) {
	if cfg == nil {
		return nil, fmt.Errorf("release.Run: config is nil")
	}

	projectDir := cfg.projectDir
	if projectDir == "" {
		projectDir = "."
	}

	// Resolve to absolute path
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, fmt.Errorf("release.Run: failed to resolve project directory: %w", err)
	}

	// Step 1: Determine version
	version := cfg.version
	if version == "" {
		version, err = DetermineVersion(absProjectDir)
		if err != nil {
			return nil, fmt.Errorf("release.Run: failed to determine version: %w", err)
		}
	}

	// Step 2: Generate changelog
	changelog, err := Generate(absProjectDir, "", version)
	if err != nil {
		// Non-fatal: continue with empty changelog
		changelog = fmt.Sprintf("Release %s", version)
	}

	// Step 3: Build artifacts
	artifacts, err := buildArtifacts(ctx, cfg, absProjectDir, version)
	if err != nil {
		return nil, fmt.Errorf("release.Run: build failed: %w", err)
	}

	release := &Release{
		Version:    version,
		Artifacts:  artifacts,
		Changelog:  changelog,
		ProjectDir: absProjectDir,
	}

	// Step 4: Publish to configured targets
	if len(cfg.Publishers) > 0 {
		// Convert to publisher types
		pubRelease := publishers.NewRelease(release.Version, release.Artifacts, release.Changelog, release.ProjectDir)

		for _, pubCfg := range cfg.Publishers {
			publisher, err := getPublisher(pubCfg.Type)
			if err != nil {
				return release, fmt.Errorf("release.Run: %w", err)
			}

			// Build extended config for publisher-specific settings
			extendedCfg := buildExtendedConfig(pubCfg)
			publisherCfg := publishers.NewPublisherConfig(pubCfg.Type, pubCfg.Prerelease, pubCfg.Draft, extendedCfg)
			if err := publisher.Publish(ctx, pubRelease, publisherCfg, cfg, dryRun); err != nil {
				return release, fmt.Errorf("release.Run: publish to %s failed: %w", pubCfg.Type, err)
			}
		}
	}

	return release, nil
}

// buildArtifacts builds all artifacts for the release.
func buildArtifacts(ctx context.Context, cfg *Config, projectDir, version string) ([]build.Artifact, error) {
	// Load build configuration
	buildCfg, err := build.LoadConfig(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load build config: %w", err)
	}

	// Determine targets
	var targets []build.Target
	if len(cfg.Build.Targets) > 0 {
		for _, t := range cfg.Build.Targets {
			targets = append(targets, build.Target{OS: t.OS, Arch: t.Arch})
		}
	} else if len(buildCfg.Targets) > 0 {
		targets = buildCfg.ToTargets()
	} else {
		// Default targets
		targets = []build.Target{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
			{OS: "darwin", Arch: "amd64"},
			{OS: "darwin", Arch: "arm64"},
			{OS: "windows", Arch: "amd64"},
		}
	}

	// Determine binary name
	binaryName := cfg.Project.Name
	if binaryName == "" {
		binaryName = buildCfg.Project.Binary
	}
	if binaryName == "" {
		binaryName = buildCfg.Project.Name
	}
	if binaryName == "" {
		binaryName = filepath.Base(projectDir)
	}

	// Determine output directory
	outputDir := filepath.Join(projectDir, "dist")

	// Get builder (detect project type)
	projectType, err := build.PrimaryType(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to detect project type: %w", err)
	}

	builder, err := getBuilder(projectType)
	if err != nil {
		return nil, err
	}

	// Build configuration
	buildConfig := &build.Config{
		ProjectDir: projectDir,
		OutputDir:  outputDir,
		Name:       binaryName,
		Version:    version,
		LDFlags:    buildCfg.Build.LDFlags,
	}

	// Build
	artifacts, err := builder.Build(ctx, buildConfig, targets)
	if err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	// Archive artifacts
	archivedArtifacts, err := build.ArchiveAll(artifacts)
	if err != nil {
		return nil, fmt.Errorf("archive failed: %w", err)
	}

	// Compute checksums
	checksummedArtifacts, err := build.ChecksumAll(archivedArtifacts)
	if err != nil {
		return nil, fmt.Errorf("checksum failed: %w", err)
	}

	// Write CHECKSUMS.txt
	checksumPath := filepath.Join(outputDir, "CHECKSUMS.txt")
	if err := build.WriteChecksumFile(checksummedArtifacts, checksumPath); err != nil {
		return nil, fmt.Errorf("failed to write checksums file: %w", err)
	}

	// Add CHECKSUMS.txt as an artifact
	checksumArtifact := build.Artifact{
		Path: checksumPath,
	}
	checksummedArtifacts = append(checksummedArtifacts, checksumArtifact)

	return checksummedArtifacts, nil
}

// getBuilder returns the appropriate builder for the project type.
func getBuilder(projectType build.ProjectType) (build.Builder, error) {
	switch projectType {
	case build.ProjectTypeWails:
		return builders.NewWailsBuilder(), nil
	case build.ProjectTypeGo:
		return builders.NewGoBuilder(), nil
	case build.ProjectTypeNode:
		return nil, fmt.Errorf("Node.js builder not yet implemented")
	case build.ProjectTypePHP:
		return nil, fmt.Errorf("PHP builder not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported project type: %s", projectType)
	}
}

// getPublisher returns the publisher for the given type.
func getPublisher(pubType string) (publishers.Publisher, error) {
	switch pubType {
	case "github":
		return publishers.NewGitHubPublisher(), nil
	case "linuxkit":
		return publishers.NewLinuxKitPublisher(), nil
	case "docker":
		return publishers.NewDockerPublisher(), nil
	default:
		return nil, fmt.Errorf("unsupported publisher type: %s", pubType)
	}
}

// buildExtendedConfig builds a map of extended configuration for a publisher.
func buildExtendedConfig(pubCfg PublisherConfig) map[string]any {
	ext := make(map[string]any)

	// LinuxKit-specific config
	if pubCfg.Config != "" {
		ext["config"] = pubCfg.Config
	}
	if len(pubCfg.Formats) > 0 {
		ext["formats"] = toAnySlice(pubCfg.Formats)
	}
	if len(pubCfg.Platforms) > 0 {
		ext["platforms"] = toAnySlice(pubCfg.Platforms)
	}

	// Docker-specific config
	if pubCfg.Registry != "" {
		ext["registry"] = pubCfg.Registry
	}
	if pubCfg.Image != "" {
		ext["image"] = pubCfg.Image
	}
	if pubCfg.Dockerfile != "" {
		ext["dockerfile"] = pubCfg.Dockerfile
	}
	if len(pubCfg.Tags) > 0 {
		ext["tags"] = toAnySlice(pubCfg.Tags)
	}
	if len(pubCfg.BuildArgs) > 0 {
		args := make(map[string]any)
		for k, v := range pubCfg.BuildArgs {
			args[k] = v
		}
		ext["build_args"] = args
	}

	return ext
}

// toAnySlice converts a string slice to an any slice.
func toAnySlice(s []string) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}
