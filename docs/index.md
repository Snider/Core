---
title: Core
---

# Overview

Core is a framework for building Go desktop apps with Wails. It provides a set of modules that can be used to quickly build robust and maintainable applications.

- Site: [https://dappco.re](https://dappco.re)
- Help: [https://core.help](https://core.help)
- Repo: [github.com:Snider/Core](https://github.com/Snider/Core)

## Modules

The Core framework is organized into the following modules, all located under the `pkg/` directory:

- `pkg/config` — Application and UI state persistence.
- `pkg/crypt` — Key management, encryption/decryption, and signing/verification.
- `pkg/display` — Window and tray management.
- `pkg/e` — Standardized error handling.
- `pkg/help` — In-app help and documentation.
- `pkg/i18n` — Internationalization and localization.
- `pkg/io` — Local and remote filesystem abstractions.
- `pkg/runtime` — Service container and application lifecycle management.
- `pkg/workspace` — Workspace and project management.

## Quick Start

Here is a basic example of how to create a new Core application:

```go
package main

import (
	"log"

	"github.com/Snider/Core/pkg/runtime"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
	rt, err := runtime.New()
	if err != nil {
		log.Fatal(err)
	}

	app := application.New(application.Options{
		Services: []application.Service{
			rt,
		},
	})

	app.Run()
}
```

See the navigation on the left for more detailed information on each module.
