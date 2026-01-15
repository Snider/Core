# Workspace Service

The Workspace service (`pkg/workspace`) manages isolated user workspaces with encrypted storage and PGP key pairs.

## Features

- Isolated workspace environments
- PGP key pair generation per workspace
- Encrypted workspace identification
- File operations within workspace context
- Multiple workspace support

## Basic Usage

```go
import "github.com/Snider/Core/pkg/workspace"

// With IO medium (standalone)
medium, _ := local.New("/app/workspaces")
ws, err := workspace.New(medium)

// With Core framework (recommended)
c, _ := core.New(
    core.WithService(workspace.Register),
)
ws := core.MustServiceFor[*workspace.Service](c, "workspace")
```

## Creating Workspaces

```go
// Create a new encrypted workspace
workspaceID, err := ws.CreateWorkspace("my-project", "secure-password")
// Returns obfuscated workspace ID

// Workspace structure created:
// workspaces/
//   <workspace-id>/
//     config/
//     log/
//     data/
//     files/
//     keys/
//       key.pub   (PGP public key)
//       key.priv  (PGP private key)
```

## Switching Workspaces

```go
// Switch to a workspace
err := ws.SwitchWorkspace(workspaceID)

// Switch to default workspace
err := ws.SwitchWorkspace("default")
```

## Workspace File Operations

```go
// Write file to active workspace
err := ws.WorkspaceFileSet("config/settings.json", jsonData)

// Read file from active workspace
content, err := ws.WorkspaceFileGet("config/settings.json")
```

## Listing Workspaces

```go
// Get all workspace IDs
workspaces := ws.ListWorkspaces()
for _, id := range workspaces {
    fmt.Println(id)
}
```

## Active Workspace

```go
// Get current workspace info
active := ws.ActiveWorkspace()
if active != nil {
    fmt.Println("Name:", active.Name)
    fmt.Println("Path:", active.Path)
}
```

## Workspace Structure

Each workspace contains:

| Directory | Purpose |
|-----------|---------|
| `config/` | Workspace configuration files |
| `log/` | Workspace logs |
| `data/` | Application data |
| `files/` | User files |
| `keys/` | PGP key pair |

## Security Model

Workspaces use a two-level hashing scheme:

1. **Real Name**: Hash of the identifier
2. **Workspace ID**: Hash of `workspace/{real_name}`

This prevents workspace enumeration while allowing consistent access.

## IPC Events

The workspace service responds to IPC messages:

```go
// Switch workspace via IPC
c.ACTION(core.Message{
    Type: "workspace.switch_workspace",
    Data: map[string]any{
        "name": workspaceID,
    },
})
```

## Frontend Usage (TypeScript)

```typescript
import {
    CreateWorkspace,
    SwitchWorkspace,
    WorkspaceFileGet,
    WorkspaceFileSet,
    ListWorkspaces,
    ActiveWorkspace
} from '@bindings/workspace/service';

// Create workspace
const wsId = await CreateWorkspace("my-project", "password");

// Switch workspace
await SwitchWorkspace(wsId);

// Read/write files
const config = await WorkspaceFileGet("config/app.json");
await WorkspaceFileSet("config/app.json", JSON.stringify(newConfig));

// List all workspaces
const workspaces = await ListWorkspaces();

// Get active workspace
const active = await ActiveWorkspace();
console.log(`Current: ${active.Name} at ${active.Path}`);
```
