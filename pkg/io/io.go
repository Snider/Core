package io

import (
	"errors"

	"github.com/Snider/Core/pkg/io/local"
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
}

// Local is a pre-initialized medium for the local filesystem.
// It uses "/" as root, providing unsandboxed access to the filesystem.
// For sandboxed access, create a new local.Medium with a specific root path.
var Local Medium

func init() {
	var err error
	Local, err = local.New("/")
	if err != nil {
		panic("io: failed to initialize Local medium: " + err.Error())
	}
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
