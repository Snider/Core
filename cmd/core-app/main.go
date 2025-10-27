package main

import (
	"embed"
	"log"

	"github.com/Snider/Core"
	"github.com/Snider/Core/config"
	"github.com/Snider/Core/crypt"
	"github.com/Snider/Core/display"
	"github.com/Snider/Core/help"
	"github.com/Snider/Core/i18n"
	"github.com/Snider/Core/workspace"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:public
var assets embed.FS

func main() {

	app := application.New(application.Options{
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	coreService, err := core.New(
		core.WithWails(app),
		core.WithAssets(assets),
		core.WithService(config.Register),
		core.WithService(display.Register),
		core.WithService(crypt.Register),
		core.WithService(help.Register),
		core.WithService(i18n.Register),
		core.WithService(workspace.Register),
		core.WithServiceLock(),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Core services: %v", err)
	}

	app.RegisterService(application.NewService(coreService))

	err = app.Run()
	if err != nil {
		panic(err)
	}
}
