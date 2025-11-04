---
title: io
---
# Service: `io`

The `io` service provides a standardized interface for interacting with different storage backends, such as the local disk, S3, or SFTP.

## Types

### `type Medium`

`Medium` defines the standard interface for a storage backend.

```go
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
```

### `type MockMedium`

`MockMedium` implements the `Medium` interface for testing purposes.

```go
type MockMedium struct {
	Files map[string]string
	Dirs  map[string]bool
}
```

#### Methods

- `EnsureDir(path string) error`: Mocks creating a directory.
- `FileGet(path string) (string, error)`: Mocks reading a file.
- `FileSet(path, content string) error`: Mocks writing a file.
- `IsFile(path string) bool`: Mocks checking if a path is a file.
- `Read(path string) (string, error)`: Mocks reading a file.
- `Write(path, content string) error`: Mocks writing a file.

## Functions

- `Copy(sourceMedium Medium, sourcePath string, destMedium Medium, destPath string) error`: Copies a file from a source medium to a destination medium.
- `EnsureDir(m Medium, path string) error`: Ensures a directory exists on the given medium.
- `IsFile(m Medium, path string) bool`: Checks if a path is a file on the given medium.
- `Read(m Medium, path string) (string, error)`: Retrieves the content of a file from the given medium.
- `Write(m Medium, path, content string) error`: Saves content to a file on the given medium.

Would you like to see some examples of how to use this service?
