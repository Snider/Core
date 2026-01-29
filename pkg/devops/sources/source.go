// Package sources provides image download sources for core-devops.
package sources

import (
	"context"
)

// ImageSource defines the interface for downloading dev images.
type ImageSource interface {
	// Name returns the source identifier.
	Name() string
	// Available checks if this source can be used.
	Available() bool
	// LatestVersion returns the latest available version.
	LatestVersion(ctx context.Context) (string, error)
	// Download downloads the image to the destination path.
	// Reports progress via the callback if provided.
	Download(ctx context.Context, dest string, progress func(downloaded, total int64)) error
}

// SourceConfig holds configuration for a source.
type SourceConfig struct {
	// GitHub configuration
	GitHubRepo string
	// Registry configuration
	RegistryImage string
	// CDN configuration
	CDNURL string
	// Image name (e.g., core-devops-darwin-arm64.qcow2)
	ImageName string
}
