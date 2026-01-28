package main

import (
	"embed"
	"log"

	"github.com/host-uk/core
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:public
var assets embed.FS

func main() {
	// 1. Initialize Wails application
	app := application.New(application.Options{
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	// 2. Instantiate all services using the static runtime
	appServices, err := core.New()
	if err != nil {
		log.Fatalf("Failed to build application with static runtime: %v", err)
	}

	app.RegisterService(application.NewService(appServices.Config))
	app.RegisterService(application.NewService(appServices.Display))
	app.RegisterService(application.NewService(appServices.Help))
	app.RegisterService(application.NewService(appServices.Crypt))
	app.RegisterService(application.NewService(appServices.I18n))
	app.RegisterService(application.NewService(appServices.Workspace))

	log.Println("Application starting with static runtime...")

	err = app.Run()
	if err != nil {
		log.Fatalf("Wails application failed to run: %v", err)
	}
}
