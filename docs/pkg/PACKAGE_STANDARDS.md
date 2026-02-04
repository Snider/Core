# Core Package Standards

This document defines the standards for creating packages in the Core framework. The `pkg/i18n` package is the reference implementation; all new packages should follow its patterns.

## Package Structure

A well-structured Core package follows this layout:

```
pkg/mypackage/
├── types.go           # Public types, constants, interfaces
├── service.go         # Service struct with framework integration
├── mypackage.go       # Global convenience functions
├── actions.go         # ACTION messages for Core IPC (if needed)
├── hooks.go           # Event hooks with atomic handlers (if needed)
├── [feature].go       # Additional feature files
├── [feature]_test.go  # Tests alongside implementation
└── service_test.go    # Service tests
```

## Core Principles

1. **Service-oriented**: Packages expose a `Service` struct that integrates with the Core framework
2. **Thread-safe**: All public APIs must be safe for concurrent use
3. **Global convenience**: Provide package-level functions that use a default service instance
4. **Options pattern**: Use functional options for configuration
5. **ACTION-based IPC**: Communicate via Core's ACTION system, not callbacks

---

## Service Pattern

### Service Struct

Embed `framework.ServiceRuntime[T]` for Core integration:

```go
// pkg/mypackage/service.go
package mypackage

import (
    "sync"
    "github.com/host-uk/core/pkg/framework"
)

// Service provides mypackage functionality with Core integration.
type Service struct {
    *framework.ServiceRuntime[Options]

    // Internal state (protected by mutex)
    data map[string]any
    mu   sync.RWMutex
}

// Options configures the service.
type Options struct {
    // Document each option
    BufferSize int
    EnableFoo  bool
}
```

### Service Factory

Create a factory function for Core registration:

```go
// NewService creates a service factory for Core registration.
//
//    core, _ := framework.New(
//        framework.WithName("mypackage", mypackage.NewService(mypackage.Options{})),
//    )
func NewService(opts Options) func(*framework.Core) (any, error) {
    return func(c *framework.Core) (any, error) {
        // Apply defaults
        if opts.BufferSize == 0 {
            opts.BufferSize = DefaultBufferSize
        }

        svc := &Service{
            ServiceRuntime: framework.NewServiceRuntime(c, opts),
            data:           make(map[string]any),
        }
        return svc, nil
    }
}
```

### Lifecycle Hooks

Implement `framework.Startable` and/or `framework.Stoppable`:

```go
// OnStartup implements framework.Startable.
func (s *Service) OnStartup(ctx context.Context) error {
    // Register query/task handlers
    s.Core().RegisterQuery(s.handleQuery)
    s.Core().RegisterAction(s.handleAction)
    return nil
}

// OnShutdown implements framework.Stoppable.
func (s *Service) OnShutdown(ctx context.Context) error {
    // Cleanup resources
    return nil
}
```

---

## Global Default Pattern

Following `pkg/i18n`, provide a global default service with atomic access:

```go
// pkg/mypackage/mypackage.go
package mypackage

import (
    "sync"
    "sync/atomic"

    "github.com/host-uk/core/pkg/framework"
)

// Global default service
var (
    defaultService atomic.Pointer[Service]
    defaultOnce    sync.Once
    defaultErr     error
)

// Default returns the global service instance.
// Returns nil if not initialised.
func Default() *Service {
    return defaultService.Load()
}

// SetDefault sets the global service instance.
// Thread-safe. Panics if s is nil.
func SetDefault(s *Service) {
    if s == nil {
        panic("mypackage: SetDefault called with nil service")
    }
    defaultService.Store(s)
}

// Init initialises the default service with a Core instance.
func Init(c *framework.Core) error {
    defaultOnce.Do(func() {
        factory := NewService(Options{})
        svc, err := factory(c)
        if err != nil {
            defaultErr = err
            return
        }
        defaultService.Store(svc.(*Service))
    })
    return defaultErr
}
```

### Global Convenience Functions

Expose the most common operations at package level:

```go
// ErrServiceNotInitialised is returned when the service is not initialised.
var ErrServiceNotInitialised = errors.New("mypackage: service not initialised")

// DoSomething performs an operation using the default service.
func DoSomething(arg string) (string, error) {
    svc := Default()
    if svc == nil {
        return "", ErrServiceNotInitialised
    }
    return svc.DoSomething(arg)
}
```

---

## Options Pattern

Use functional options for complex configuration:

```go
// Option configures a Service during construction.
type Option func(*Service)

// WithBufferSize sets the buffer size.
func WithBufferSize(size int) Option {
    return func(s *Service) {
        s.bufSize = size
    }
}

// WithFoo enables foo feature.
func WithFoo(enabled bool) Option {
    return func(s *Service) {
        s.fooEnabled = enabled
    }
}

// New creates a service with options.
func New(opts ...Option) (*Service, error) {
    s := &Service{
        bufSize: DefaultBufferSize,
    }
    for _, opt := range opts {
        opt(s)
    }
    return s, nil
}
```

---

## ACTION Messages (IPC)

For services that need to communicate events, define ACTION message types:

```go
// pkg/mypackage/actions.go
package mypackage

import "time"

// ActionItemCreated is broadcast when an item is created.
type ActionItemCreated struct {
    ID        string
    Name      string
    CreatedAt time.Time
}

// ActionItemUpdated is broadcast when an item changes.
type ActionItemUpdated struct {
    ID      string
    Changes map[string]any
}

// ActionItemDeleted is broadcast when an item is removed.
type ActionItemDeleted struct {
    ID string
}
```

Dispatch actions via `s.Core().ACTION()`:

```go
func (s *Service) CreateItem(name string) (*Item, error) {
    item := &Item{ID: generateID(), Name: name}

    // Store item...

    // Broadcast to listeners
    s.Core().ACTION(ActionItemCreated{
        ID:        item.ID,
        Name:      item.Name,
        CreatedAt: time.Now(),
    })

    return item, nil
}
```

Consumers register handlers:

```go
core.RegisterAction(func(c *framework.Core, msg framework.Message) error {
    switch m := msg.(type) {
    case mypackage.ActionItemCreated:
        log.Printf("Item created: %s", m.Name)
    case mypackage.ActionItemDeleted:
        log.Printf("Item deleted: %s", m.ID)
    }
    return nil
})
```

---

## Hooks Pattern

For user-customisable behaviour, use atomic handlers (see `pkg/i18n/hooks.go`):

```go
// pkg/mypackage/hooks.go
package mypackage

import (
    "sync/atomic"
)

// ErrorHandler is called when an error occurs.
type ErrorHandler func(err error)

var errorHandler atomic.Value // stores ErrorHandler

// OnError registers an error handler.
// Thread-safe. Pass nil to clear.
func OnError(h ErrorHandler) {
    if h == nil {
        errorHandler.Store((ErrorHandler)(nil))
        return
    }
    errorHandler.Store(h)
}

// dispatchError calls the registered error handler.
func dispatchError(err error) {
    v := errorHandler.Load()
    if v == nil {
        return
    }
    h, ok := v.(ErrorHandler)
    if !ok || h == nil {
        return
    }
    h(err)
}
```

---

## Thread Safety

### Mutex Patterns

Use `sync.RWMutex` for state that is read more than written:

```go
type Service struct {
    data map[string]any
    mu   sync.RWMutex
}

func (s *Service) Get(key string) (any, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    v, ok := s.data[key]
    return v, ok
}

func (s *Service) Set(key string, value any) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
}
```

### Atomic Values

Use `atomic.Pointer[T]` for single values accessed frequently:

```go
var config atomic.Pointer[Config]

func GetConfig() *Config {
    return config.Load()
}

func SetConfig(c *Config) {
    config.Store(c)
}
```

---

## Error Handling

### Error Types

Define package-level errors:

```go
// Errors
var (
    ErrNotFound    = errors.New("mypackage: not found")
    ErrInvalidArg  = errors.New("mypackage: invalid argument")
    ErrNotRunning  = errors.New("mypackage: not running")
)
```

### Wrapped Errors

Use `fmt.Errorf` with `%w` for context:

```go
func (s *Service) Load(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    // ...
}
```

### Error Struct (optional)

For errors needing additional context:

```go
type ServiceError struct {
    Op      string // Operation that failed
    Path    string // Resource path
    Err     error  // Underlying error
}

func (e *ServiceError) Error() string {
    return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
}

func (e *ServiceError) Unwrap() error {
    return e.Err
}
```

---

## Testing

### Test File Organisation

Place tests alongside implementation:

```
mypackage.go      → mypackage_test.go
service.go        → service_test.go
buffer.go         → buffer_test.go
```

### Test Helpers

Create helpers for common setup:

```go
func newTestService(t *testing.T) (*Service, *framework.Core) {
    t.Helper()

    core, err := framework.New(
        framework.WithName("mypackage", NewService(Options{})),
    )
    require.NoError(t, err)

    svc, err := framework.ServiceFor[*Service](core, "mypackage")
    require.NoError(t, err)

    return svc, core
}
```

### Test Naming Convention

Use descriptive subtests:

```go
func TestService_DoSomething(t *testing.T) {
    t.Run("valid input", func(t *testing.T) {
        // ...
    })

    t.Run("empty input returns error", func(t *testing.T) {
        // ...
    })

    t.Run("concurrent access", func(t *testing.T) {
        // ...
    })
}
```

### Testing Actions

Verify ACTION broadcasts:

```go
func TestService_BroadcastsActions(t *testing.T) {
    core, _ := framework.New(
        framework.WithName("mypackage", NewService(Options{})),
    )

    var received []ActionItemCreated
    var mu sync.Mutex

    core.RegisterAction(func(c *framework.Core, msg framework.Message) error {
        if m, ok := msg.(ActionItemCreated); ok {
            mu.Lock()
            received = append(received, m)
            mu.Unlock()
        }
        return nil
    })

    svc, _ := framework.ServiceFor[*Service](core, "mypackage")
    svc.CreateItem("test")

    mu.Lock()
    assert.Len(t, received, 1)
    assert.Equal(t, "test", received[0].Name)
    mu.Unlock()
}
```

---

## Documentation

### Package Doc

Every package needs a doc comment in the main file:

```go
// Package mypackage provides functionality for X.
//
// # Getting Started
//
//    svc, err := mypackage.New()
//    result := svc.DoSomething("input")
//
// # Core Integration
//
//    core, _ := framework.New(
//        framework.WithName("mypackage", mypackage.NewService(mypackage.Options{})),
//    )
package mypackage
```

### Function Documentation

Document public functions with examples:

```go
// DoSomething performs X operation with the given input.
// Returns ErrInvalidArg if input is empty.
//
//    result, err := svc.DoSomething("hello")
//    if err != nil {
//        return err
//    }
func (s *Service) DoSomething(input string) (string, error) {
    // ...
}
```

---

## Checklist

When creating a new package, ensure:

- [ ] `Service` struct embeds `framework.ServiceRuntime[Options]`
- [ ] `NewService()` factory function for Core registration
- [ ] `Default()` / `SetDefault()` with `atomic.Pointer`
- [ ] Package-level convenience functions
- [ ] Thread-safe public APIs (mutex or atomic)
- [ ] ACTION messages for events (if applicable)
- [ ] Hooks with atomic handlers (if applicable)
- [ ] Comprehensive tests with helpers
- [ ] Package documentation with examples

## Reference Implementations

- **`pkg/i18n`** - Full reference with handlers, modes, hooks, grammar
- **`pkg/process`** - Simpler example with ACTION events and runner orchestration
- **`pkg/cli`** - Service integration with runtime lifecycle

---

## Background Operations

For long-running operations that could block the UI, use the framework's background task mechanism.

### Principles

1. **Non-blocking**: Long-running operations must not block the main IPC thread.
2. **Lifecycle Events**: Use `PerformAsync` to automatically broadcast start and completion events.
3. **Progress Reporting**: Services should broadcast `ActionTaskProgress` for granular updates.

### Using PerformAsync

The `Core.PerformAsync(task)` method runs any registered task in a background goroutine and returns a unique `TaskID` immediately.

```go
// From the frontend or another service
taskID := core.PerformAsync(git.TaskPush{Path: "/repo"})
// taskID is returned immediately, e.g., "task-123"
```

The framework automatically broadcasts lifecycle actions:
- `ActionTaskStarted`: When the background goroutine begins.
- `ActionTaskCompleted`: When the task finishes (contains Result and Error).

### Reporting Progress

For very long operations, the service handler should broadcast progress:

```go
func (s *Service) handleTask(c *framework.Core, t framework.Task) (any, bool, error) {
    switch m := t.(type) {
    case MyLongTask:
        // Optional: If you need to report progress, you might need to pass
        // a TaskID or use a specific progress channel.
        // For now, simple tasks just use ActionTaskCompleted.
        return s.doLongWork(m), true, nil
    }
    return nil, false, nil
}
```

### Implementing Background-Safe Handlers

Ensure that handlers for long-running tasks:
1. Use `context.Background()` or a long-lived context, as the request context might expire.
2. Are thread-safe and don't hold global locks for the duration of the work.
3. Do not use interactive CLI functions like `cli.Scanln` if they are intended for GUI use.
