---
title: Core.Docs
---

# Core.Docs

Short: In‑app help and deep‑links.

## Overview
Renders MkDocs content inside your app. Opens specific sections in new windows for contextual help.

## Setup
```go
package demo
import (
    "github.com/Snider/Core"
    "github.com/Snider/Core/docs"
)

core.New(
    core.WithService(docs.Register),
    core.WithServiceLock(),
)
```

## Use
- Open docs home in a window: `docs.Open()`
- Open a section: `docs.OpenAt("core/display#setup")`
- Use short, descriptive headings to create stable anchors.

## API
- `Register(c *core.Core) error`
- `Open()` — show docs home
- `OpenAt(anchor string)` — open specific section

## Notes
- Docs are built with MkDocs Material and included in the demo app assets.
- When viewed in the app, this documentation is served from Core.Docs and is bundled into the app binary by default.
