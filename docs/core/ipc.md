# IPC & Actions

Core provides an inter-process communication system for services to communicate without tight coupling.

## Message Structure

```go
type Message struct {
    Type      string         // Message type identifier
    Data      map[string]any // Message payload
    Source    string         // Originating service (optional)
    Timestamp time.Time      // When message was created
}
```

## Sending Messages

```go
c.ACTION(core.Message{
    Type: "user.created",
    Data: map[string]any{
        "id":    "123",
        "email": "user@example.com",
    },
})
```

## Handling Messages

Register action handlers during service initialization:

```go
func NewNotificationService(c *core.Core) (any, error) {
    svc := &NotificationService{}

    // Register handler
    c.RegisterAction(func(c *core.Core, msg core.Message) error {
        return svc.handleAction(msg)
    })

    return svc, nil
}

func (s *NotificationService) handleAction(msg core.Message) error {
    switch msg.Type {
    case "user.created":
        email := msg.Data["email"].(string)
        return s.sendWelcomeEmail(email)
    }
    return nil
}
```

## Auto-Discovery

Services implementing `HandleIPCEvents` are automatically registered:

```go
type MyService struct{}

// Automatically registered when using WithService
func (s *MyService) HandleIPCEvents(c *core.Core, msg core.Message) error {
    // Handle messages
    return nil
}
```

## Common Patterns

### Request/Response

```go
// Sender
responseChan := make(chan any)
c.ACTION(core.Message{
    Type: "data.request",
    Data: map[string]any{
        "query":    "SELECT * FROM users",
        "response": responseChan,
    },
})
result := <-responseChan

// Handler
func (s *DataService) handleAction(msg core.Message) error {
    if msg.Type == "data.request" {
        query := msg.Data["query"].(string)
        respChan := msg.Data["response"].(chan any)

        result, err := s.execute(query)
        if err != nil {
            return err
        }

        respChan <- result
    }
    return nil
}
```

### Event Broadcasting

```go
// Broadcast to all listeners
c.ACTION(core.Message{
    Type: "system.config.changed",
    Data: map[string]any{
        "key":   "theme",
        "value": "dark",
    },
})
```

## Best Practices

1. **Use namespaced types** - `service.action` format
2. **Keep payloads simple** - Use primitive types when possible
3. **Handle errors** - Return errors from handlers
4. **Document message types** - Create constants for message types
