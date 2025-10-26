package main

import (
	"log"

	"github.com/Snider/Core/pkg/runtime"
)

// This main function demonstrates the "runtime" pattern of application startup.
// It uses a dedicated runtime package to build and inject all services statically.
// This provides a clear, compile-time view of the application's dependency graph,
// but is less dynamic than the core.WithService pattern.
func main() {
	// The runtime.New() function is the composition root of the application.
	// It instantiates all services and injects their dependencies.
	_, err := runtime.New()
	if err != nil {
		log.Fatalf("Failed to build application: %v", err)
	}

	// The fully wired application is now ready to be used.
	// For example, we can directly access services via the app container.
	log.Println("Application starting...")
	//if err := app.Docs.Show(); err != nil {
	//	log.Printf("Failed to show docs: %v", err)
	//}

	// In a real Wails application, you would pass the necessary components
	// from the 'app' container to the wails.Run() function.
	log.Println("Application finished.")
}
