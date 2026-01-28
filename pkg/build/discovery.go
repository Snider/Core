package build

import (
	"os"
	"path/filepath"
	"slices"
)

// Marker files for project type detection.
const (
	markerGoMod       = "go.mod"
	markerWails       = "wails.json"
	markerNodePackage = "package.json"
	markerComposer    = "composer.json"
)

// projectMarker maps a marker file to its project type.
type projectMarker struct {
	file        string
	projectType ProjectType
}

// markers defines the detection order. More specific types come first.
// Wails projects have both wails.json and go.mod, so wails is checked first.
var markers = []projectMarker{
	{markerWails, ProjectTypeWails},
	{markerGoMod, ProjectTypeGo},
	{markerNodePackage, ProjectTypeNode},
	{markerComposer, ProjectTypePHP},
}

// Discover detects project types in the given directory by checking for marker files.
// Returns a slice of detected project types, ordered by priority (most specific first).
// For example, a Wails project returns [wails, go] since it has both wails.json and go.mod.
func Discover(dir string) ([]ProjectType, error) {
	var detected []ProjectType

	for _, m := range markers {
		path := filepath.Join(dir, m.file)
		if fileExists(path) {
			// Avoid duplicates (shouldn't happen with current markers, but defensive)
			if !slices.Contains(detected, m.projectType) {
				detected = append(detected, m.projectType)
			}
		}
	}

	return detected, nil
}

// PrimaryType returns the most specific project type detected in the directory.
// Returns empty string if no project type is detected.
func PrimaryType(dir string) (ProjectType, error) {
	types, err := Discover(dir)
	if err != nil {
		return "", err
	}
	if len(types) == 0 {
		return "", nil
	}
	return types[0], nil
}

// IsGoProject checks if the directory contains a Go project (go.mod or wails.json).
func IsGoProject(dir string) bool {
	return fileExists(filepath.Join(dir, markerGoMod)) ||
		fileExists(filepath.Join(dir, markerWails))
}

// IsWailsProject checks if the directory contains a Wails project.
func IsWailsProject(dir string) bool {
	return fileExists(filepath.Join(dir, markerWails))
}

// IsNodeProject checks if the directory contains a Node.js project.
func IsNodeProject(dir string) bool {
	return fileExists(filepath.Join(dir, markerNodePackage))
}

// IsPHPProject checks if the directory contains a PHP project.
func IsPHPProject(dir string) bool {
	return fileExists(filepath.Join(dir, markerComposer))
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
