// Package main provides the BugSETI system tray application.
// BugSETI - "Distributed Bug Fixing like SETI@home but for code"
//
// The application runs as a system tray app that:
// - Pulls OSS issues from GitHub
// - Uses AI to prepare context for each issue
// - Presents issues to users for fixing
// - Automates PR submission
package main

import (
	"embed"
	"io/fs"
	"log"
	"runtime"

	"github.com/host-uk/core/cmd/bugseti/icons"
	"github.com/host-uk/core/internal/bugseti"
	"github.com/host-uk/core/internal/bugseti/updater"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist/bugseti/browser
var assets embed.FS

func main() {
	// Strip the embed path prefix so files are served from root
	staticAssets, err := fs.Sub(assets, "frontend/dist/bugseti/browser")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the config service
	configService := bugseti.NewConfigService()
	if err := configService.Load(); err != nil {
		log.Printf("Warning: Could not load config: %v", err)
	}

	// Initialize core services
	notifyService := bugseti.NewNotifyService()
	statsService := bugseti.NewStatsService(configService)
	fetcherService := bugseti.NewFetcherService(configService, notifyService)
	queueService := bugseti.NewQueueService(configService)
	seederService := bugseti.NewSeederService(configService)
	submitService := bugseti.NewSubmitService(configService, notifyService, statsService)
	versionService := bugseti.NewVersionService()

	// Initialize update service
	updateService, err := updater.NewService(configService)
	if err != nil {
		log.Printf("Warning: Could not initialize update service: %v", err)
	}

	// Create the tray service (we'll set the app reference later)
	trayService := NewTrayService(nil)

	// Build services list
	services := []application.Service{
		application.NewService(configService),
		application.NewService(notifyService),
		application.NewService(statsService),
		application.NewService(fetcherService),
		application.NewService(queueService),
		application.NewService(seederService),
		application.NewService(submitService),
		application.NewService(versionService),
		application.NewService(trayService),
	}

	// Add update service if available
	if updateService != nil {
		services = append(services, application.NewService(updateService))
	}

	// Create the application
	app := application.New(application.Options{
		Name:        "BugSETI",
		Description: "Distributed Bug Fixing - like SETI@home but for code",
		Services:    services,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(staticAssets),
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	})

	// Set the app reference and services in tray service
	trayService.app = app
	trayService.SetServices(fetcherService, queueService, configService, statsService)

	// Set up system tray
	setupSystemTray(app, fetcherService, queueService, configService)

	// Start update service background checker
	if updateService != nil {
		updateService.Start()
	}

	log.Println("Starting BugSETI...")
	log.Println("  - System tray active")
	log.Println("  - Waiting for issues...")
	log.Printf("  - Version: %s (%s)", bugseti.GetVersion(), bugseti.GetChannel())

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	// Stop update service on exit
	if updateService != nil {
		updateService.Stop()
	}
}

// setupSystemTray configures the system tray icon and menu
func setupSystemTray(app *application.App, fetcher *bugseti.FetcherService, queue *bugseti.QueueService, config *bugseti.ConfigService) {
	systray := app.SystemTray.New()
	systray.SetTooltip("BugSETI - Distributed Bug Fixing")

	// Set tray icon based on OS
	if runtime.GOOS == "darwin" {
		systray.SetTemplateIcon(icons.TrayTemplate)
	} else {
		systray.SetDarkModeIcon(icons.TrayDark)
		systray.SetIcon(icons.TrayLight)
	}

	// Create tray panel window (workbench preview)
	trayWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "tray-panel",
		Title:            "BugSETI",
		Width:            420,
		Height:           520,
		URL:              "/tray",
		Hidden:           true,
		Frameless:        true,
		BackgroundColour: application.NewRGB(22, 27, 34),
	})
	systray.AttachWindow(trayWindow).WindowOffset(5)

	// Create main workbench window
	workbenchWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "workbench",
		Title:            "BugSETI Workbench",
		Width:            1200,
		Height:           800,
		URL:              "/workbench",
		Hidden:           true,
		BackgroundColour: application.NewRGB(22, 27, 34),
	})

	// Create settings window
	settingsWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "settings",
		Title:            "BugSETI Settings",
		Width:            600,
		Height:           500,
		URL:              "/settings",
		Hidden:           true,
		BackgroundColour: application.NewRGB(22, 27, 34),
	})

	// Create onboarding window
	onboardingWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "onboarding",
		Title:            "Welcome to BugSETI",
		Width:            700,
		Height:           600,
		URL:              "/onboarding",
		Hidden:           true,
		BackgroundColour: application.NewRGB(22, 27, 34),
	})

	// Build tray menu
	trayMenu := app.Menu.New()

	// Status item (dynamic)
	statusItem := trayMenu.Add("Status: Idle")
	statusItem.SetEnabled(false)

	trayMenu.AddSeparator()

	// Start/Pause toggle
	startPauseItem := trayMenu.Add("Start Fetching")
	startPauseItem.OnClick(func(ctx *application.Context) {
		if fetcher.IsRunning() {
			fetcher.Pause()
			startPauseItem.SetLabel("Start Fetching")
			statusItem.SetLabel("Status: Paused")
		} else {
			fetcher.Start()
			startPauseItem.SetLabel("Pause")
			statusItem.SetLabel("Status: Running")
		}
	})

	trayMenu.AddSeparator()

	// Current Issue
	currentIssueItem := trayMenu.Add("Current Issue: None")
	currentIssueItem.OnClick(func(ctx *application.Context) {
		if issue := queue.CurrentIssue(); issue != nil {
			workbenchWindow.Show()
			workbenchWindow.Focus()
		}
	})

	// Open Workbench
	trayMenu.Add("Open Workbench").OnClick(func(ctx *application.Context) {
		workbenchWindow.Show()
		workbenchWindow.Focus()
	})

	trayMenu.AddSeparator()

	// Settings
	trayMenu.Add("Settings...").OnClick(func(ctx *application.Context) {
		settingsWindow.Show()
		settingsWindow.Focus()
	})

	// Stats submenu
	statsMenu := trayMenu.AddSubmenu("Stats")
	statsMenu.Add("Issues Fixed: 0").SetEnabled(false)
	statsMenu.Add("PRs Merged: 0").SetEnabled(false)
	statsMenu.Add("Repos Contributed: 0").SetEnabled(false)

	trayMenu.AddSeparator()

	// Quit
	trayMenu.Add("Quit BugSETI").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	systray.SetMenu(trayMenu)

	// Check if onboarding needed
	if !config.IsOnboarded() {
		onboardingWindow.Show()
		onboardingWindow.Focus()
	}
}
