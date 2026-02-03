// Package local provides a local filesystem implementation of the io.Medium interface.
package local

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Medium is a local filesystem storage backend.
type Medium struct {
	root string
}

// New creates a new local Medium rooted at the given directory.
// Pass "/" for full filesystem access, or a specific path to sandbox.
func New(root string) (*Medium, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Medium{root: abs}, nil
}

// path sanitizes and returns the full path.
// Replaces .. with . to prevent traversal, then joins with root.
// Absolute paths are sandboxed under root (unless root is "/").
func (m *Medium) path(p string) string {
	if p == "" {
		return m.root
	}
	clean := strings.ReplaceAll(p, "..", ".")
	if filepath.IsAbs(clean) {
		// If root is "/", allow absolute paths through
		if m.root == "/" {
			return filepath.Clean(clean)
		}
		// Otherwise, sandbox absolute paths by stripping volume + leading separators
		vol := filepath.VolumeName(clean)
		clean = strings.TrimPrefix(clean, vol)
		clean = strings.TrimLeft(clean, string(os.PathSeparator)+"/")
		return filepath.Join(m.root, clean)
	}
	return filepath.Join(m.root, clean)
}

// Read returns file contents as string.
func (m *Medium) Read(p string) (string, error) {
	data, err := os.ReadFile(m.path(p))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Write saves content to file, creating parent directories as needed.
func (m *Medium) Write(p, content string) error {
	full := m.path(p)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0644)
}

// EnsureDir creates directory if it doesn't exist.
func (m *Medium) EnsureDir(p string) error {
	return os.MkdirAll(m.path(p), 0755)
}

// IsDir returns true if path is a directory.
func (m *Medium) IsDir(p string) bool {
	if p == "" {
		return false
	}
	info, err := os.Stat(m.path(p))
	return err == nil && info.IsDir()
}

// IsFile returns true if path is a regular file.
func (m *Medium) IsFile(p string) bool {
	if p == "" {
		return false
	}
	info, err := os.Stat(m.path(p))
	return err == nil && info.Mode().IsRegular()
}

// Exists returns true if path exists.
func (m *Medium) Exists(p string) bool {
	_, err := os.Stat(m.path(p))
	return err == nil
}

// List returns directory entries.
func (m *Medium) List(p string) ([]fs.DirEntry, error) {
	return os.ReadDir(m.path(p))
}

// Stat returns file info.
func (m *Medium) Stat(p string) (fs.FileInfo, error) {
	return os.Stat(m.path(p))
}

// Delete removes a file or empty directory.
func (m *Medium) Delete(p string) error {
	full := m.path(p)
	if len(full) < 3 {
		return nil
	}
	return os.Remove(full)
}

// DeleteAll removes a file or directory recursively.
func (m *Medium) DeleteAll(p string) error {
	full := m.path(p)
	if len(full) < 3 {
		return nil
	}
	return os.RemoveAll(full)
}

// Rename moves a file or directory.
func (m *Medium) Rename(oldPath, newPath string) error {
	return os.Rename(m.path(oldPath), m.path(newPath))
}

// FileGet is an alias for Read.
func (m *Medium) FileGet(p string) (string, error) {
	return m.Read(p)
}

// FileSet is an alias for Write.
func (m *Medium) FileSet(p, content string) error {
	return m.Write(p, content)
}
