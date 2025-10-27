---
title: Core.Help
---

# Core.Help

Short: In‑app help and deep‑links.

## Overview
Renders MkDocs content inside your app. Opens specific sections in new windows for contextual help.

## Setup
```go
package demo
import (
    "github.com/Snider/Core"
    "github.com/Snider/Core/pkg/help"
)

core.New(
    core.WithService(help.Register),
    core.WithServiceLock(),
)
```

## Use
- Open docs home in a window: `help.Show()`
- Open a section: `help.ShowAt("core/display#setup")`
- Use short, descriptive headings to create stable anchors.

## API
- `New(cfg core.Config, disp core.Display) (*Service, error)`
- `Register(c *core.Core) (any, error)`
- `(s *Service) Show() error`
- `(s *Service) ShowAt(anchor string) error`

## Notes
- Docs are built with MkDocs Material and included in the demo app assets.
- When viewed in the app, this documentation is served from Core.Help and is bundled into the app binary by default.
