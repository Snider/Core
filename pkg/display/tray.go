package display

import (
	_ "embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// setupTray configures and creates the system tray icon and menu.
func (s *Service) systemTray() {

	systray := s.Core().App.SystemTray.New()
	systray.SetTooltip("Core")
	systray.SetLabel("Core")
	//appTrayIcon, _ := d.assets.ReadFile("assets/apptray.png")
	//
	//if runtime.GOOS == "darwin" {
	//	systray.SetTemplateIcon(appTrayIcon)
	//} else {
	//	// Support for light/dark mode icons
	//	systray.SetDarkModeIcon(appTrayIcon)
	//	systray.SetIcon(appTrayIcon)
	//}
	// Create a hidden window for the system tray menu to interact with
	trayWindow, _ := s.NewWithStruct(&Window{
		Name:      "system-tray",
		Title:     "System Tray Status",
		URL:       "system-tray.html",
		Width:     400,
		Frameless: true,
		Hidden:    true,
	})
	systray.AttachWindow(trayWindow).WindowOffset(5)

	// --- Build Tray Menu ---
	trayMenu := s.Core().App.Menu.New()
	trayMenu.Add("Open Desktop").OnClick(func(ctx *application.Context) {
		for _, window := range s.Core().App.Window.GetAll() {
			window.Show()
		}
	})
	trayMenu.Add("Close Desktop").OnClick(func(ctx *application.Context) {
		for _, window := range s.Core().App.Window.GetAll() {
			window.Hide()
		}
	})

	trayMenu.Add("Environment Info").OnClick(func(ctx *application.Context) {
		s.ShowEnvironmentDialog()
	})
	// Add brand-specific menu items
	//switch d.brand {
	//case AdminHub:
	//	trayMenu.Add("Manage Workspace").OnClick(func(ctx *application.Context) { /* TODO */ })
	//case ServerHub:
	//	trayMenu.Add("Server Control").OnClick(func(ctx *application.Context) { /* TODO */ })
	//case GatewayHub:
	//	trayMenu.Add("Routing Table").OnClick(func(ctx *application.Context) { /* TODO */ })
	//case DeveloperHub:
	//	trayMenu.Add("Debug Console").OnClick(func(ctx *application.Context) { /* TODO */ })
	//case ClientHub:
	//	trayMenu.Add("Connect").OnClick(func(ctx *application.Context) { /* TODO */ })
	//	trayMenu.Add("Disconnect").OnClick(func(ctx *application.Context) { /* TODO */ })
	//}

	trayMenu.AddSeparator()
	trayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		s.Core().App.Quit()
	})

	systray.SetMenu(trayMenu)
}
