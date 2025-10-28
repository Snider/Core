---
title: Core.Workspace
---

# Core.Workspace

Short: Manages user workspaces.

## Overview
Provides functionality for creating, switching, and managing isolated user workspaces, including file storage and cryptographic operations within those workspaces.

## Setup
```go
import (
  "github.com/Snider/Core"
  "github.com/Snider/Core/workspace"
  "github.com/Snider/Core/io"
)

workspaceService, err := workspace.New()
if err != nil {
  // handle error
}
app := core.New(
  core.WithService(workspaceService),
  core.WithServiceLock(),
)

// Example of dynamic dependency injection (used with core.WithService)
// If using dynamic injection, the io.Medium dependency would be resolved
// during ServiceStartup, typically by another service providing it.
// app := core.New(
//   core.WithService(workspace.Register),
//   core.WithServiceLock(),
// )
```

## Use
- Create a new workspace: `ws.CreateWorkspace("my-project", "my-password")`
- Switch to an existing workspace: `ws.SwitchWorkspace("my-project-id")`
- Get a file from the active workspace: `ws.WorkspaceFileGet("config.json")`
- Set a file in the active workspace: `ws.WorkspaceFileSet("data.txt", "some content")`

## Notes
- Workspaces are obfuscated and secured with OpenPGP.
- Integrates with `Core.Config` for workspace directory management.
