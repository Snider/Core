// Package build provides project type detection and cross-compilation for the Core build system.
// It supports Go, Wails, Node.js, and PHP projects with automatic detection based on
// marker files (go.mod, wails.json, package.json, composer.json).
package build

import (
	"context"
)

// ProjectType represents a detected project type.
type ProjectType string

const (
	ProjectTypeGo     ProjectType = "go"
	ProjectTypeWails  ProjectType = "wails"
	ProjectTypeNode   ProjectType = "node"
	ProjectTypePHP    ProjectType = "php"
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
