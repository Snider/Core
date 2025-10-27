---
title: Core
---

# Core

Short: Framework bootstrap and service container.

## What it is
Core wires modules together, provides lifecycle hooks, and locks the service graph for clarity and safety.

## Setup
```go
import (
  "github.com/Snider/Core/"
  "github.com/wailsapp/wails/v3/pkg/application"
  // Import other service packages you want to register, e.g.:
  // "github.com/Snider/Core/config"
  // "github.com/Snider/Core/display"
)

// Create a new Wails application instance (if not already available)
wailsApp := application.New(application.Options{})

// Initialize Core with services and options
coreApp, err := core.New(
  core.WithWails(wailsApp), // Integrate Wails app
  // Register services using their Register function or a direct instance:
  // core.WithService(config.Register), // Dynamic registration
  // core.WithService(func(c *core.Core) (any, error) { return myServiceInstance, nil }), // Static registration
  core.WithServiceLock(), // Lock services after initialization
)
if err != nil {
  // handle error
}

// You can then run your Wails application
// if err := wailsApp.Run(context.Background()); err != nil {
//   // handle error
// }
```

## Use
- Initialize the Core framework with `core.New()`.
- Register services using `core.WithService()`.
- Access registered services using `core.Service()` or `core.ServiceFor[T]()`.
- Send actions to services using `(c *Core) ACTION()`.

## API

### Functions
- `New(opts ...Option) (*Core, error)`: Initializes a new Core instance.
- `WithService(factory func(*Core) (any, error)) Option`: Registers a service with Core.
- `WithWails(app *application.App) Option`: Integrates a Wails application instance with Core.
- `WithAssets(fs embed.FS) Option`: Sets embedded file system assets for Core.
- `WithServiceLock() Option`: Locks the service graph after initialization.
- `App() *application.App`: Returns the global Wails application instance.
- `ServiceFor[T any](c *Core, name string) T`: Retrieves a registered service by name and asserts its type.

### Methods on `*Core`
- `(c *Core) ServiceStartup(context.Context, application.ServiceOptions) error`: Handles service startup logic.
- `(c *Core) ACTION(msg Message) error`: Dispatches an action message to registered handlers.
- `(c *Core) RegisterAction(handler func(*Core, Message) error)`: Registers an action handler.
- `(c *Core) RegisterActions(handlers ...func(*Core, Message) error)`: Registers multiple action handlers.
- `(c *Core) RegisterService(name string, api any) error`: Registers a service with a given name.
- `(c *Core) Service(name string) any`: Retrieves a registered service by name.
- `(c *Core) Config() Config`: Returns the registered Config service.
- `(c *Core) Core() *Core`: Returns the Core instance itself.
