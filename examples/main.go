package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/Snider/Core"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Embed a basic index.html for the example
//go:embed frontend/index.html
var assets embed.FS

func main() {
	// Create the Core service
	coreService := core.NewCoreService()

	// Display service information
	fmt.Printf("Service: %s\n", coreService.ServiceName())
	fmt.Printf("Version: %s\n\n", coreService.GetVersion())

	// Demonstrate service methods locally (without Wails UI)
	fmt.Println("=== Local Service Demo ===")
	fmt.Println("Greet('Wails Developer'):", coreService.Greet("Wails Developer"))
	fmt.Println("Add(15, 27):", coreService.Add(15, 27))
	fmt.Println("Multiply(8, 9):", coreService.Multiply(8, 9))
	
	result, err := coreService.Calculate("divide", 100, 4)
	if err != nil {
		fmt.Printf("Calculate error: %v\n", err)
	} else {
		fmt.Printf("Calculate('divide', 100, 4): %d\n", result)
	}
	fmt.Println()

	// Create a Wails application with the Core service
	fmt.Println("=== Starting Wails Application ===")
	fmt.Println("The Core service is now available to the frontend!")
	fmt.Println("Open the application window to interact with the service.")
	fmt.Println()

	app := application.New(application.Options{
		Name:        "Core Service Example",
		Description: "Example Wails v3 application using the Core service",
		Services: []application.Service{
			application.NewService(coreService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create application window
	app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:  "Core Service Example",
		Width:  800,
		Height: 600,
		URL:    "/frontend/index.html",
	})

	// Run the application
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
