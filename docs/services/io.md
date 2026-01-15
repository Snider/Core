# IO Service

The IO package (`pkg/io`) provides a unified interface for file operations across different storage backends (local filesystem, S3, SFTP, etc.).

## Features

- Abstract `Medium` interface for storage backends
- Local filesystem implementation
- Copy between different mediums
- Mock implementation for testing

## Medium Interface

All storage backends implement the `Medium` interface:

```go
type Medium interface {
    Read(path string) (string, error)
    Write(path, content string) error
    EnsureDir(path string) error
    IsFile(path string) bool
    FileGet(path string) (string, error)
    FileSet(path, content string) error
}
```

## Local Filesystem

```go
import (
    "github.com/Snider/Core/pkg/io"
    "github.com/Snider/Core/pkg/io/local"
)

// Pre-initialized global medium (root = "/")
content, err := io.Local.Read("/etc/hosts")

// Create sandboxed medium
medium, err := local.New("/app/data")
content, err := medium.Read("config.json")  // Reads /app/data/config.json
```

## Basic Operations

```go
// Read file
content, err := medium.Read("path/to/file.txt")

// Write file
err := medium.Write("path/to/file.txt", "content")

// Check if file exists
if medium.IsFile("config.json") {
    // File exists
}

// Ensure directory exists
err := medium.EnsureDir("path/to/dir")

// Convenience methods
content, err := medium.FileGet("file.txt")
err := medium.FileSet("file.txt", "content")
```

## Helper Functions

Package-level functions that work with any Medium:

```go
// Read from medium
content, err := io.Read(medium, "file.txt")

// Write to medium
err := io.Write(medium, "file.txt", "content")

// Ensure directory
err := io.EnsureDir(medium, "path/to/dir")

// Check if file
exists := io.IsFile(medium, "file.txt")
```

## Copy Between Mediums

```go
localMedium, _ := local.New("/local/path")
remoteMedium := s3.New(bucket, region)  // hypothetical S3 implementation

// Copy from local to remote
err := io.Copy(localMedium, "data.json", remoteMedium, "backup/data.json")
```

## Mock Medium for Testing

```go
import "github.com/Snider/Core/pkg/io"

func TestMyFunction(t *testing.T) {
    mock := io.NewMockMedium()

    // Pre-populate files
    mock.Files["config.json"] = `{"key": "value"}`
    mock.Dirs["data"] = true

    // Use in tests
    myService := NewService(mock)

    // Verify writes
    err := myService.SaveData("test")
    if mock.Files["data/test.json"] != expectedContent {
        t.Error("unexpected content")
    }
}
```

## Creating Custom Backends

Implement the `Medium` interface for custom storage:

```go
type S3Medium struct {
    bucket string
    client *s3.Client
}

func (m *S3Medium) Read(path string) (string, error) {
    // Implement S3 read
}

func (m *S3Medium) Write(path, content string) error {
    // Implement S3 write
}

// ... implement remaining methods
```

## Error Handling

```go
content, err := medium.Read("missing.txt")
if err != nil {
    // File not found or read error
    log.Printf("Read failed: %v", err)
}
```

## Frontend Usage

The IO package is primarily used server-side. Frontend file operations should use the Display service dialogs or direct API calls:

```typescript
import { OpenFileDialog, SaveFileDialog } from '@bindings/display/service';

// Open file picker
const path = await OpenFileDialog({
    title: "Select File",
    filters: [{ displayName: "Text", pattern: "*.txt" }]
});

// Save file picker
const savePath = await SaveFileDialog({
    title: "Save As",
    defaultFilename: "document.txt"
});
```
