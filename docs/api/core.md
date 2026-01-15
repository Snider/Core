# Core API Reference

Complete API reference for the Core framework (`pkg/core`).

## Core Struct

The central application container.

### Creation

```go
func New(opts ...Option) (*Core, error)
```

Creates a new Core instance with the specified options.

### Methods

#### Service Access

```go
func ServiceFor[T any](c *Core, name string) (T, error)
```

Retrieves a service by name with type safety.

```go
func MustServiceFor[T any](c *Core, name string) T
```

Retrieves a service by name, panics if not found or wrong type.

#### Actions

```go
func (c *Core) ACTION(msg Message) error
```

Broadcasts a message to all registered action handlers.

```go
func (c *Core) RegisterAction(handler func(*Core, Message) error)
```

Registers an action handler.

#### Service Registration

```go
func (c *Core) AddService(name string, svc any) error
```

Manually adds a service to the registry.

#### Config Access

```go
func (c *Core) Config() *config.Service
```

Returns the config service if registered.

## Options

### WithService

```go
func WithService(factory ServiceFactory) Option
```

Registers a service using its factory function.

```go
c, _ := core.New(
    core.WithService(config.Register),
    core.WithService(display.NewService),
)
```

### WithName

```go
func WithName(name string, factory ServiceFactory) Option
```

Registers a service with an explicit name.

```go
c, _ := core.New(
    core.WithName("mydb", database.NewService),
)
```

### WithAssets

```go
func WithAssets(assets embed.FS) Option
```

Sets embedded assets for the application.

### WithServiceLock

```go
func WithServiceLock() Option
```

Prevents late service registration after initialization.

## ServiceFactory

```go
type ServiceFactory func(c *Core) (any, error)
```

Factory function signature for service creation.

## Message

```go
type Message interface{}
```

Messages can be any type. Common patterns:

```go
// Map-based message
c.ACTION(map[string]any{
    "action": "user.created",
    "id":     "123",
})

// Typed message
type UserCreated struct {
    ID    string
    Email string
}
c.ACTION(UserCreated{ID: "123", Email: "user@example.com"})
```

## ServiceRuntime

Generic helper for services that need Core access.

```go
type ServiceRuntime[T any] struct {
    core    *Core
    options T
}
```

### Creation

```go
func NewServiceRuntime[T any](c *Core, opts T) *ServiceRuntime[T]
```

### Methods

```go
func (r *ServiceRuntime[T]) Core() *Core
func (r *ServiceRuntime[T]) Options() T
func (r *ServiceRuntime[T]) Config() *config.Service
```

### Usage

```go
type MyOptions struct {
    Timeout time.Duration
}

type MyService struct {
    *core.ServiceRuntime[MyOptions]
}

func NewMyService(c *core.Core) (any, error) {
    opts := MyOptions{Timeout: 30 * time.Second}
    return &MyService{
        ServiceRuntime: core.NewServiceRuntime(c, opts),
    }, nil
}

func (s *MyService) DoSomething() {
    timeout := s.Options().Timeout
    cfg := s.Config()
    // ...
}
```

## Lifecycle Interfaces

### Startable

```go
type Startable interface {
    OnStartup(ctx context.Context) error
}
```

Implement for initialization on app start.

### Stoppable

```go
type Stoppable interface {
    OnShutdown(ctx context.Context) error
}
```

Implement for cleanup on app shutdown.

### IPC Handler

```go
type IPCHandler interface {
    HandleIPCEvents(c *Core, msg Message) error
}
```

Automatically registered when using `WithService`.

## Built-in Actions

### ActionServiceStartup

```go
type ActionServiceStartup struct{}
```

Sent to all services when application starts.

### ActionServiceShutdown

```go
type ActionServiceShutdown struct{}
```

Sent to all services when application shuts down.

## Error Helpers

```go
func E(service, operation string, err error) error
```

Creates a contextual error with service and operation info.

```go
if err != nil {
    return core.E("myservice", "Connect", err)
}
// Error: myservice.Connect: connection refused
```

## Complete Example

```go
package main

import (
    "context"
    "github.com/Snider/Core/pkg/core"
    "github.com/Snider/Core/pkg/config"
)

type MyService struct {
    *core.ServiceRuntime[struct{}]
    data string
}

func NewMyService(c *core.Core) (any, error) {
    return &MyService{
        ServiceRuntime: core.NewServiceRuntime(c, struct{}{}),
        data:           "initialized",
    }, nil
}

func (s *MyService) OnStartup(ctx context.Context) error {
    // Startup logic
    return nil
}

func (s *MyService) OnShutdown(ctx context.Context) error {
    // Cleanup logic
    return nil
}

func (s *MyService) HandleIPCEvents(c *core.Core, msg core.Message) error {
    switch m := msg.(type) {
    case map[string]any:
        if m["action"] == "myservice.update" {
            s.data = m["data"].(string)
        }
    }
    return nil
}

func main() {
    c, err := core.New(
        core.WithService(config.Register),
        core.WithService(NewMyService),
        core.WithServiceLock(),
    )
    if err != nil {
        panic(err)
    }

    svc := core.MustServiceFor[*MyService](c, "main")
    _ = svc
}
```
