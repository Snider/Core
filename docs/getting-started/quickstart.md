# Quick Start

Build a simple Core application in 5 minutes.

## Create Project

```bash
mkdir myapp && cd myapp
go mod init myapp
```

## Install Dependencies

```bash
go get github.com/Snider/Core@latest
go get github.com/wailsapp/wails/v3@latest
```

## Create Main File

Create `main.go`:

```go
package main

import (
    "context"
    "embed"
    "log"

    "github.com/Snider/Core/pkg/core"
    "github.com/Snider/Core/pkg/display"
    "github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    // Initialize Core with display service
    c, err := core.New(
        core.WithAssets(assets),
        core.WithService(display.NewService),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Get display service for window creation
    displaySvc := core.MustServiceFor[*display.Service](c, "display")

    // Create Wails application
    app := application.New(application.Options{
        Name: "My App",
        Assets: application.AssetOptions{
            FS: assets,
        },
    })

    // Create main window
    app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
        Title:  "My App",
        Width:  1200,
        Height: 800,
        URL:    "/",
    })

    // Register display service with Wails
    app.RegisterService(displaySvc)

    // Run application
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Create Frontend

Create a minimal frontend:

```bash
mkdir -p frontend/dist
```

Create `frontend/dist/index.html`:

```html
<!DOCTYPE html>
<html>
<head>
    <title>My App</title>
    <style>
        body {
            font-family: system-ui, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: #1a1a2e;
            color: #eee;
        }
    </style>
</head>
<body>
    <h1>Hello from Core!</h1>
</body>
</html>
```

## Run Development Mode

```bash
wails3 dev
```

## Build for Production

```bash
wails3 build
```

## Next Steps

- [Architecture](architecture.md) - Understand how Core works
- [Display Service](../services/display.md) - Window and dialog management
- [MCP Integration](../services/mcp.md) - AI tool support
