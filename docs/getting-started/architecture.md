# Architecture

Core follows a modular, service-based architecture designed for maintainability and testability.

## Overview

```
┌─────────────────────────────────────────────────────────┐
│                     Wails Application                    │
├─────────────────────────────────────────────────────────┤
│                        Core                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │ Display  │ │ WebView  │ │   MCP    │ │  Config  │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │  Crypt   │ │   I18n   │ │    IO    │ │Workspace │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
├─────────────────────────────────────────────────────────┤
│                    Plugin System                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐                │
│  │ Plugin A │ │ Plugin B │ │ Plugin C │                │
│  └──────────┘ └──────────┘ └──────────┘                │
└─────────────────────────────────────────────────────────┘
```

## Core Container

The `Core` struct is the central service container:

```go
type Core struct {
    services       map[string]any      // Service registry
    actions        []ActionHandler     // IPC handlers
    Features       *Features           // Feature flags
    servicesLocked bool                // Prevent late registration
}
```

### Service Registration

Services are registered using factory functions:

```go
core.New(
    core.WithService(display.NewService),  // Auto-discovered name
    core.WithName("custom", myFactory),    // Explicit name
)
```

### Service Retrieval

Type-safe service retrieval:

```go
// Returns error if not found
svc, err := core.ServiceFor[*display.Service](c, "display")

// Panics if not found (use in init code)
svc := core.MustServiceFor[*display.Service](c, "display")
```

## Service Lifecycle

Services can implement lifecycle interfaces:

```go
// Called when app starts
type Startable interface {
    OnStartup(ctx context.Context) error
}

// Called when app shuts down
type Stoppable interface {
    OnShutdown(ctx context.Context) error
}
```

## IPC / Actions

Services communicate via the action system:

```go
// Register a handler
c.RegisterAction(func(c *core.Core, msg core.Message) error {
    if msg.Type == "my-action" {
        // Handle message
    }
    return nil
})

// Send a message
c.ACTION(core.Message{
    Type: "my-action",
    Data: map[string]any{"key": "value"},
})
```

## Frontend Bindings

Wails generates TypeScript bindings automatically:

```typescript
// Auto-generated from Go service
import { ShowNotification } from '@bindings/display/service';

await ShowNotification({
    title: "Hello",
    message: "From TypeScript!"
});
```

## Package Structure

```
pkg/
├── core/           # Core container and interfaces
├── display/        # Window, tray, dialogs, clipboard
├── webview/        # JS execution, DOM, screenshots
├── mcp/            # Model Context Protocol server
├── config/         # Configuration persistence
├── crypt/          # Encryption and signing
├── i18n/           # Internationalization
├── io/             # File system helpers
├── workspace/      # Project management
├── plugin/         # Plugin system
└── module/         # Module system
```

## Design Principles

1. **Dependency Injection**: Services receive dependencies via constructor
2. **Interface Segregation**: Small, focused interfaces
3. **Testability**: All services are mockable
4. **No Globals**: State contained in Core instance
