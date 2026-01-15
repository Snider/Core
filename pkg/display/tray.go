package display

import (
	"embed"
	"fmt"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed assets/apptray.png
var assets embed.FS

// activeTray holds the reference to the system tray for management.
var activeTray *application.SystemTray

// systemTray configures and creates the system tray icon and menu.
func (s *Service) systemTray() {

	systray := s.app.SystemTray().New()
	activeTray = systray
	systray.SetTooltip("Core")
	systray.SetLabel("Core")

	// Load and set tray icon
	appTrayIcon, err := assets.ReadFile("assets/apptray.png")
	if err == nil {
		if runtime.GOOS == "darwin" {
			systray.SetTemplateIcon(appTrayIcon)
		} else {
			// Support for light/dark mode icons
			systray.SetDarkModeIcon(appTrayIcon)
			systray.SetIcon(appTrayIcon)
		}
	}
	// Create a hidden window for the system tray menu to interact with
	trayWindow, _ := s.NewWithStruct(&Window{
		Name:      "system-tray",
		Title:     "System Tray Status",
		URL:       "/system-tray",
		Width:     400,
		Frameless: true,
		Hidden:    true,
	})
	systray.AttachWindow(trayWindow).WindowOffset(5)

	// --- Build Tray Menu ---
	trayMenu := s.app.Menu().New()
	trayMenu.Add("Open Desktop").OnClick(func(ctx *application.Context) {
		for _, window := range s.app.Window().GetAll() {
			window.Show()
		}
	})
	trayMenu.Add("Close Desktop").OnClick(func(ctx *application.Context) {
		for _, window := range s.app.Window().GetAll() {
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
		s.app.Quit()
	})

	systray.SetMenu(trayMenu)
}

// SetTrayIcon sets the system tray icon from raw PNG data.
func (s *Service) SetTrayIcon(iconData []byte) error {
	if activeTray == nil {
		return fmt.Errorf("system tray not initialized")
	}
	if runtime.GOOS == "darwin" {
		activeTray.SetTemplateIcon(iconData)
	} else {
		activeTray.SetIcon(iconData)
	}
	return nil
}

// SetTrayTooltip sets the system tray tooltip text.
func (s *Service) SetTrayTooltip(tooltip string) error {
	if activeTray == nil {
		return fmt.Errorf("system tray not initialized")
	}
	activeTray.SetTooltip(tooltip)
	return nil
}

// SetTrayLabel sets the system tray label text.
func (s *Service) SetTrayLabel(label string) error {
	if activeTray == nil {
		return fmt.Errorf("system tray not initialized")
	}
	activeTray.SetLabel(label)
	return nil
}

// TrayMenuItem represents a menu item for the system tray.
type TrayMenuItem struct {
	Label    string         `json:"label"`
	Type     string         `json:"type,omitempty"`    // "normal", "separator", "checkbox", "radio"
	Checked  bool           `json:"checked,omitempty"` // for checkbox/radio items
	Disabled bool           `json:"disabled,omitempty"`
	Tooltip  string         `json:"tooltip,omitempty"`
	Submenu  []TrayMenuItem `json:"submenu,omitempty"`
	ActionID string         `json:"actionId,omitempty"` // ID for callback
}

// trayMenuCallbacks stores callbacks for tray menu items.
var trayMenuCallbacks = make(map[string]func())

// SetTrayMenu sets the system tray menu from a list of menu items.
func (s *Service) SetTrayMenu(items []TrayMenuItem) error {
	if activeTray == nil {
		return fmt.Errorf("system tray not initialized")
	}

	menu := s.app.Menu().New()
	s.buildTrayMenu(menu, items)
	activeTray.SetMenu(menu)
	return nil
}

// buildTrayMenu recursively builds a menu from TrayMenuItem items.
func (s *Service) buildTrayMenu(menu *application.Menu, items []TrayMenuItem) {
	for _, item := range items {
		switch item.Type {
		case "separator":
			menu.AddSeparator()
		case "checkbox":
			menuItem := menu.AddCheckbox(item.Label, item.Checked)
			if item.Disabled {
				menuItem.SetEnabled(false)
			}
			if item.ActionID != "" {
				actionID := item.ActionID
				menuItem.OnClick(func(ctx *application.Context) {
					if cb, ok := trayMenuCallbacks[actionID]; ok {
						cb()
					}
				})
			}
		default:
			if len(item.Submenu) > 0 {
				submenu := menu.AddSubmenu(item.Label)
				s.buildTrayMenu(submenu, item.Submenu)
			} else {
				menuItem := menu.Add(item.Label)
				if item.Disabled {
					menuItem.SetEnabled(false)
				}
				if item.Tooltip != "" {
					menuItem.SetTooltip(item.Tooltip)
				}
				if item.ActionID != "" {
					actionID := item.ActionID
					menuItem.OnClick(func(ctx *application.Context) {
						if cb, ok := trayMenuCallbacks[actionID]; ok {
							cb()
						}
					})
				}
			}
		}
	}
}

// RegisterTrayMenuCallback registers a callback for a tray menu action ID.
func (s *Service) RegisterTrayMenuCallback(actionID string, callback func()) {
	trayMenuCallbacks[actionID] = callback
}

// GetTrayInfo returns information about the current tray state.
func (s *Service) GetTrayInfo() map[string]any {
	if activeTray == nil {
		return map[string]any{"active": false}
	}
	return map[string]any{
		"active": true,
	}
}
