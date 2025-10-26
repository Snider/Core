---
title: Core.Workspace
---

# Core.Workspace

Short: Projects and paths.

## Overview
Provides a consistent way to resolve app/project directories, temp/cache locations, and user data paths across platforms.

## Setup
```go
import (
  core "github.com/Snider/Core"
  workspace "github.com/Snider/Core/workspace"
)

app := core.New(
  core.WithService(workspace.Register),
  core.WithServiceLock(),
)
```

## Use
- Get app data dir: `ws.DataDir()`
- Get cache dir: `ws.CacheDir()`
- Resolve project path: `ws.Project("my-app")`

## API
- `Register(c *core.Core) error`
- `DataDir() string`
- `CacheDir() string`
- `Project(name string) string`

## Notes
- Follows OS directory standards (AppData, ~/Library, XDG, etc.).


