// Package bugseti provides version information for the BugSETI application.
package bugseti

import (
	"fmt"
	"runtime"
)

// Version information - these are set at build time via ldflags
// Example: go build -ldflags "-X github.com/host-uk/core/internal/bugseti.Version=1.0.0"
var (
	// Version is the semantic version (e.g., "1.0.0", "1.0.0-beta.1", "nightly-20260205")
	Version = "dev"

	// Channel is the release channel (stable, beta, nightly)
	Channel = "dev"

	// Commit is the git commit SHA
	Commit = "unknown"

	// BuildTime is the UTC build timestamp
	BuildTime = "unknown"
)

// VersionInfo contains all version-related information.
type VersionInfo struct {
	Version   string `json:"version"`
	Channel   string `json:"channel"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
	GoVersion string `json:"goVersion"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// GetVersion returns the current version string.
func GetVersion() string {
	return Version
}

// GetChannel returns the release channel.
func GetChannel() string {
	return Channel
}

// GetVersionInfo returns complete version information.
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		Channel:   Channel,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// GetVersionString returns a formatted version string for display.
func GetVersionString() string {
	if Channel == "dev" {
		return fmt.Sprintf("BugSETI %s (development build)", Version)
	}
	if Channel == "nightly" {
		return fmt.Sprintf("BugSETI %s (nightly)", Version)
	}
	if Channel == "beta" {
		return fmt.Sprintf("BugSETI v%s (beta)", Version)
	}
	return fmt.Sprintf("BugSETI v%s", Version)
}

// GetShortCommit returns the first 7 characters of the commit hash.
func GetShortCommit() string {
	if len(Commit) >= 7 {
		return Commit[:7]
	}
	return Commit
}

// IsDevelopment returns true if this is a development build.
func IsDevelopment() bool {
	return Channel == "dev" || Version == "dev"
}

// IsPrerelease returns true if this is a prerelease build (beta or nightly).
func IsPrerelease() bool {
	return Channel == "beta" || Channel == "nightly"
}

// VersionService provides version information to the frontend via Wails.
type VersionService struct{}

// NewVersionService creates a new VersionService.
func NewVersionService() *VersionService {
	return &VersionService{}
}

// ServiceName returns the service name for Wails.
func (v *VersionService) ServiceName() string {
	return "VersionService"
}

// GetVersion returns the version string.
func (v *VersionService) GetVersion() string {
	return GetVersion()
}

// GetChannel returns the release channel.
func (v *VersionService) GetChannel() string {
	return GetChannel()
}

// GetVersionInfo returns complete version information.
func (v *VersionService) GetVersionInfo() VersionInfo {
	return GetVersionInfo()
}

// GetVersionString returns a formatted version string.
func (v *VersionService) GetVersionString() string {
	return GetVersionString()
}
