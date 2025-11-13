package main

import (
	"embed"
	"log"

	"github.com/Snider/Core/runtime"
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

	rt, err := runtime.New(app)
	if err != nil {
		log.Fatal(err)
	}

	app.RegisterService(application.NewService(rt))

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
