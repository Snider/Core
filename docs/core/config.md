---
title: Core.Config
---

# Core.Config

Short: App config and UI state persistence.

## Overview
Stores and retrieves configuration, including window positions/sizes and user prefs.

## Setup
```go
import (
  "github.com/Snider/Core"
  "github.com/Snider/Core/config"
)
// Example of static dependency injection
configService, err := config.New()
if err != nil {
  // handle error
}
app := core.New(
  core.WithService(configService),
  core.WithServiceLock(),
)

// Example of dynamic dependency injection (used with core.WithService)
// app := core.New(
//   core.WithService(config.Register),
//   core.WithServiceLock(),
// )
```

## Use
- Persist UI state automatically when using `Core.Display`.
- Read/write your own settings via the config API.

## API
- `New() (*Service, error)`
- `Register(c *core.Core) (any, error)`
- `(s *Service) Save() error`
- `(s *Service) Get(key string, out any) error`
- `(s *Service) Set(key string, v any) error`
