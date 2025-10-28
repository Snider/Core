---
title: Core.IO
---

# Core.IO

Short: Local/remote filesystem helpers.

## Overview
Abstracts filesystems (local, SFTP, WebDAV) behind a unified API for reading/writing and listing.

## Setup
```go
import (
  "github.com/Snider/Core"
  "github.com/Snider/Core/io"
  "github.com/Snider/Core/io/sftp"
  "github.com/Snider/Core/io/webdav"
)

// Example of creating a local medium (pre-initialized)
localMedium := io.Local

// Example of creating an SFTP medium
sftpConfig := sftp.ConnectionConfig{
  // ... configure SFTP connection
}
sftpMedium, err := io.NewSFTPMedium(sftpConfig)
if err != nil {
  // handle error
}

// Example of creating a WebDAV medium
webdavConfig := webdav.ConnectionConfig{
  // ... configure WebDAV connection
}
webdavMedium, err := io.NewWebDAVMedium(webdavConfig)
if err != nil {
  // handle error
}

// You can then pass these mediums to services that require an io.Medium
// For example, if a service's New function accepts an io.Medium:
// myService := myservice.New(localMedium)

// If a service registers with core.WithService and needs an io.Medium,
// it would typically receive it during its ServiceStartup or via its New constructor.

// The core.Core itself does not directly register io.Medium implementations
// as services in the same way as other modules, but rather consumes them.
app := core.New(
  // ... other services
  core.WithServiceLock(),
)
```

## Use
- Access the local filesystem: `io.Local`
- Create SFTP/WebDAV mediums: `io.NewSFTPMedium(...)`, `io.NewWebDAVMedium(...)`
- Read/write files using a `Medium`: `medium.Read(path)`, `medium.Write(path, data)`
- List directories (if supported by `Medium` implementation): `medium.List(path)` (Note: `List` is not currently in the `Medium` interface, but `FileGet`, `FileSet`, `EnsureDir`, `IsFile` are)
- Copy files between mediums: `io.Copy(sourceMedium, sourcePath, destMedium, destPath)`

## Notes
- See package `pkg/io/sftp` and `pkg/io/webdav` for specific medium configurations.
