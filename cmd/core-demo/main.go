package main

import (
	"embed"
	"io/fs"
	"log"

	core "github.com/host-uk/core"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

//go:embed all:frontend/dist/frontend/browser
var assets embed.FS

// Default MCP port for the embedded server
const mcpPort = 9877

func main() {
	// Create the Core runtime with plugin support
	rt, err := core.NewRuntime()
	if err != nil {
		log.Fatal(err)
	}

	// Create the notifications service for native system notifications
	notifier := notifications.New()

	// Wire the notifier to the display service for native notifications
	rt.Display.SetNotifier(notifier)

	// Create the MCP bridge for Claude Code integration
	// This provides WebView access, console capture, window control, and process management
	mcpBridge := NewMCPBridge(mcpPort, rt.Display)

	// Collect all services including plugins
	// Display service registered separately so Wails calls its Startup() for tray/window
	services := []application.Service{
		application.NewService(rt.Runtime),
		application.NewService(rt.Display),
		application.NewService(notifier), // Native notifications
		application.NewService(rt.Docs),
		application.NewService(rt.Config),
		application.NewService(rt.I18n),
		application.NewService(rt.Help),
		application.NewService(rt.Crypt),
		application.NewService(rt.IDE),
		application.NewService(rt.Module),
		application.NewService(rt.Workspace),
		application.NewService(mcpBridge), // MCP Bridge for Claude Code
	}
	services = append(services, rt.PluginServices()...)

	// Strip the embed path prefix so files are served from root
	staticAssets, err := fs.Sub(assets, "frontend/dist/frontend/browser")
	if err != nil {
		log.Fatal(err)
	}

	app := application.New(application.Options{
		Services: services,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(staticAssets),
		},
	})

	log.Printf("Starting Core GUI with MCP server on port %d", mcpPort)

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
