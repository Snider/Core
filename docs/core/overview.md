# Core Framework Overview

The Core package (`pkg/core`) is the foundation of the framework, providing service management, lifecycle handling, and inter-service communication.

## Creating a Core Instance

```go
import "github.com/Snider/Core/pkg/core"

c, err := core.New(
    core.WithAssets(assets),           // Embed frontend assets
    core.WithService(myServiceFactory), // Register services
    core.WithServiceLock(),            // Prevent late registration
)
```

## Options

| Option | Description |
|--------|-------------|
| `WithService(factory)` | Register a service with auto-discovered name |
| `WithName(name, factory)` | Register a service with explicit name |
| `WithAssets(fs)` | Embed filesystem for frontend assets |
| `WithServiceLock()` | Lock service registration after init |

## Service Factory Pattern

Services use factory functions for dependency injection:

```go
func NewMyService(c *core.Core) (any, error) {
    // Get dependencies
    config := core.MustServiceFor[*config.Service](c, "config")

    return &MyService{
        config: config,
    }, nil
}
```

## Features

Core includes a feature flag system:

```go
// Check if feature is enabled
if c.Features.IsEnabled("experimental") {
    // Use experimental feature
}

// Enable a feature
c.Features.Enable("experimental")
```

## Error Handling

Use the `E()` helper for contextual errors:

```go
import "github.com/Snider/Core/pkg/core"

func (s *Service) DoSomething() error {
    if err := someOperation(); err != nil {
        return core.E("Service.DoSomething", "operation failed", err)
    }
    return nil
}
```

## Best Practices

1. **Register all services before starting** - Use `WithServiceLock()` to catch mistakes
2. **Use factory functions** - Enables proper dependency injection
3. **Implement lifecycle interfaces** - For proper startup/shutdown
4. **Use typed service retrieval** - Catches type mismatches at compile time
