package io

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// DeleteAll removes a file or directory and all its contents recursively.
	DeleteAll(path string) error

	// Rename moves a file or directory from oldPath to newPath.
	Rename(oldPath, newPath string) error

	// List returns the directory entries for the given path.
	List(path string) ([]fs.DirEntry, error)

	// Stat returns file information for the given path.
	Stat(path string) (fs.FileInfo, error)

	// Exists checks if a path exists (file or directory).
	Exists(path string) bool

	// IsDir checks if a path exists and is a directory.
	IsDir(path string) bool
}

// FileInfo provides a simple implementation of fs.FileInfo for mock testing.
type FileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (fi FileInfo) Name() string       { return fi.name }
func (fi FileInfo) Size() int64        { return fi.size }
func (fi FileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi FileInfo) ModTime() time.Time { return fi.modTime }
func (fi FileInfo) IsDir() bool        { return fi.isDir }
func (fi FileInfo) Sys() any           { return nil }

// DirEntry provides a simple implementation of fs.DirEntry for mock testing.
type DirEntry struct {
	name  string
	isDir bool
	mode  fs.FileMode
	info  fs.FileInfo
}

func (de DirEntry) Name() string               { return de.name }
func (de DirEntry) IsDir() bool                { return de.isDir }
func (de DirEntry) Type() fs.FileMode          { return de.mode.Type() }
func (de DirEntry) Info() (fs.FileInfo, error) { return de.info, nil }

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
	if _, ok := m.Files[path]; ok {
		delete(m.Files, path)
		return nil
	}
	if _, ok := m.Dirs[path]; ok {
		// Check if directory is empty (no files or subdirs with this prefix)
		prefix := path
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		for f := range m.Files {
			if strings.HasPrefix(f, prefix) {
				return coreerr.E("io.MockMedium.Delete", "directory not empty: "+path, os.ErrExist)
			}
		}
		for d := range m.Dirs {
			if d != path && strings.HasPrefix(d, prefix) {
				return coreerr.E("io.MockMedium.Delete", "directory not empty: "+path, os.ErrExist)
			}
		}
		delete(m.Dirs, path)
		return nil
	}
	return coreerr.E("io.MockMedium.Delete", "path not found: "+path, os.ErrNotExist)
}

// DeleteAll removes a file or directory and all contents from the mock filesystem.
func (m *MockMedium) DeleteAll(path string) error {
	found := false
	if _, ok := m.Files[path]; ok {
		delete(m.Files, path)
		found = true
	}
	if _, ok := m.Dirs[path]; ok {
		delete(m.Dirs, path)
		found = true
	}

	// Delete all entries under this path
	prefix := path
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	for f := range m.Files {
		if strings.HasPrefix(f, prefix) {
			delete(m.Files, f)
			found = true
		}
	}
	for d := range m.Dirs {
		if strings.HasPrefix(d, prefix) {
			delete(m.Dirs, d)
			found = true
		}
	}

	if !found {
		return coreerr.E("io.MockMedium.DeleteAll", "path not found: "+path, os.ErrNotExist)
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
		// Move directory and all contents
		m.Dirs[newPath] = true
		delete(m.Dirs, oldPath)

		oldPrefix := oldPath
		if !strings.HasSuffix(oldPrefix, "/") {
			oldPrefix += "/"
		}
		newPrefix := newPath
		if !strings.HasSuffix(newPrefix, "/") {
			newPrefix += "/"
		}

		// Collect files to move first (don't mutate during iteration)
		filesToMove := make(map[string]string)
		for f, content := range m.Files {
			if strings.HasPrefix(f, oldPrefix) {
				newF := newPrefix + strings.TrimPrefix(f, oldPrefix)
				filesToMove[f] = newF
				_ = content // content will be copied in next loop
			}
		}
		for oldF, newF := range filesToMove {
			m.Files[newF] = m.Files[oldF]
			delete(m.Files, oldF)
		}

		// Collect directories to move first
		dirsToMove := make(map[string]string)
		for d := range m.Dirs {
			if strings.HasPrefix(d, oldPrefix) {
				newD := newPrefix + strings.TrimPrefix(d, oldPrefix)
				dirsToMove[d] = newD
			}
		}
		for oldD, newD := range dirsToMove {
			m.Dirs[newD] = true
			delete(m.Dirs, oldD)
		}
		return nil
	}
	return coreerr.E("io.MockMedium.Rename", "path not found: "+oldPath, os.ErrNotExist)
}

// List returns directory entries for the mock filesystem.
func (m *MockMedium) List(path string) ([]fs.DirEntry, error) {
	if _, ok := m.Dirs[path]; !ok {
		// Check if it's the root or has children
		hasChildren := false
		prefix := path
		if path != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		for f := range m.Files {
			if strings.HasPrefix(f, prefix) {
				hasChildren = true
				break
			}
		}
		if !hasChildren {
			for d := range m.Dirs {
				if strings.HasPrefix(d, prefix) {
					hasChildren = true
					break
				}
			}
		}
		if !hasChildren && path != "" {
			return nil, coreerr.E("io.MockMedium.List", "directory not found: "+path, os.ErrNotExist)
		}
	}

	prefix := path
	if path != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	seen := make(map[string]bool)
	var entries []fs.DirEntry

	// Find immediate children (files)
	for f, content := range m.Files {
		if !strings.HasPrefix(f, prefix) {
			continue
		}
		rest := strings.TrimPrefix(f, prefix)
		if rest == "" || strings.Contains(rest, "/") {
			// Skip if it's not an immediate child
			if idx := strings.Index(rest, "/"); idx != -1 {
				// This is a subdirectory
				dirName := rest[:idx]
				if !seen[dirName] {
					seen[dirName] = true
					entries = append(entries, DirEntry{
						name:  dirName,
						isDir: true,
						mode:  fs.ModeDir | 0755,
						info: FileInfo{
							name:  dirName,
							isDir: true,
							mode:  fs.ModeDir | 0755,
						},
					})
				}
			}
			continue
		}
		if !seen[rest] {
			seen[rest] = true
			entries = append(entries, DirEntry{
				name:  rest,
				isDir: false,
				mode:  0644,
				info: FileInfo{
					name: rest,
					size: int64(len(content)),
					mode: 0644,
				},
			})
		}
	}

	// Find immediate subdirectories
	for d := range m.Dirs {
		if !strings.HasPrefix(d, prefix) {
			continue
		}
		rest := strings.TrimPrefix(d, prefix)
		if rest == "" {
			continue
		}
		// Get only immediate child
		if idx := strings.Index(rest, "/"); idx != -1 {
			rest = rest[:idx]
		}
		if !seen[rest] {
			seen[rest] = true
			entries = append(entries, DirEntry{
				name:  rest,
				isDir: true,
				mode:  fs.ModeDir | 0755,
				info: FileInfo{
					name:  rest,
					isDir: true,
					mode:  fs.ModeDir | 0755,
				},
			})
		}
	}

	return entries, nil
}

// Stat returns file information for the mock filesystem.
func (m *MockMedium) Stat(path string) (fs.FileInfo, error) {
	if content, ok := m.Files[path]; ok {
		return FileInfo{
			name: filepath.Base(path),
			size: int64(len(content)),
			mode: 0644,
		}, nil
	}
	if _, ok := m.Dirs[path]; ok {
		return FileInfo{
			name:  filepath.Base(path),
			isDir: true,
			mode:  fs.ModeDir | 0755,
		}, nil
	}
	return nil, coreerr.E("io.MockMedium.Stat", "path not found: "+path, os.ErrNotExist)
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
