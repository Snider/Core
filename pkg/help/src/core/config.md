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
  core "github.com/Snider/Core"
  config "github.com/Snider/Core/config"
)

app := core.New(
  core.WithService(config.Register),
  core.WithServiceLock(),
)

```

## Use
- Persist UI state automatically when using `Core.Display`.
- Read/write your own settings via the config API.

## API
- `Register(c *core.Core) error`
- `Get(path string, out any) error`
- `Set(path string, v any) error`
