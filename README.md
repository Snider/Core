# Core

A Helper for GoLang projects, who also use, but not exclusive to Wails.io v3+

You need a file called apptray.png in your assets folder
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