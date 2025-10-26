---
title: Core.Docs
---

# Core.Docs

Short: In‑app help and deep‑links.

## Overview
Renders MkDocs content inside your app. Opens specific sections in new windows for contextual help.

## Setup
```go
import (
  core "github.com/Snider/Core"
  docs "github.com/Snider/Core/docs"
)

app := core.New(
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
- You are viewing Core.Docs right now, this website is bundled into the app binary by default.
