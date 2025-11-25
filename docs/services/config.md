---
title: config
---
# Service: `config`

The `config` service provides a unified interface for managing application configuration. It handles retrieving and setting configuration values, persistent storage, and feature flags.

## Interfaces

### `type Config`

`Config` defines the contract for the configuration service.

```go
type Config interface {
    // Get retrieves a configuration value by key and stores it in the 'out' variable.
    Get(key string, out any) error

    // Set stores a configuration value by key.
    Set(key string, v any) error
}
```

## Standard Implementation

While `Config` is an interface, the standard implementation typically provides the following functionality:

- **Persistent Storage**: Saves configuration to disk (e.g., `config.json`).
- **Feature Flags**: Checking if specific application features are enabled.
- **Defaults**: Providing default values for configuration settings.

### Common Methods

Although not part of the minimal `Config` interface, implementations often provide:

- `Save() error`: Explicitly saves the current configuration to disk.
- `IsFeatureEnabled(feature string) bool`: Checks if a feature flag is active.
