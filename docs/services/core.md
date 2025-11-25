---
title: core
---
# Service: `core`

The `core` service is the foundation of the application. It serves as a central hub for managing service lifecycles, dependency injection, asset handling, and inter-process communication (IPC).

## Types

### `type Core`

`Core` is the central application object. It holds references to the Wails application instance, configuration, and registered services.

```go
type Core struct {
    // App is the Wails application instance
    App *application.App
    // Features manages feature flags
    Features *Features
    // ... internal fields
}
```

### `type Option`

`Option` is a function type used to configure the `Core` during initialization.

```go
type Option func(*Core) error
```

### `type Startable`

`Startable` is an interface for services that require initialization logic when the application starts.

```go
type Startable interface {
    OnStartup(ctx context.Context) error
}
```

### `type Stoppable`

`Stoppable` is an interface for services that require cleanup logic when the application shuts down.

```go
type Stoppable interface {
    OnShutdown(ctx context.Context) error
}
```

### `type Message`

`Message` is a marker interface for structs that can be sent via the Core's IPC system.

```go
type Message interface{}
```

### `type ActionServiceStartup`

`ActionServiceStartup` is a message dispatched when the application services are starting up.

```go
type ActionServiceStartup struct{}
```

### `type ActionServiceShutdown`

`ActionServiceShutdown` is a message dispatched when the application services are shutting down.

```go
type ActionServiceShutdown struct{}
```

## Functions

### `func New(opts ...Option) (*Core, error)`

`New` initializes a new `Core` instance with the provided options. It sets up the internal service registry and feature flags.

### `func WithService(factory func(*Core) (any, error)) Option`

`WithService` registers a service using a factory function. It automatically discovers the service name from the package path and registers any IPC handlers if the service implements `HandleIPCEvents`.

### `func WithName(name string, factory func(*Core) (any, error)) Option`

`WithName` registers a service with a specific name using a factory function. Unlike `WithService`, it does not automatically discover IPC handlers.

### `func WithWails(app *application.App) Option`

`WithWails` injects the Wails application instance into the `Core`. This is required for services that need to interact with the Wails runtime.

### `func WithAssets(fs embed.FS) Option`

`WithAssets` registers the application's embedded assets (filesystem).

### `func WithServiceLock() Option`

`WithServiceLock` locks the service registry after initialization, preventing further service registration.

### `func ServiceFor[T any](c *Core, name string) (T, error)`

`ServiceFor` retrieves a registered service by its name and asserts that it is of type `T`. Returns an error if the service is not found or the type assertion fails.

### `func MustServiceFor[T any](c *Core, name string) T`

`MustServiceFor` is a helper that wraps `ServiceFor` and panics if the service is not found or the type is incorrect.

### `func App() *application.App`

`App` returns the global Wails application instance. It panics if the Core has not been initialized.

## Methods

### `func (c *Core) Config() Config`

`Config` returns the registered `Config` service. This is a convenience method for accessing application configuration.

### `func (c *Core) Assets() embed.FS`

`Assets` returns the embedded filesystem containing the application's assets.

### `func (c *Core) ServiceStartup(ctx context.Context, options application.ServiceOptions) error`

`ServiceStartup` is the entry point for the Core's startup lifecycle. It initializes all registered `Startable` services and dispatches the `ActionServiceStartup` message.

### `func (c *Core) ServiceShutdown(ctx context.Context) error`

`ServiceShutdown` is the entry point for the Core's shutdown lifecycle. It dispatches the `ActionServiceShutdown` message and calls `OnShutdown` on all `Stoppable` services in reverse registration order.

### `func (c *Core) RegisterAction(handler func(*Core, Message) error)`

`RegisterAction` registers a new IPC message handler.

### `func (c *Core) RegisterActions(handlers ...func(*Core, Message) error)`

`RegisterActions` registers multiple IPC message handlers at once.

### `func (c *Core) ACTION(msg Message) error`

`ACTION` dispatches a message to all registered IPC handlers. It aggregates errors from all handlers.

### `func (c *Core) RegisterService(name string, api any) error`

`RegisterService` adds a new service to the Core. It detects if the service implements `Startable` or `Stoppable` and registers it for lifecycle events.

### `func (c *Core) Service(name string) any`

`Service` retrieves a registered service instance by name as an `any` type. Returns `nil` if not found.
