---
title: Core.Help
---

# Overview

Core is an opinionated framework for building Go desktop apps with Wails, providing a small set of focused modules you can mix into your app. It ships with sensible defaults and a demo app that doubles as in‑app help.

- Site: [https://dappco.re](https://dappco.re)
- Help: [https://core.help](https://core.help)
- Repo: [github.com:Snider/Core](https://github.com/Snider/Core)

## Modules

- Core — framework bootstrap and service container
- Core.Config — app and UI state persistence
- Core.Crypt — keys, encrypt/decrypt, sign/verify
- Core.Display — windows, tray, window state
- Core.Docs — in‑app help and deep‑links
- Core.IO — local/remote filesystem helpers
- Core.Workspace — projects and paths

## Quick start
```go

import (
    "github.com/wailsapp/wails/v3/pkg/application"
    "github.com/Snider/Core"
    "github.com/Snider/Core/display"
)

func main() {
    app := core.New(
        core.WithServiceLock(),
    )
    wailsApp := application.NewWithOptions(&application.Options{
        Bind: []interface{}{app},
    })
    wailsApp.Run()
}
```

## Services
```go

import (
    core "github.com/Snider/Core"
)

// Register your service
func Register(c *core.Core) error {
    return c.RegisterService("demo", &Demo{core: c})
}
```
