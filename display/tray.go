package display

import (
	_ "embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// setupTray configures and creates the system tray icon and menu.
func (d *API) systemTray() {

	systray := d.core.App.SystemTray.New()
	systray.SetTooltip("Lethean Desktop")
	systray.SetLabel("hey")
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
	trayWindow := d.core.App.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "System Tray Status",
		URL:       "/#/system-tray",
		Width:     400,
		Frameless: true,
		Hidden:    true,
	})
	systray.AttachWindow(trayWindow).WindowOffset(5)

	// --- Build Tray Menu ---
	trayMenu := d.core.App.Menu.New()
	trayMenu.Add("Open Desktop").OnClick(func(ctx *application.Context) {
		for _, window := range d.core.App.Window.GetAll() {
			window.Show()
		}
	})
	trayMenu.Add("Close Desktop").OnClick(func(ctx *application.Context) {
		for _, window := range d.core.App.Window.GetAll() {
			window.Hide()
		}
	})

	trayMenu.Add("Environment Info").OnClick(func(ctx *application.Context) {
		d.ShowEnvironmentDialog()
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
		d.core.App.Quit()
	})

	systray.SetMenu(trayMenu)
}
