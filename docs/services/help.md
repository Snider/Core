# Help Service

The Help service (`pkg/help`) provides an embeddable documentation system that displays MkDocs-based help content in a dedicated window.

## Features

- Embedded help content (MkDocs static site)
- Context-sensitive help navigation
- Works with or without Display service
- Multiple content sources (embedded, filesystem, custom)

## Basic Usage

```go
import "github.com/Snider/Core/pkg/help"

// Create with default embedded content
helpService, err := help.New(help.Options{})

// Initialize with core dependencies
helpService.Init(coreInstance, displayService)
```

## Showing Help

```go
// Show main help window
err := helpService.Show()

// Show specific section
err := helpService.ShowAt("getting-started")
err := helpService.ShowAt("api/config")
```

## Options

```go
type Options struct {
    Source string  // Path to help content directory
    Assets fs.FS   // Custom filesystem for assets
}
```

### Default Embedded Content

```go
// Uses embedded MkDocs site
helpService, _ := help.New(help.Options{})
```

### Custom Directory

```go
// Use local directory
helpService, _ := help.New(help.Options{
    Source: "/path/to/docs/site",
})
```

### Custom Filesystem

```go
//go:embed docs/*
var docsFS embed.FS

helpService, _ := help.New(help.Options{
    Assets: docsFS,
})
```

## Integration with Core

The help service can work standalone or integrated with Core:

### With Display Service

When Display service is available, help opens through the IPC action system:

```go
// Automatically uses display.open_window action
helpService.Init(core, displayService)
helpService.Show()
```

### Without Display Service

Falls back to direct Wails window creation:

```go
// Creates window directly via Wails
helpService.Init(core, nil)
helpService.Show()
```

## Lifecycle

```go
// Called on application startup
err := helpService.ServiceStartup(ctx)
```

## Building Help Content

Help content is a static MkDocs site. To update:

1. Edit documentation in `docs/` directory
2. Build with MkDocs:
   ```bash
   mkdocs build
   ```
3. The built site goes to `pkg/help/public/`
4. Content is embedded at compile time

## Frontend Usage (TypeScript)

```typescript
import { Show, ShowAt } from '@bindings/help/service';

// Open help window
await Show();

// Open specific section
await ShowAt("configuration");
await ShowAt("api/display");
```

## Help Window Options

The help window opens with default settings:

| Property | Value |
|----------|-------|
| Title | "Help" |
| Width | 800px |
| Height | 600px |

## IPC Action

When using Display service, help triggers this action:

```go
{
    "action": "display.open_window",
    "name":   "help",
    "options": {
        "Title":  "Help",
        "Width":  800,
        "Height": 600,
        "URL":    "/#anchor",  // When using ShowAt
    },
}
```
