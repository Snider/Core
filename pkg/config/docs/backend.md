# Backend Documentation (`pkg/config`)

The `pkg/config` package provides a robust configuration management service for Go applications. It handles loading, saving, and accessing application settings, supporting both a main JSON configuration file and auxiliary data stored in various formats like YAML, INI, and XML.

## Core Concepts

### Service

The `Service` struct is the central point for all configuration-related operations. It manages:
- File paths and directories.
- Default values.
- Serialization and deserialization of data.
- Integration with the core application framework (if used).

The `Service` struct fields are automatically saved to and loaded from a JSON configuration file (`config.json`).

### Initialization

There are two ways to initialize the configuration service:

#### 1. Static Injection (`New`)

Use `config.New()` when you want to manage the service instance manually.

```go
cfg, err := config.New()
if err != nil {
    log.Fatalf("Failed to initialize config: %v", err)
}
// Use cfg...
```

#### 2. Dynamic Injection (`Register`)

Use `config.Register(coreInstance)` when integrating with the `core` package for dynamic dependency injection.

```go
// Assuming 'c' is your *core.Core instance
svc, err := config.Register(c)
if err != nil {
    // handle error
}
```

## Basic Configuration (Get/Set)

The service provides type-safe methods to get and set configuration values that correspond to the fields defined in the `Service` struct.

### Set a Value

```go
err := cfg.Set("language", "fr")
if err != nil {
    log.Fatalf("failed to set config value: %v", err)
}
```
*Note: `Set` automatically persists the changes to the `config.json` file.*

### Get a Value

```go
var lang string
err := cfg.Get("language", &lang)
if err != nil {
    log.Fatalf("failed to get config value: %v", err)
}
fmt.Printf("Language: %s\n", lang)
```

## Arbitrary Struct Persistence

You can save and load arbitrary Go structs to JSON files within the configuration directory using `SaveStruct` and `LoadStruct`. This is useful for complex data that doesn't fit into the main configuration schema.

### Saving a Struct

```go
type UserPreferences struct {
    Theme         string `json:"theme"`
    Notifications bool   `json:"notifications"`
}

prefs := UserPreferences{Theme: "dark", Notifications: true}

// Saves to <ConfigDir>/user_prefs.json
err := cfg.SaveStruct("user_prefs", prefs)
if err != nil {
    log.Printf("Error saving user preferences: %v", err)
}
```

### Loading a Struct

```go
var prefs UserPreferences
err := cfg.LoadStruct("user_prefs", &prefs)
if err != nil {
    log.Printf("Error loading user preferences: %v", err)
}
```

## Generic Key-Value Persistence

For more flexible data storage, the service supports generic key-value pairs in multiple file formats. The format is determined by the file extension.

### Supported Formats

- **JSON** (`.json`)
- **YAML** (`.yaml`, `.yml`)
- **INI** (`.ini`)
- **XML** (`.xml`)

### Saving Key-Values

```go
data := map[string]interface{}{
    "host": "localhost",
    "port": 8080,
}

// Save as YAML
if err := cfg.SaveKeyValues("database.yml", data); err != nil {
    log.Printf("Error saving database config: %v", err)
}
```

### Loading Key-Values

```go
dbConfig, err := cfg.LoadKeyValues("database.yml")
if err != nil {
    log.Printf("Error loading database config: %v", err)
}

port := dbConfig["port"]
```

## Configuration Directory

The service automatically resolves appropriate directories for storing configuration and data, respecting XDG standards on Linux/Unix-like systems and standard paths on other OSs.

- **Config Path**: `<ConfigDir>/config.json`
- **Config Dir**: `~/.local/share/lethean/config` (example on Linux)
- **Data Dir**: `~/.local/share/lethean/data`
- **Cache Dir**: `~/.cache/lethean`

These paths are accessible via the `Service` struct fields (e.g., `cfg.ConfigDir`).
