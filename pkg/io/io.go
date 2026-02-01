package io

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// DeleteAll removes a path and all its contents recursively.
	DeleteAll(path string) error

	// Rename moves or renames a file or directory.
	Rename(oldPath, newPath string) error

	// Exists checks if a path exists (file or directory).
	Exists(path string) bool

	// IsDir checks if a path exists and is a directory.
	IsDir(path string) bool

	// List returns the contents of a directory.
	List(path string) ([]os.DirEntry, error)

	// Stat returns file information for a path.
	Stat(path string) (os.FileInfo, error)
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
		return err
	}
	return dst.Write(dstPath, content)
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
		return "", errors.New("file not found: " + path)
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

// Delete removes a file or directory from the mock filesystem.
func (m *MockMedium) Delete(path string) error {
	if _, ok := m.Files[path]; ok {
		delete(m.Files, path)
		return nil
	}
	if _, ok := m.Dirs[path]; ok {
		delete(m.Dirs, path)
		return nil
	}
	return errors.New("path not found: " + path)
}

// DeleteAll removes a path and all contents with matching prefix.
func (m *MockMedium) DeleteAll(path string) error {
	found := false
	prefix := path + "/"
	for k := range m.Files {
		if k == path || strings.HasPrefix(k, prefix) {
			delete(m.Files, k)
			found = true
		}
	}
	for k := range m.Dirs {
		if k == path || strings.HasPrefix(k, prefix) {
			delete(m.Dirs, k)
			found = true
		}
	}
	if !found {
		return errors.New("path not found: " + path)
	}
	return nil
}

// Rename moves a file or directory in the mock filesystem.
func (m *MockMedium) Rename(oldPath, newPath string) error {
	if content, ok := m.Files[oldPath]; ok {
		m.Files[newPath] = content
		delete(m.Files, oldPath)
		return nil
	}
	if _, ok := m.Dirs[oldPath]; ok {
		m.Dirs[newPath] = true
		delete(m.Dirs, oldPath)
		return nil
	}
	return errors.New("path not found: " + oldPath)
}

// Exists checks if a path exists in the mock filesystem.
func (m *MockMedium) Exists(path string) bool {
	if _, ok := m.Files[path]; ok {
		return true
	}
	if _, ok := m.Dirs[path]; ok {
		return true
	}
	return false
}

// IsDir checks if a path is a directory in the mock filesystem.
func (m *MockMedium) IsDir(path string) bool {
	_, ok := m.Dirs[path]
	return ok
}

// List returns entries in a directory from the mock filesystem.
func (m *MockMedium) List(path string) ([]os.DirEntry, error) {
	if _, ok := m.Dirs[path]; !ok && path != "" && path != "." {
		return nil, errors.New("directory not found: " + path)
	}
	var entries []os.DirEntry
	seen := make(map[string]bool)
	prefix := path
	if prefix != "" && prefix != "." {
		prefix = prefix + "/"
	} else {
		prefix = ""
	}

	for k, v := range m.Files {
		if prefix == "" || strings.HasPrefix(k, prefix) {
			rel := strings.TrimPrefix(k, prefix)
			name := strings.Split(rel, "/")[0]
			if !seen[name] && name != "" {
				seen[name] = true
				// Check if it's directly in this dir or nested
				if !strings.Contains(rel, "/") {
					entries = append(entries, &mockDirEntry{
						name:  name,
						isDir: false,
						size:  int64(len(v)),
					})
				}
			}
		}
	}
	for k := range m.Dirs {
		if prefix == "" || strings.HasPrefix(k, prefix) {
			rel := strings.TrimPrefix(k, prefix)
			name := strings.Split(rel, "/")[0]
			if !seen[name] && name != "" && !strings.Contains(rel, "/") {
				seen[name] = true
				entries = append(entries, &mockDirEntry{
					name:  name,
					isDir: true,
				})
			}
		}
	}
	return entries, nil
}

// Stat returns file info for a path in the mock filesystem.
func (m *MockMedium) Stat(path string) (os.FileInfo, error) {
	if content, ok := m.Files[path]; ok {
		return &mockFileInfo{
			name:  filepath.Base(path),
			size:  int64(len(content)),
			isDir: false,
		}, nil
	}
	if _, ok := m.Dirs[path]; ok {
		return &mockFileInfo{
			name:  filepath.Base(path),
			isDir: true,
		}, nil
	}
	return nil, errors.New("path not found: " + path)
}

// mockDirEntry implements os.DirEntry for MockMedium.
type mockDirEntry struct {
	name  string
	isDir bool
	size  int64
}

func (e *mockDirEntry) Name() string               { return e.name }
func (e *mockDirEntry) IsDir() bool                { return e.isDir }
func (e *mockDirEntry) Type() os.FileMode          { if e.isDir { return os.ModeDir }; return 0 }
func (e *mockDirEntry) Info() (os.FileInfo, error) { return &mockFileInfo{name: e.name, isDir: e.isDir, size: e.size}, nil }

// mockFileInfo implements os.FileInfo for MockMedium.
type mockFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (i *mockFileInfo) Name() string       { return i.name }
func (i *mockFileInfo) Size() int64        { return i.size }
func (i *mockFileInfo) Mode() os.FileMode  { if i.isDir { return os.ModeDir | 0755 }; return 0644 }
func (i *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (i *mockFileInfo) IsDir() bool        { return i.isDir }
func (i *mockFileInfo) Sys() any           { return nil }
