package io

import (
	"io/fs"
	"os"
	"strings"

	coreerr "github.com/host-uk/core/pkg/framework/core"
	"github.com/host-uk/core/pkg/io/local"
)

// Medium defines the standard interface for a storage backend.
// This allows for different implementations (e.g., local disk, S3, SFTP)
// to be used interchangeably.
type Medium interface {
	// Read retrieves the content of a file as a string.
	Read(path string) (string, error)

	// Write saves the given content to a file, overwriting it if it exists.
	Write(path, content string) error

	// EnsureDir makes sure a directory exists, creating it if necessary.
	EnsureDir(path string) error

	// IsFile checks if a path exists and is a regular file.
	IsFile(path string) bool

	// FileGet is a convenience function that reads a file from the medium.
	FileGet(path string) (string, error)

	// FileSet is a convenience function that writes a file to the medium.
	FileSet(path, content string) error

	// Delete removes a file or empty directory.
	Delete(path string) error

	// DeleteAll removes a file or directory recursively.
	DeleteAll(path string) error

	// Rename moves or renames a file or directory.
	Rename(oldPath, newPath string) error

	// List returns directory entries.
	List(path string) ([]fs.DirEntry, error)

	// Stat returns file information.
	Stat(path string) (fs.FileInfo, error)

	// Exists returns true if path exists.
	Exists(path string) bool

	// IsDir returns true if path is a directory.
	IsDir(path string) bool
}

// Local is a pre-initialized medium for the local filesystem.
// It uses "/" as root, providing unsandboxed access to the filesystem.
// For sandboxed access, use NewSandboxed with a specific root path.
var Local Medium

func init() {
	var err error
	Local, err = local.New("/")
	if err != nil {
		panic("io: failed to initialize Local medium: " + err.Error())
	}
}

// NewSandboxed creates a new Medium sandboxed to the given root directory.
// All file operations are restricted to paths within the root.
// The root directory will be created if it doesn't exist.
func NewSandboxed(root string) (Medium, error) {
	return local.New(root)
}

// --- Helper Functions ---

// Read retrieves the content of a file from the given medium.
func Read(m Medium, path string) (string, error) {
	return m.Read(path)
}

// Write saves the given content to a file in the given medium.
func Write(m Medium, path, content string) error {
	return m.Write(path, content)
}

// EnsureDir makes sure a directory exists in the given medium.
func EnsureDir(m Medium, path string) error {
	return m.EnsureDir(path)
}

// IsFile checks if a path exists and is a regular file in the given medium.
func IsFile(m Medium, path string) bool {
	return m.IsFile(path)
}

// Copy copies a file from one medium to another.
func Copy(src Medium, srcPath string, dst Medium, dstPath string) error {
	content, err := src.Read(srcPath)
	if err != nil {
		return coreerr.E("io.Copy", "read failed: "+srcPath, err)
	}
	if err := dst.Write(dstPath, content); err != nil {
		return coreerr.E("io.Copy", "write failed: "+dstPath, err)
	}
	return nil
}

// --- MockMedium ---

// MockMedium is an in-memory implementation of Medium for testing.
type MockMedium struct {
	Files map[string]string
	Dirs  map[string]bool
}

// NewMockMedium creates a new MockMedium instance.
func NewMockMedium() *MockMedium {
	return &MockMedium{
		Files: make(map[string]string),
		Dirs:  make(map[string]bool),
	}
}

// Read retrieves the content of a file from the mock filesystem.
func (m *MockMedium) Read(path string) (string, error) {
	content, ok := m.Files[path]
	if !ok {
		return "", coreerr.E("io.MockMedium.Read", "file not found: "+path, os.ErrNotExist)
	}
	return content, nil
}

// Write saves the given content to a file in the mock filesystem.
func (m *MockMedium) Write(path, content string) error {
	m.Files[path] = content
	return nil
}

// EnsureDir records that a directory exists in the mock filesystem.
func (m *MockMedium) EnsureDir(path string) error {
	m.Dirs[path] = true
	return nil
}

// IsFile checks if a path exists as a file in the mock filesystem.
func (m *MockMedium) IsFile(path string) bool {
	_, ok := m.Files[path]
	return ok
}

// FileGet is a convenience function that reads a file from the mock filesystem.
func (m *MockMedium) FileGet(path string) (string, error) {
	return m.Read(path)
}

// FileSet is a convenience function that writes a file to the mock filesystem.
func (m *MockMedium) FileSet(path, content string) error {
	return m.Write(path, content)
}

// Delete removes a file or empty directory from the mock filesystem.
func (m *MockMedium) Delete(path string) error {
	delete(m.Files, path)
	delete(m.Dirs, path)
	return nil
}

// DeleteAll removes a file or directory recursively from the mock filesystem.
func (m *MockMedium) DeleteAll(path string) error {
	delete(m.Files, path)
	delete(m.Dirs, path)

	prefix := path + "/"
	for k := range m.Files {
		if strings.HasPrefix(k, prefix) {
			delete(m.Files, k)
		}
	}
	for k := range m.Dirs {
		if strings.HasPrefix(k, prefix) {
			delete(m.Dirs, k)
		}
	}
	return nil
}

// Rename moves or renames a file in the mock filesystem.
func (m *MockMedium) Rename(oldPath, newPath string) error {
	if content, ok := m.Files[oldPath]; ok {
		m.Files[newPath] = content
		delete(m.Files, oldPath)
	}
	if m.Dirs[oldPath] {
		m.Dirs[newPath] = true
		delete(m.Dirs, oldPath)
	}
	return nil
}

// List returns directory entries from the mock filesystem.
func (m *MockMedium) List(path string) ([]fs.DirEntry, error) {
	return []fs.DirEntry{}, nil
}

// Stat returns file information from the mock filesystem.
func (m *MockMedium) Stat(path string) (fs.FileInfo, error) {
	if _, ok := m.Files[path]; ok {
		return nil, nil // Mock returns nil info for simplicity
	}
	if _, ok := m.Dirs[path]; ok {
		return nil, nil
	}
	return nil, os.ErrNotExist
}

// Exists returns true if path exists in the mock filesystem.
func (m *MockMedium) Exists(path string) bool {
	if _, ok := m.Files[path]; ok {
		return true
	}
	_, ok := m.Dirs[path]
	return ok
}

// IsDir returns true if path is a directory in the mock filesystem.
func (m *MockMedium) IsDir(path string) bool {
	_, ok := m.Dirs[path]
	return ok
}
