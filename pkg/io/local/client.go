// Package local provides a local filesystem implementation of the io.Medium interface.
package local

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Medium is a local filesystem storage backend.
type Medium struct {
	root string
}

// New creates a new local Medium with the specified root directory.
// The root directory will be created if it doesn't exist.
func New(root string) (*Medium, error) {
	// Ensure root is an absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	// Create root directory if it doesn't exist
	if err := os.MkdirAll(absRoot, 0755); err != nil {
		return nil, err
	}

	return &Medium{root: absRoot}, nil
}

// path sanitizes and joins the relative path with the root directory.
// Returns an error if a path traversal attempt is detected.
func (m *Medium) path(relativePath string) (string, error) {
	// Clean the path to remove any .. or . components
	cleanPath := filepath.Clean(relativePath)

	// Check for path traversal attempts
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, string(filepath.Separator)+"..") {
		return "", errors.New("path traversal attempt detected")
	}

	fullPath := filepath.Join(m.root, cleanPath)

	// Verify the resulting path is still within root
	if !strings.HasPrefix(fullPath, m.root) {
		return "", errors.New("path traversal attempt detected")
	}

	return fullPath, nil
}

// Read retrieves the content of a file as a string.
func (m *Medium) Read(relativePath string) (string, error) {
	fullPath, err := m.path(relativePath)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// Write saves the given content to a file, overwriting it if it exists.
// Parent directories are created automatically.
func (m *Medium) Write(relativePath, content string) error {
	fullPath, err := m.path(relativePath)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(fullPath, []byte(content), 0644)
}

// EnsureDir makes sure a directory exists, creating it if necessary.
func (m *Medium) EnsureDir(relativePath string) error {
	fullPath, err := m.path(relativePath)
	if err != nil {
		return err
	}

	return os.MkdirAll(fullPath, 0755)
}

// IsFile checks if a path exists and is a regular file.
func (m *Medium) IsFile(relativePath string) bool {
	fullPath, err := m.path(relativePath)
	if err != nil {
		return false
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}

// FileGet is a convenience function that reads a file from the medium.
func (m *Medium) FileGet(relativePath string) (string, error) {
	return m.Read(relativePath)
}

// FileSet is a convenience function that writes a file to the medium.
func (m *Medium) FileSet(relativePath, content string) error {
	return m.Write(relativePath, content)
}
