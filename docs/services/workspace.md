---
title: workspace
---
# Service: `workspace`

The `workspace` service manages user workspaces, which are isolated environments for user data and configuration.

## Types

### `type Workspace`

`Workspace` represents a single user workspace.

```go
type Workspace struct {
    Name string
    Path string
}
```

## Methods

### `func CreateWorkspace(identifier, password string) (string, error)`

`CreateWorkspace` creates a new, secure workspace.
- **identifier**: A unique name or ID for the workspace.
- **password**: A password used to secure the workspace (if encryption is supported).

Returns the workspace ID or path.

### `func SwitchWorkspace(name string) error`

`SwitchWorkspace` changes the currently active workspace to the one specified by `name`.

### `func WorkspaceFileGet(filename string) (string, error)`

`WorkspaceFileGet` retrieves the content of a file located within the active workspace.

### `func WorkspaceFileSet(filename, content string) error`

`WorkspaceFileSet` writes content to a file within the active workspace.

### `func HandleIPCEvents(c *core.Core, msg core.Message) error`

`HandleIPCEvents` processes workspace-related IPC messages.

### `func ServiceStartup(ctx context.Context, options application.ServiceOptions) error`

`ServiceStartup` initializes the workspace service and loads the list of available workspaces.
