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
	"github.com/Snider/Core/docs"
)

core.New(
	core.WithService(docs.Register), 
	core.WithServiceLock(),
)
```

## Use
- Open Help home in a window: `help.Open()`
- Open a section: `help.OpenAt("core/display#setup")`
- Use short, descriptive headings to create stable anchors.

## API
- `Register(c *core.Core) error`
- `Open()` — show help home
- `OpenAt(anchor string)` — open specific section

## Notes
- Help is built with MkDocs Material and included in the demo app assets.
- When viewed in the app, this documentation is served from Core.Help and is bundled into the app binary by default.
