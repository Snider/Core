package main

import (
	"embed"

	"github.com/Snider/Core"
	"github.com/Snider/Core/config"
	"github.com/Snider/Core/crypt"
	"github.com/Snider/Core/display"
	"github.com/Snider/Core/help"
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

	coreService := core.New(
		core.WithWails(app),
		core.WithAssets(assets),
		core.WithService(config.New),
		core.WithService(display.New),
		core.WithService(crypt.New),
		core.WithService(help.New),
		core.WithServiceLock(),
	)

	app.RegisterService(application.NewService(coreService))

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
