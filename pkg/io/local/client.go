// Package local provides a local filesystem implementation of the io.Medium interface.
package local

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Sentinel errors for path validation.
var (
	// ErrPathTraversal indicates an attempt to access paths outside the root.
	ErrPathTraversal = errors.New("path traversal attempt detected")
	// ErrSymlinkTraversal indicates a symlink points outside the root.
	ErrSymlinkTraversal = errors.New("symlink points outside workspace")
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
// Resolves symlinks to prevent bypass attacks.
func (m *Medium) path(relativePath string) (string, error) {
	// Clean the path to remove any .. or . components
	cleanPath := filepath.Clean(relativePath)

	// Check for obvious path traversal attempts
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, string(filepath.Separator)+"..") {
		return "", ErrPathTraversal
	}

	fullPath := filepath.Join(m.root, cleanPath)

	// Verify the resulting path is still within root (before symlink resolution)
	if !strings.HasPrefix(fullPath, m.root) {
		return "", ErrPathTraversal
	}

	// Resolve symlinks to get the real path
	// First resolve the root to handle any symlinks in the root path
	realRoot, err := filepath.EvalSymlinks(m.root)
	if err != nil {
		return "", err
	}

	// Try to resolve the full path - this may fail if path doesn't exist yet
	realPath, err := resolvePathWithSymlinks(fullPath)
	if err != nil {
		return "", err
	}

	// Verify resolved path is within resolved root
	if !strings.HasPrefix(realPath, realRoot) && realPath != realRoot {
		return "", ErrSymlinkTraversal
	}

	return fullPath, nil
}

// resolvePathWithSymlinks resolves symlinks in a path, even if the path doesn't exist.
// It walks up the directory tree to find the nearest existing ancestor,
// resolves symlinks for that ancestor, then appends the remaining path components.
func resolvePathWithSymlinks(path string) (string, error) {
	// If the path exists, just resolve it directly
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved, nil
	}

	// Path doesn't exist - walk up to find existing ancestor
	current := path
	var remainder []string

	for {
		parent := filepath.Dir(current)
		if parent == current {
			// Reached root, nothing more to resolve
			break
		}

		remainder = append([]string{filepath.Base(current)}, remainder...)
		current = parent

		// Try to resolve this ancestor
		if resolved, err := filepath.EvalSymlinks(current); err == nil {
			// Found existing ancestor, build full path
			return filepath.Join(append([]string{resolved}, remainder...)...), nil
		}
	}

	// No existing ancestor found, return original path
	return path, nil
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
