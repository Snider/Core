# Config Service

The Config service (`pkg/config`) provides unified configuration management with automatic persistence, feature flags, and XDG-compliant directory paths.

## Features

- JSON configuration with auto-save
- Feature flag management
- XDG Base Directory support
- Struct serialization helpers
- Type-safe get/set operations

## Basic Usage

```go
import "github.com/Snider/Core/pkg/config"

// Standalone usage
cfg, err := config.New()
if err != nil {
    log.Fatal(err)
}

// With Core framework
c, _ := core.New(
    core.WithService(config.Register),
)
cfg := core.MustServiceFor[*config.Service](c, "config")
```

## Get & Set Values

```go
// Set a value (auto-saves)
err := cfg.Set("language", "fr")

// Get a value
var lang string
err := cfg.Get("language", &lang)
```

Available configuration keys:

| Key | Type | Description |
|-----|------|-------------|
| `language` | string | UI language code |
| `default_route` | string | Default navigation route |
| `configDir` | string | Config files directory |
| `dataDir` | string | Data files directory |
| `cacheDir` | string | Cache directory |
| `workspaceDir` | string | Workspaces directory |

## Feature Flags

```go
// Enable a feature
cfg.EnableFeature("dark_mode")

// Check if enabled
if cfg.IsFeatureEnabled("dark_mode") {
    // Apply dark theme
}

// Disable a feature
cfg.DisableFeature("dark_mode")
```

## Struct Serialization

Store complex data structures in separate JSON files:

```go
type UserPrefs struct {
    Theme         string `json:"theme"`
    Notifications bool   `json:"notifications"`
}

// Save struct to config/user_prefs.json
prefs := UserPrefs{Theme: "dark", Notifications: true}
err := cfg.SaveStruct("user_prefs", prefs)

// Load struct from file
var loaded UserPrefs
err := cfg.LoadStruct("user_prefs", &loaded)
```

## Directory Paths

The service automatically creates XDG-compliant directories:

```go
// Access directory paths
fmt.Println(cfg.ConfigDir)    // ~/.config/lethean or ~/lethean/config
fmt.Println(cfg.DataDir)      // Data storage
fmt.Println(cfg.CacheDir)     // Cache files
fmt.Println(cfg.WorkspaceDir) // User workspaces
```

## Manual Save

Changes are auto-saved, but you can save explicitly:

```go
err := cfg.Save()
```

## Frontend Usage (TypeScript)

```typescript
import { Get, Set, IsFeatureEnabled } from '@bindings/config/service';

// Get configuration
const lang = await Get("language");

// Set configuration
await Set("default_route", "/dashboard");

// Check feature flag
if (await IsFeatureEnabled("dark_mode")) {
    applyDarkTheme();
}
```
