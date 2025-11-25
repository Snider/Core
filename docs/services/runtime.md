---
title: runtime
---
# Service: `runtime`

The `runtime` service provides the main entry point for the application and a helper structure for services to interact with the `Core`.

## Types

### `type Runtime`

`Runtime` is the top-level container that holds the `Core` instance and the Wails application. It serves as the bridge between Wails and the Core framework.

```go
type Runtime struct {
    // Core is the central service manager
    Core *Core
    // app is the Wails application instance
    app  *application.App
}
```

### `type ServiceRuntime[T any]`

`ServiceRuntime` is a generic helper struct designed to be embedded in service implementations. It provides easy access to the `Core` and service-specific options.

```go
type ServiceRuntime[T any] struct {
    core *Core
    opts T
}
```

### `type ServiceFactory`

`ServiceFactory` is a function type that creates a service instance.

```go
type ServiceFactory func() (any, error)
```

## Functions

### `func NewRuntime(app *application.App) (*Runtime, error)`

`NewRuntime` creates and wires together all application services using default settings. It is the standard way to initialize the runtime.

### `func NewWithFactories(app *application.App, factories map[string]ServiceFactory) (*Runtime, error)`

`NewWithFactories` creates a new `Runtime` instance using a provided map of service factories. This allows for flexible, dynamic service registration.

### `func NewServiceRuntime[T any](c *Core, opts T) *ServiceRuntime[T]`

`NewServiceRuntime` creates a new `ServiceRuntime` instance. This is typically used in a service's factory or constructor.

## Methods

### `func (r *Runtime) ServiceName() string`

`ServiceName` returns the name of the service ("Core"). This is used by Wails for service identification.

### `func (r *Runtime) ServiceStartup(ctx context.Context, options application.ServiceOptions)`

`ServiceStartup` delegates the startup lifecycle event to the underlying `Core`, which in turn initializes all registered services.

### `func (r *Runtime) ServiceShutdown(ctx context.Context)`

`ServiceShutdown` delegates the shutdown lifecycle event to the underlying `Core`.

### `func (r *ServiceRuntime[T]) Core() *Core`

`Core` returns the central `Core` instance, giving the service access to other services and features.

### `func (r *ServiceRuntime[T]) Config() Config`

`Config` returns the registered `Config` service from the `Core`. It is a convenience method for accessing configuration.
