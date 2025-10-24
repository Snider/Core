# Core.Framework

Core is an opinionated framework for building Go applications. It was written to ensure information exchange between the front-end and back-end of a Web3 environment.

There is no assumption of what front-end framework you use or what back-end framework you use.

However, Core provides a set of services beneficial for building applications for the Web3 environment.

With working defaults, Core is designed to be as simple as possible, yet powerful enough to be used in a wide variety of applications.

## Libraries used in Core
- [Wails.io](https://github.com/wailsapp/wails) - A framework for building desktop applications using Go and web technologies.
- [ProtonMail OpenPGP](https://github.com/ProtonMail/go-crypto)
## Features
- [Core](#core) - Manages application configuration
- [Core.Config](#coreconfig) - Manages application configuration
- [Core.Crypto](#corecrypto) - Manages cryptographic operations
- [Core.Database](#coredatabase) - Manages database operations
- [Core.IPC](#coreipc) - Manages inter-process communication
- [Core.Log](#corelog) - Manages logging
- [Core.Network](#corenetwork) - Manages network operations
- [Core.Storage](#corestorage) - Manages storage operations
- [Core.UI](#coreui) - Manages user interface operations
- [Core.Utils](#coreutils) - Manages utility operations
- [Core.Display](#coredisplay) - Manages Windows, system tray and their states


## Core

Core allows you to easily instantiate complex services within a locked-down environment enforcing logical enclaves.

### Basic Usage
```go
package main

import (
    "cmd/demo"
	
    "https://github.com/Snider/Core"
)

func main() {
	demoService := core.New(
		core.WithService(demo.Register),
		core.WithServiceLock(), // This can go anywhere in the chain to be the Final Service (provides security clarity)
	)
	demoService.Run() // handled by wails, not implemented yet
	demoService.Wait() // handled by wails, not implemented yet
	demoService.Close() // handled by wails, not implemented yet
}
```

### Multiple Services
```go
package main

import (
	"cmd/demo"

	"https://github.com/Snider/Core"
)

func main() {
	coreService := core.New(
		core.WithService(demo.Register),
		core.WithService(demo.RegisterDemo2),
		core.WithServiceLock(),
	)

	rickService := core.New(
		core.WithService(demo.Register),
		core.WithService(demo.RegisterDemo2),
		core.WithServiceLock(),
	)
	mortyService := core.New(
		core.WithService(demo.Register),
		core.WithService(demo.RegisterDemo2),
		core.WithServiceLock(),
	)

	core.Mod[API](coreService,"demo").name = "demo"
	core.Mod[API](rickService).name = "demo2"
	core.Mod[API](mortyService).name = "demo2"

}
```
### Registering Services
Core aunties that your application is not able to access any of the services that are not explicitly registered.

That said, 

## Core.Display

Display is a Wails.io service that provides the ability to open windows.
Core.Display remembers the locations of your users windows and will reopen them when the application is restarted.
Also allows you to adjust your windowing programmatically via JavaScript bindings, and for it to be persistent.

### Open A Window On Startup

```go
func (d *API) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	d.OpenWindow(
		OptName("main"),
		OptHeight(900),
		OptWidth(1280),
		OptURL("/"),
		OptTitle("Core"),
	)
	return nil
}
```

```go
window := d.OpenWindow(
	OptName("main"),
	OptHeight(900),
	OptWidth(1280),
	OptURL("/"),
	OptTitle("Core"),
)
```

### Full Wails Example
```go
package main

import (
	"embed"

	"github.com/Snider/Core"
	"github.com/Snider/Core/display"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:public/*
var assets embed.FS

func main() {

	app := application.New(application.Options{
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.RegisterService(application.NewService(core.Service(
		core.WithWails(app), // Provides the Wails application instance to core services
		core.WithAssets(assets), // Provides the embed.FS to core services
		core.WithService(display.Register), // Provides the ability to open windows
		core.WithService(config.Register), // Provides the ability to persist UI state (windows reopen where they closed)
		core.WithServiceLock(), // locks core from accepting new services blocking access to IPC
	)))

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
```
