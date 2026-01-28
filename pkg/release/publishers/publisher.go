// Package publishers provides release publishing implementations.
package publishers

import (
	"context"

	"github.com/host-uk/core/pkg/build"
)

// Release represents a release to be published.
type Release struct {
	// Version is the semantic version string (e.g., "v1.2.3").
	Version string
	// Artifacts are the built release artifacts.
	Artifacts []build.Artifact
	// Changelog is the generated markdown changelog.
	Changelog string
	// ProjectDir is the root directory of the project.
	ProjectDir string
}

// PublisherConfig holds configuration for a publisher.
type PublisherConfig struct {
	// Type is the publisher type (e.g., "github", "linuxkit", "docker").
	Type string
	// Prerelease marks the release as a prerelease.
	Prerelease bool
	// Draft creates the release as a draft.
	Draft bool
	// Extended holds publisher-specific configuration.
	Extended any
}

// ReleaseConfig holds release configuration needed by publishers.
type ReleaseConfig interface {
	GetRepository() string
	GetProjectName() string
}

// Publisher defines the interface for release publishers.
type Publisher interface {
	// Name returns the publisher's identifier.
	Name() string
	// Publish publishes the release to the target.
	// If dryRun is true, it prints what would be done without executing.
	Publish(ctx context.Context, release *Release, pubCfg PublisherConfig, relCfg ReleaseConfig, dryRun bool) error
}

// NewRelease creates a Release from the release package's Release type.
// This is a helper to convert between packages.
func NewRelease(version string, artifacts []build.Artifact, changelog, projectDir string) *Release {
	return &Release{
		Version:    version,
		Artifacts:  artifacts,
		Changelog:  changelog,
		ProjectDir: projectDir,
	}
}

// NewPublisherConfig creates a PublisherConfig.
func NewPublisherConfig(pubType string, prerelease, draft bool, extended any) PublisherConfig {
	return PublisherConfig{
		Type:       pubType,
		Prerelease: prerelease,
		Draft:      draft,
		Extended:   extended,
	}
}
