---
title: workspace
---
# Service: `workspace`

## Constants

```go
defaultWorkspacelistFile
```

## Types

### `type Options`

```go
type Options struct {
    // Options holds configuration for the workspace service.
}
```

### `type Service`

```go
type Service struct {
    core.Runtime[*Options]
    activeWorkspace *Workspace
    workspaceList   map[string]string // Maps Workspace ID to Public Key
    medium          io.Medium
}
```

Service manages user workspaces.

#### Methods

- `CreateWorkspace(identifier, password string) (string, error)`: CreateWorkspace creates a new, obfuscated workspace on the local medium.
- `HandleIPCEvents(c *core.Core, msg core.Message) error`: HandleIPCEvents processes IPC messages, including injecting dependencies on startup.
- `ServiceStartup(context.Context, application.ServiceOptions) error`: ServiceStartup initializes the service, loading the workspace list.
- `SwitchWorkspace(name string) error`: SwitchWorkspace changes the active workspace.
- `WorkspaceFileGet(filename string) (string, error)`: WorkspaceFileGet retrieves a file from the active workspace.
- `WorkspaceFileSet(filename, content string) error`: WorkspaceFileSet writes a file to the active workspace.
- `getWorkspaceDir() (string, error)`: getWorkspaceDir retrieves the WorkspaceDir from the config service.

### `type Workspace`

```go
type Workspace struct {
    Name string
    Path string
}
```

Workspace represents a user's workspace.

### `type localMedium`

```go
type localMedium struct{}
```

localMedium implements the Medium interface for the local disk.

#### Methods

- `EnsureDir(path string) error`: EnsureDir creates a directory on the local disk.
- `FileGet(path string) (string, error)`: FileGet reads a file from the local disk.
- `FileSet(path, content string) error`: FileSet writes a file to the local disk.
- `IsFile(path string) bool`: IsFile checks if a path exists and is a file on the local disk.
- `Read(path string) (string, error)`: Read reads a file from the local disk.
- `Write(path, content string) error`: Write writes a file to the local disk.

## Functions

- `NewLocalMedium() io.Medium`: NewLocalMedium creates a new instance of the local storage medium.
- `Register(c *core.Core) (any, error)`: Register is the constructor for dynamic dependency injection (used with core.WithService). It creates a Service instance and initializes its core.Runtime field. Dependencies are injected during ServiceStartup.
