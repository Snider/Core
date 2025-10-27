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
	// 1. Initialize Wails application
	app := application.New(application.Options{
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	// 2. Instantiate all services using the static runtime
	appServices, err := runtime.New()
	if err != nil {
		log.Fatalf("Failed to build application with static runtime: %v", err)
	}

	// 3. Explicitly register each service with Wails
	// This demonstrates how services instantiated via static DI are bound to Wails.
	app.RegisterService(application.NewService(appServices.Config))
	app.RegisterService(application.NewService(appServices.Display))
	app.RegisterService(application.NewService(appServices.Help))
	app.RegisterService(application.NewService(appServices.Crypt))
	app.RegisterService(application.NewService(appServices.I18n))
	app.RegisterService(application.NewService(appServices.Workspace))

	log.Println("Application starting with static runtime...")

	// Example usage of a service from the static runtime:
	// if err := appServices.Help.Show(); err != nil {
	// 	log.Printf("Didn't show help: %v", err)
	// }

	// 4. Run the Wails application
	err = app.Run()
	if err != nil {
		log.Fatalf("Wails application failed to run: %v", err)
	}
}
