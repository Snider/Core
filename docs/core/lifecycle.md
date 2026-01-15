# Service Lifecycle

Core provides lifecycle hooks for services to initialize and clean up resources.

## Lifecycle Interfaces

### Startable

Called when the application starts:

```go
type Startable interface {
    OnStartup(ctx context.Context) error
}
```

### Stoppable

Called when the application shuts down:

```go
type Stoppable interface {
    OnShutdown(ctx context.Context) error
}
```

## Implementation Example

```go
type DatabaseService struct {
    db *sql.DB
}

func (s *DatabaseService) OnStartup(ctx context.Context) error {
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        return err
    }

    // Verify connection
    if err := db.PingContext(ctx); err != nil {
        return err
    }

    s.db = db
    return nil
}

func (s *DatabaseService) OnShutdown(ctx context.Context) error {
    if s.db != nil {
        return s.db.Close()
    }
    return nil
}
```

## Lifecycle Order

1. **Registration**: Services registered via `core.New()`
2. **Wails Binding**: Services bound to Wails app
3. **Startup**: `OnStartup()` called for each Startable service
4. **Running**: Application runs
5. **Shutdown**: `OnShutdown()` called for each Stoppable service

## Context Usage

The context passed to lifecycle methods includes:

- Cancellation signal for graceful shutdown
- Deadline for timeout handling

```go
func (s *Service) OnStartup(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-s.initialize():
        return nil
    }
}
```

## Error Handling

If `OnStartup` returns an error, the application will fail to start:

```go
func (s *Service) OnStartup(ctx context.Context) error {
    if err := s.validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}
```

## Best Practices

1. **Keep startup fast** - Defer heavy initialization
2. **Handle context cancellation** - Support graceful shutdown
3. **Clean up resources** - Always implement OnShutdown for services with resources
4. **Log lifecycle events** - Helps with debugging
