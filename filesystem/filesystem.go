package filesystem

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

// Pre-initialized, sandboxed medium for the local filesystem.
var Local Medium
