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
