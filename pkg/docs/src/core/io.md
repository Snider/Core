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
  core "github.com/Snider/Core"
  ioapi "github.com/Snider/Core/filesystem"
)

app := core.New(
  core.WithService(ioapi.Register),
  core.WithServiceLock(),
)
```

## Use
- Open a filesystem: `fs := ioapi.Local()` or `ioapi.SFTP(cfg)`
- Read/write files: `fs.Read(path)`, `fs.Write(path, data)`
- List directories: `fs.List(path)`

## API
- `Register(c *core.Core) error`
- `Local() FS`
- `SFTP(cfg Config) (FS, error)`
- `WebDAV(cfg Config) (FS, error)`

## Notes
- See package `pkg/v1/core/filesystem/*` for drivers.

