// Package main provides the Core App — a native desktop application
// embedding Laravel via FrankenPHP inside a Wails v3 window.
//
// A single Go binary that boots the PHP runtime, extracts the embedded
// Laravel application, and serves it through FrankenPHP's ServeHTTP into
// a native webview via Wails v3's AssetOptions.Handler.
package main

import (
	"context"
	"log"
	"runtime"

	"github.com/host-uk/core/cmd/core-app/icons"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
	// Set up PHP handler (extracts Laravel, prepares env, inits FrankenPHP).
	handler, env, cleanup, err := NewPHPHandler()
	if err != nil {
		log.Fatalf("Failed to initialise PHP handler: %v", err)
	}
	defer cleanup()

	// Create the app service and native bridge.
	appService := NewAppService(env)
	bridge, err := NewNativeBridge(appService)
	if err != nil {
		log.Fatalf("Failed to start native bridge: %v", err)
	}
	defer bridge.Shutdown(context.Background())

	// Inject the bridge URL into the Laravel .env so PHP can call Go.
	if err := appendEnv(handler.laravelRoot, "NATIVE_BRIDGE_URL", bridge.URL()); err != nil {
		log.Printf("Warning: couldn't inject bridge URL into .env: %v", err)
	}

	app := application.New(application.Options{
		Name:        "Core App",
		Description: "Host UK Native App — Laravel powered by FrankenPHP",
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: handler,
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	})

	appService.app = app

	setupSystemTray(app)

	// Main application window
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "Core App",
		Width:            1200,
		Height:           800,
		URL:              "/",
		BackgroundColour: application.NewRGB(13, 17, 23),
	})

	log.Println("Starting Core App...")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// setupSystemTray configures the system tray icon and menu.
func setupSystemTray(app *application.App) {
	systray := app.SystemTray.New()
	systray.SetTooltip("Core App")

	if runtime.GOOS == "darwin" {
		systray.SetTemplateIcon(icons.TrayTemplate)
	} else {
		systray.SetDarkModeIcon(icons.TrayDark)
		systray.SetIcon(icons.TrayLight)
	}

	trayMenu := app.Menu.New()

	trayMenu.Add("Open Core App").OnClick(func(ctx *application.Context) {
		if w, ok := app.Window.Get("main"); ok {
			w.Show()
			w.Focus()
		}
	})

	trayMenu.AddSeparator()

	trayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	systray.SetMenu(trayMenu)
}
