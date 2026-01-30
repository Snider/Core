# Core Framework IPC Design

> Design document for refactoring CLI commands to use the Core framework's IPC system.

## Overview

The Core framework provides a dependency injection and inter-process communication (IPC) system originally designed for orchestrating services. This design extends the framework with request/response patterns and applies it to CLI commands.

Commands build "worker bundles" - sandboxed Core instances with specific services. The bundle configuration acts as a permissions layer: if a service isn't registered, that capability isn't available.

## Dispatch Patterns

Four patterns for service communication:

| Method | Behaviour | Returns | Use Case |
|--------|-----------|---------|----------|
| `ACTION` | Broadcast to all handlers | `error` | Events, notifications |
| `QUERY` | First responder wins | `(any, bool, error)` | Get data |
| `QUERYALL` | Broadcast, collect all | `([]any, error)` | Aggregate from multiple services |
| `PERFORM` | First responder executes | `(any, bool, error)` | Execute a task with side effects |

### ACTION (existing)

Fire-and-forget broadcast. All registered handlers receive the message. Errors are aggregated.

```go
c.ACTION(ActionServiceStartup{})
```

### QUERY (new)

Request data from services. Stops at first handler that returns `handled=true`.

```go
result, handled, err := c.QUERY(git.QueryStatus{Paths: paths})
if !handled {
    // No service registered to handle this query
}
statuses := result.([]git.RepoStatus)
```

### QUERYALL (new)

Broadcast query to all handlers, collect all responses. Useful for aggregating results from multiple services (e.g., multiple QA/lint tools).

```go
results, err := c.QUERYALL(qa.QueryLint{Paths: paths})
for _, r := range results {
    lint := r.(qa.LintResult)
    fmt.Printf("%s found %d issues\n", lint.Tool, len(lint.Issues))
}
```

### PERFORM (new)

Execute a task with side effects. Stops at first handler that returns `handled=true`.

```go
result, handled, err := c.PERFORM(agentic.TaskCommit{
    Path: repo.Path,
    Name: repo.Name,
})
if !handled {
    // Agentic service not in bundle - commits not available
}
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ cmd/dev/dev_work.go                                         │
│   - Builds worker bundle                                    │
│   - Triggers PERFORM(TaskWork{})                            │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│ cmd/dev/bundles.go                                          │
│   - NewWorkBundle() - git + agentic + dev                   │
│   - NewStatusBundle() - git + dev only                      │
│   - Bundle config = permissions                             │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│ pkg/dev/service.go                                          │
│   - Orchestrates workflow                                   │
│   - QUERY(git.QueryStatus{})                                │
│   - PERFORM(agentic.TaskCommit{})                           │
│   - PERFORM(git.TaskPush{})                                 │
└─────────────────────┬───────────────────────────────────────┘
                      │
        ┌─────────────┴─────────────┐
        ▼                           ▼
┌───────────────────┐     ┌───────────────────┐
│ pkg/git/service   │     │ pkg/agentic/svc   │
│                   │     │                   │
│ Queries:          │     │ Tasks:            │
│ - QueryStatus     │     │ - TaskCommit      │
│ - QueryDirtyRepos │     │ - TaskPrompt      │
│ - QueryAheadRepos │     │                   │
│                   │     │                   │
│ Tasks:            │     │                   │
│ - TaskPush        │     │                   │
│ - TaskPull        │     │                   │
└───────────────────┘     └───────────────────┘
```

## Permissions Model

Permissions are implicit through bundle configuration:

```go
// Full capabilities - can commit and push
func NewWorkBundle(opts WorkBundleOptions) (*framework.Runtime, error) {
    return framework.NewWithFactories(nil, map[string]framework.ServiceFactory{
        "dev":     func() (any, error) { return dev.NewService(opts.Dev)(nil) },
        "git":     func() (any, error) { return git.NewService(opts.Git)(nil) },
        "agentic": func() (any, error) { return agentic.NewService(opts.Agentic)(nil) },
    })
}

// Read-only - status queries only, no commits
func NewStatusBundle(opts StatusBundleOptions) (*framework.Runtime, error) {
    return framework.NewWithFactories(nil, map[string]framework.ServiceFactory{
        "dev": func() (any, error) { return dev.NewService(opts.Dev)(nil) },
        "git": func() (any, error) { return git.NewService(opts.Git)(nil) },
        // No agentic service - TaskCommit will be unhandled
    })
}
```

Service options provide fine-grained control:

```go
agentic.NewService(agentic.ServiceOptions{
    AllowEdit: false,  // Claude can only use read-only tools
})

agentic.NewService(agentic.ServiceOptions{
    AllowEdit: true,   // Claude can use Write/Edit tools
})
```

**Key principle**: Code never checks permissions explicitly. It dispatches actions and either they're handled or they're not. The bundle configuration is the single source of truth for what's allowed.

## Framework Changes

### New Types (interfaces.go)

```go
type Query interface{}
type Task interface{}

type QueryHandler func(*Core, Query) (any, bool, error)
type TaskHandler func(*Core, Task) (any, bool, error)
```

### Core Struct Additions (interfaces.go)

```go
type Core struct {
    // ... existing fields

    queryMu       sync.RWMutex
    queryHandlers []QueryHandler

    taskMu        sync.RWMutex
    taskHandlers  []TaskHandler
}
```

### New Methods (core.go)

```go
// QUERY - first responder wins
func (c *Core) QUERY(q Query) (any, bool, error)

// QUERYALL - broadcast, collect all responses
func (c *Core) QUERYALL(q Query) ([]any, error)

// PERFORM - first responder executes
func (c *Core) PERFORM(t Task) (any, bool, error)

// Registration
func (c *Core) RegisterQuery(h QueryHandler)
func (c *Core) RegisterTask(h TaskHandler)
```

### Re-exports (framework.go)

```go
type Query = core.Query
type Task = core.Task
type QueryHandler = core.QueryHandler
type TaskHandler = core.TaskHandler
```

## Service Implementation Pattern

Services register handlers during startup:

```go
func (s *Service) OnStartup(ctx context.Context) error {
    s.Core().RegisterAction(s.handleAction)
    s.Core().RegisterQuery(s.handleQuery)
    s.Core().RegisterTask(s.handleTask)
    return nil
}

func (s *Service) handleQuery(c *framework.Core, q framework.Query) (any, bool, error) {
    switch m := q.(type) {
    case QueryStatus:
        result := s.getStatus(m.Paths, m.Names)
        return result, true, nil
    case QueryDirtyRepos:
        return s.DirtyRepos(), true, nil
    }
    return nil, false, nil  // Not handled
}

func (s *Service) handleTask(c *framework.Core, t framework.Task) (any, bool, error) {
    switch m := t.(type) {
    case TaskPush:
        err := s.push(m.Path)
        return nil, true, err
    case TaskPull:
        err := s.pull(m.Path)
        return nil, true, err
    }
    return nil, false, nil  // Not handled
}
```

## Git Service Queries & Tasks

```go
// pkg/git/queries.go
type QueryStatus struct {
    Paths []string
    Names map[string]string
}

type QueryDirtyRepos struct{}
type QueryAheadRepos struct{}

// pkg/git/tasks.go
type TaskPush struct {
    Path string
    Name string
}

type TaskPull struct {
    Path string
    Name string
}

type TaskPushMultiple struct {
    Paths []string
    Names map[string]string
}
```

## Agentic Service Tasks

```go
// pkg/agentic/tasks.go
type TaskCommit struct {
    Path    string
    Name    string
    CanEdit bool
}

type TaskPrompt struct {
    Prompt       string
    WorkDir      string
    AllowedTools []string
}
```

## Dev Workflow Service

```go
// pkg/dev/tasks.go
type TaskWork struct {
    RegistryPath string
    StatusOnly   bool
    AutoCommit   bool
}

type TaskCommitAll struct {
    RegistryPath string
}

type TaskPushAll struct {
    RegistryPath string
    Force        bool
}
```

## Command Simplification

Before (dev_work.go - 327 lines of orchestration):

```go
func runWork(registryPath string, statusOnly, autoCommit bool) error {
    // Load registry
    // Get git status
    // Display table
    // Loop dirty repos, shell out to claude
    // Re-check status
    // Confirm push
    // Push repos
    // Handle diverged branches
    // ...
}
```

After (dev_work.go - minimal):

```go
func runWork(registryPath string, statusOnly, autoCommit bool) error {
    bundle, err := NewWorkBundle(WorkBundleOptions{
        RegistryPath: registryPath,
    })
    if err != nil {
        return err
    }

    ctx := context.Background()
    bundle.Core.ServiceStartup(ctx, nil)
    defer bundle.Core.ServiceShutdown(ctx)

    _, _, err = bundle.Core.PERFORM(dev.TaskWork{
        StatusOnly: statusOnly,
        AutoCommit: autoCommit,
    })
    return err
}
```

All orchestration logic moves to `pkg/dev/service.go` where it can be tested independently and reused.

## Implementation Tasks

1. **Framework Core** - Add Query, Task types and QUERY/QUERYALL/PERFORM methods
2. **Framework Re-exports** - Update framework.go with new types
3. **Git Service** - Add query and task handlers
4. **Agentic Service** - Add task handlers
5. **Dev Service** - Create workflow orchestration service
6. **Bundles** - Create bundle factories in cmd/dev/
7. **Commands** - Simplify cmd/dev/*.go to use bundles

## Future: CLI-Wide Runtime

Phase 2 will add a CLI-wide Core instance that:

- Handles signals (SIGINT, SIGTERM)
- Manages UI state
- Spawns worker bundles as "interactable elements"
- Provides cross-bundle communication

Worker bundles become sandboxed children of the CLI runtime, with the runtime controlling what capabilities each bundle receives.

## Testing

Each layer is independently testable:

- **Framework**: Unit tests for QUERY/QUERYALL/PERFORM dispatch
- **Services**: Unit tests with mock Core instances
- **Bundles**: Integration tests with real services
- **Commands**: E2E tests via CLI invocation

The permission model is testable by creating bundles with/without specific services and verifying behaviour.
