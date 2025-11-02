package main

import (
	"embed"
	"log"

	"github.com/Snider/Core/pkg/runtime"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:public
var assets embed.FS

func main() {
	rt, err := runtime.New()
	if err != nil {
		log.Fatal(err)
	}

	app := application.New(application.Options{
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Services: []application.Service{
			rt,
		},
	})

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
