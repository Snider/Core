# Services

Services are the building blocks of a Core application. Each service encapsulates a specific domain of functionality.

## Creating a Service

```go
package myservice

import (
    "context"
    "github.com/Snider/Core/pkg/core"
)

type Service struct {
    core *core.Core
}

// Factory function for registration
func NewService(c *core.Core) (any, error) {
    return &Service{core: c}, nil
}

// Implement Startable for startup logic
func (s *Service) OnStartup(ctx context.Context) error {
    // Initialize resources
    return nil
}

// Implement Stoppable for cleanup
func (s *Service) OnShutdown(ctx context.Context) error {
    // Release resources
    return nil
}
```

## Registering Services

```go
c, err := core.New(
    // Auto-discover name from package path
    core.WithService(myservice.NewService),

    // Explicit name
    core.WithName("custom", func(c *core.Core) (any, error) {
        return &CustomService{}, nil
    }),
)
```

## Retrieving Services

```go
// Safe retrieval with error
svc, err := core.ServiceFor[*myservice.Service](c, "myservice")
if err != nil {
    log.Printf("Service not found: %v", err)
}

// Must retrieval (panics if not found)
svc := core.MustServiceFor[*myservice.Service](c, "myservice")
```

## Service Dependencies

Services can depend on other services:

```go
func NewOrderService(c *core.Core) (any, error) {
    // Get required dependencies
    userSvc := core.MustServiceFor[*user.Service](c, "user")
    paymentSvc := core.MustServiceFor[*payment.Service](c, "payment")

    return &OrderService{
        users:    userSvc,
        payments: paymentSvc,
    }, nil
}
```

!!! warning "Dependency Order"
    Register dependencies before services that use them. Core does not automatically resolve dependency order.

## Exposing to Frontend

Services are automatically exposed to the frontend via Wails bindings:

```go
// Go service method
func (s *Service) GetUser(id string) (*User, error) {
    return s.db.FindUser(id)
}
```

```typescript
// TypeScript (auto-generated)
import { GetUser } from '@bindings/myservice/service';

const user = await GetUser("123");
```

## Testing Services

```go
func TestMyService(t *testing.T) {
    // Create mock core
    c, _ := core.New()

    // Create service
    svc, err := NewService(c)
    if err != nil {
        t.Fatal(err)
    }

    // Test methods
    result := svc.(*Service).DoSomething()
    assert.Equal(t, expected, result)
}
```
