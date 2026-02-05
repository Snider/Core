// Package build provides project type detection and cross-compilation for the Core build system.
// It supports Go, Wails, Node.js, and PHP projects with automatic detection based on
// marker files (go.mod, wails.json, package.json, composer.json).
package build

import (
	"context"
)

// ProjectType represents a detected project type.
type ProjectType string

// Project type constants for build detection.
const (
	// ProjectTypeGo indicates a standard Go project with go.mod.
	ProjectTypeGo ProjectType = "go"
	// ProjectTypeWails indicates a Wails desktop application.
	ProjectTypeWails ProjectType = "wails"
	// ProjectTypeNode indicates a Node.js project with package.json.
	ProjectTypeNode ProjectType = "node"
	// ProjectTypePHP indicates a PHP/Laravel project with composer.json.
	ProjectTypePHP ProjectType = "php"
	// ProjectTypeDocker indicates a Docker-based project with Dockerfile.
	ProjectTypeDocker ProjectType = "docker"
	// ProjectTypeLinuxKit indicates a LinuxKit VM configuration.
	ProjectTypeLinuxKit ProjectType = "linuxkit"
	// ProjectTypeTaskfile indicates a project using Taskfile automation.
	ProjectTypeTaskfile ProjectType = "taskfile"
)

// Target represents a build target platform.
type Target struct {
	OS   string
	Arch string
}

// String returns the target in GOOS/GOARCH format.
func (t Target) String() string {
	return t.OS + "/" + t.Arch
}

// Artifact represents a build output file.
type Artifact struct {
	Path     string
	OS       string
	Arch     string
	Checksum string
}

// Config holds build configuration.
type Config struct {
	// ProjectDir is the root directory of the project.
	ProjectDir string
	// OutputDir is where build artifacts are placed.
	OutputDir string
	// Name is the output binary name.
	Name string
	// Version is the build version string.
	Version string
	// LDFlags are additional linker flags.
	LDFlags []string

	// Docker-specific config
	Dockerfile string            // Path to Dockerfile (default: Dockerfile)
	Registry   string            // Container registry (default: ghcr.io)
	Image      string            // Image name (owner/repo format)
	Tags       []string          // Additional tags to apply
	BuildArgs  map[string]string // Docker build arguments
	Push       bool              // Whether to push after build

	// LinuxKit-specific config
	LinuxKitConfig string   // Path to LinuxKit YAML config
	Formats        []string // Output formats (iso, qcow2, raw, vmdk)
}

// Builder defines the interface for project-specific build implementations.
type Builder interface {
	// Name returns the builder's identifier.
	Name() string
	// Detect checks if this builder can handle the project in the given directory.
	Detect(dir string) (bool, error)
	// Build compiles the project for the specified targets.
	Build(ctx context.Context, cfg *Config, targets []Target) ([]Artifact, error)
}
