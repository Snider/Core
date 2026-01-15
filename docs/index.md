# Core Framework

Core is a Web3 Framework for building production-grade Go desktop applications using [Wails v3](https://wails.io/). It replaces Electron with a native Go backend while providing a modern, service-based architecture.

## Why Core?

- **Native Performance**: Go backend with native webview, no Chromium bloat
- **Service Architecture**: Modular, testable services with dependency injection
- **MCP Integration**: Built-in Model Context Protocol support for AI tooling
- **Cross-Platform**: macOS, Windows, and Linux from a single codebase
- **TypeScript Bindings**: Auto-generated bindings for frontend integration

## Quick Example

```go
package main

import (
    "embed"
    "log"

    "github.com/Snider/Core/pkg/core"
    "github.com/Snider/Core/pkg/display"
    "github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    // Create the Core with services
    c, err := core.New(
        core.WithAssets(assets),
        core.WithService(display.NewService),
        core.WithServiceLock(),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create Wails app
    app := application.New(application.Options{
        Name: "MyApp",
    })

    // Run
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Core Services

| Service | Description |
|---------|-------------|
| **Core** | Central service container and lifecycle management |
| **Display** | Window management, dialogs, tray, clipboard |
| **WebView** | JavaScript execution, DOM interaction, screenshots |
| **MCP** | Model Context Protocol server for AI tool integration |
| **Config** | Application configuration and state persistence |
| **Crypt** | Encryption, signing, key management |
| **I18n** | Internationalization and localization |
| **IO** | File system operations |
| **Workspace** | Project and path management |

## Getting Started

1. [Installation](getting-started/installation.md) - Install Go, Wails, and Core
2. [Quick Start](getting-started/quickstart.md) - Build your first app
3. [Architecture](getting-started/architecture.md) - Understand the design

## Links

- **Repository**: [github.com/Snider/Core](https://github.com/Snider/Core)
- **Issues**: [GitHub Issues](https://github.com/Snider/Core/issues)
