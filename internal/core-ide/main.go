package main

import (
	"embed"
	"io/fs"
	"log"
	"runtime"

	"github.com/host-uk/core/internal/core-ide/icons"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist/wails-angular-template/browser
var assets embed.FS

// Default MCP port for the embedded server
const mcpPort = 9877

func main() {
	// Strip the embed path prefix so files are served from root
	staticAssets, err := fs.Sub(assets, "frontend/dist/wails-angular-template/browser")
	if err != nil {
		log.Fatal(err)
	}

	// Create the MCP bridge for Claude Code integration
	mcpBridge := NewMCPBridge(mcpPort)

	app := application.New(application.Options{
		Name:        "Core IDE",
		Description: "Host UK Core IDE - Development Environment",
		Services: []application.Service{
			application.NewService(&GreetService{}),
			application.NewService(mcpBridge),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(staticAssets),
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	})

	// System tray
	systray := app.SystemTray.New()
	systray.SetTooltip("Core IDE")

	if runtime.GOOS == "darwin" {
		systray.SetTemplateIcon(icons.AppTray)
	} else {
		systray.SetDarkModeIcon(icons.AppTray)
		systray.SetIcon(icons.AppTray)
	}

	// Tray panel window
	trayWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "tray-panel",
		Title:            "Core IDE",
		Width:            380,
		Height:           480,
		URL:              "/tray",
		Hidden:           true,
		Frameless:        true,
		BackgroundColour: application.NewRGB(26, 27, 38),
	})
	systray.AttachWindow(trayWindow).WindowOffset(5)

	// Tray menu
	trayMenu := app.Menu.New()
	trayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	systray.SetMenu(trayMenu)

	log.Println("Starting Core IDE...")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
