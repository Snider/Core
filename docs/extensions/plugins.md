# Plugin System

The Plugin system (`pkg/plugin`) allows you to extend Core applications with HTTP-based plugins that register routes under `/api/{namespace}/{name}/`.

## Features

- Namespace-based organization
- HTTP handler registration
- Lifecycle hooks (OnRegister, OnUnregister)
- Wails service integration

## Plugin Interface

All plugins implement the `Plugin` interface:

```go
type Plugin interface {
    // Name returns the unique identifier for this plugin
    Name() string

    // Namespace returns the plugin's namespace (e.g., "core", "mining")
    Namespace() string

    // ServeHTTP handles HTTP requests routed to this plugin
    http.Handler

    // OnRegister is called when the plugin is registered
    OnRegister(ctx context.Context) error

    // OnUnregister is called when the plugin is being removed
    OnUnregister(ctx context.Context) error
}
```

## Using BasePlugin

For simple plugins, embed `BasePlugin`:

```go
import "github.com/Snider/Core/pkg/plugin"

func NewMyPlugin() *plugin.BasePlugin {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello from plugin!"))
    })

    return plugin.NewBasePlugin("myapp", "greeting", handler).
        WithDescription("A simple greeting plugin").
        WithVersion("1.0.0")
}
```

## Custom Plugin Implementation

For more control, implement the full interface:

```go
type DataPlugin struct {
    db *sql.DB
}

func (p *DataPlugin) Name() string      { return "data" }
func (p *DataPlugin) Namespace() string { return "myapp" }

func (p *DataPlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/users":
        p.handleUsers(w, r)
    case "/items":
        p.handleItems(w, r)
    default:
        http.NotFound(w, r)
    }
}

func (p *DataPlugin) OnRegister(ctx context.Context) error {
    // Initialize database connection
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        return err
    }
    p.db = db
    return nil
}

func (p *DataPlugin) OnUnregister(ctx context.Context) error {
    if p.db != nil {
        return p.db.Close()
    }
    return nil
}
```

## Plugin Info

Access plugin metadata:

```go
info := myPlugin.Info()
fmt.Println(info.Name)        // "greeting"
fmt.Println(info.Namespace)   // "myapp"
fmt.Println(info.Description) // "A simple greeting plugin"
fmt.Println(info.Version)     // "1.0.0"
```

## Wails Integration

Register plugins as Wails services:

```go
app := application.New(application.Options{
    Services: []application.Service{
        application.NewServiceWithOptions(
            myPlugin,
            plugin.ServiceOptionsForPlugin(myPlugin),
        ),
    },
})
```

## URL Routing

Plugins receive requests at:

```
/api/{namespace}/{name}/{path}
```

Examples:
- `/api/myapp/greeting/` → GreetingPlugin
- `/api/myapp/data/users` → DataPlugin (path: "/users")
- `/api/core/system/health` → SystemPlugin (path: "/health")

## Built-in Plugins

### System Plugin

Located at `pkg/plugin/builtin/system`:

```go
// Provides system information endpoints
/api/core/system/info    - Application info
/api/core/system/health  - Health check
```

## Plugin Router

The Router manages plugin registration:

```go
import "github.com/Snider/Core/pkg/plugin"

router := plugin.NewRouter()

// Register plugins
router.Register(ctx, myPlugin)
router.Register(ctx, dataPlugin)

// Get all registered plugins
plugins := router.List()

// Unregister a plugin
router.Unregister(ctx, "myapp", "greeting")
```

## Best Practices

1. **Use namespaces** to group related plugins
2. **Implement OnRegister** for initialization that can fail
3. **Implement OnUnregister** to clean up resources
4. **Return meaningful errors** from lifecycle hooks
5. **Use standard HTTP patterns** in ServeHTTP
