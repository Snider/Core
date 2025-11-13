# Core Library Overview

Core is an opinionated framework for building robust, production-grade Go desktop applications using the [Wails](https://wails.io/) framework. It provides a modular, service-based architecture that simplifies development and ensures maintainability.

## Key Features

- **Modular Architecture**: Core is divided into a set of independent services, each responsible for a specific domain (e.g., `config`, `crypt`, `display`).
- **Unified Runtime**: A central `Runtime` object initializes and manages the lifecycle of all services, providing a simple and consistent entry point for your application.
- **Dependency Injection**: Services are designed to be testable and decoupled, with dependencies injected at runtime.
- **Standardized Error Handling**: A custom error package (`pkg/e`) provides a consistent way to wrap and handle errors throughout the application.
- **Automated Documentation**: This documentation site is automatically generated from the Go source code, ensuring it stays in sync with the public API.

## Getting Started

To start using the Core library, initialize the runtime in your `main.go` file:

```go
package main

import (
    "embed"
    "log"

    "github.com/Snider/Core/runtime"
    "github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:public
var assets embed.FS

func main() {
    app := application.New(application.Options{
        Assets: application.AssetOptions{
            Handler: application.AssetFileServerFS(assets),
        },
    })

    rt, err := runtime.New(app)
    if err != nil {
        log.Fatal(err)
    }

    app.Services.Add(application.NewService(rt))

    err = app.Run()
    if err != nil {
        log.Fatal(err)
    }
}
```

For more detailed information on each service, see the **Services** section in the navigation.
