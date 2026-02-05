// Package local provides a local filesystem implementation of the io.Medium interface.
package local

import (
	goio "io"
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
// Absolute paths are sandboxed under root (unless root is "/").
func (m *Medium) path(p string) string {
	if p == "" {
		return m.root
	}

	// If the path is relative and the medium is rooted at "/",
	// treat it as relative to the current working directory.
	// This makes io.Local behave more like the standard 'os' package.
	if m.root == "/" && !filepath.IsAbs(p) {
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, p)
	}

	// Use filepath.Clean with a leading slash to resolve all .. and . internally
	// before joining with the root. This is a standard way to sandbox paths.
	clean := filepath.Clean("/" + p)

	// If root is "/", allow absolute paths through
	if m.root == "/" {
		return clean
	}

	// Join cleaned relative path with root
	return filepath.Join(m.root, clean)
}

// validatePath ensures the path is within the sandbox, following symlinks if they exist.
func (m *Medium) validatePath(p string) (string, error) {
	if m.root == "/" {
		return m.path(p), nil
	}

	// Split the cleaned path into components
	parts := strings.Split(filepath.Clean("/"+p), string(os.PathSeparator))
	current := m.root

	for _, part := range parts {
		if part == "" {
			continue
		}

		next := filepath.Join(current, part)
		realNext, err := filepath.EvalSymlinks(next)
		if err != nil {
			if os.IsNotExist(err) {
				// Part doesn't exist, we can't follow symlinks anymore.
				// Since the path is already Cleaned and current is safe,
				// appending a component to current will not escape.
				current = next
				continue
			}
			return "", err
		}

		// Verify the resolved part is still within the root
		rel, err := filepath.Rel(m.root, realNext)
		if err != nil || strings.HasPrefix(rel, "..") {
			// Security event: sandbox escape attempt (path escapes root)
			return "", os.ErrPermission // Path escapes sandbox
		}
		current = realNext
	}

	return current, nil
}

// Read returns file contents as string.
func (m *Medium) Read(p string) (string, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Write saves content to file, creating parent directories as needed.
func (m *Medium) Write(p, content string) error {
	full, err := m.validatePath(p)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0644)
}

// EnsureDir creates directory if it doesn't exist.
func (m *Medium) EnsureDir(p string) error {
	full, err := m.validatePath(p)
	if err != nil {
		return err
	}
	return os.MkdirAll(full, 0755)
}

// IsDir returns true if path is a directory.
func (m *Medium) IsDir(p string) bool {
	if p == "" {
		return false
	}
	full, err := m.validatePath(p)
	if err != nil {
		return false
	}
	info, err := os.Stat(full)
	return err == nil && info.IsDir()
}

// IsFile returns true if path is a regular file.
func (m *Medium) IsFile(p string) bool {
	if p == "" {
		return false
	}
	full, err := m.validatePath(p)
	if err != nil {
		return false
	}
	info, err := os.Stat(full)
	return err == nil && info.Mode().IsRegular()
}

// Exists returns true if path exists.
func (m *Medium) Exists(p string) bool {
	full, err := m.validatePath(p)
	if err != nil {
		return false
	}
	_, err = os.Stat(full)
	return err == nil
}

// List returns directory entries.
func (m *Medium) List(p string) ([]fs.DirEntry, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return nil, err
	}
	return os.ReadDir(full)
}

// Stat returns file info.
func (m *Medium) Stat(p string) (fs.FileInfo, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return nil, err
	}
	return os.Stat(full)
}

// Open opens the named file for reading.
func (m *Medium) Open(p string) (fs.File, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return nil, err
	}
	return os.Open(full)
}

// Create creates or truncates the named file.
func (m *Medium) Create(p string) (goio.WriteCloser, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return nil, err
	}
	return os.Create(full)
}

// Append opens the named file for appending, creating it if it doesn't exist.
func (m *Medium) Append(p string) (goio.WriteCloser, error) {
	full, err := m.validatePath(p)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(full, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

// Delete removes a file or empty directory.
func (m *Medium) Delete(p string) error {
	full, err := m.validatePath(p)
	if err != nil {
		return err
	}
	if len(full) < 3 {
		return nil
	}
	return os.Remove(full)
}

// DeleteAll removes a file or directory recursively.
func (m *Medium) DeleteAll(p string) error {
	full, err := m.validatePath(p)
	if err != nil {
		return err
	}
	if len(full) < 3 {
		return nil
	}
	return os.RemoveAll(full)
}

// Rename moves a file or directory.
func (m *Medium) Rename(oldPath, newPath string) error {
	oldFull, err := m.validatePath(oldPath)
	if err != nil {
		return err
	}
	newFull, err := m.validatePath(newPath)
	if err != nil {
		return err
	}
	return os.Rename(oldFull, newFull)
}

// FileGet is an alias for Read.
func (m *Medium) FileGet(p string) (string, error) {
	return m.Read(p)
}

// FileSet is an alias for Write.
func (m *Medium) FileSet(p, content string) error {
	return m.Write(p, content)
}
