package display

import (
	"context"
	"fmt"

	"github.com/Snider/Core"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

type ActionOpenWindow struct {
	Target string
}
type API struct {
	core *core.Core
}

var instance *API

func Register(c *core.Core) error {
	instance = &API{
		core: c,
	}
	if err := c.RegisterModule("display", instance); err != nil {
		return err
	}
	c.RegisterAction(handleActionCall)
	return nil
}

func handleActionCall(c *core.Core, msg core.Message) error {
	switch m := msg.(type) {
	case *ActionOpenWindow:
		instance.NewWithStruct(&Window{
			Name:   "main",
			Title:  "Core",
			Height: 900,
			Width:  1280,
			URL:    m.Target,
		})
		return nil
	case core.ActionServiceStartup:
		err := instance.ServiceStartup(context.Background(), application.ServiceOptions{})
		if err != nil {
			return err
		}
		c.App.Logger.Info("Display service started")
		return nil
	default:
		c.App.Logger.Error("Unknown message type", "type", fmt.Sprintf("%T", m))
		return nil
	}
}

func (d *API) analyzeScreens() {
	d.core.App.Logger.Info("Screen analysis", "count", len(d.core.App.Screen.GetAll()))

	primary := d.core.App.Screen.GetPrimary()
	if primary != nil {
		d.core.App.Logger.Info("Primary screen",
			"name", primary.Name,
			"size", fmt.Sprintf("%dx%d", primary.Size.Width, primary.Size.Height),
			"scaleFactor", primary.ScaleFactor,
			"workArea", primary.WorkArea,
		)
		scaleFactor := primary.ScaleFactor

		switch {
		case scaleFactor == 1.0:
			d.core.App.Logger.Info("Standard DPI display", "screen", primary.Name)
		case scaleFactor == 1.25:
			d.core.App.Logger.Info("125% scaled display", "screen", primary.Name)
		case scaleFactor == 1.5:
			d.core.App.Logger.Info("150% scaled display", "screen", primary.Name)
		case scaleFactor == 2.0:
			d.core.App.Logger.Info("High DPI display (200%)", "screen", primary.Name)
		default:
			d.core.App.Logger.Info("Custom scale display",
				"screen", primary.Name,
				"scale", scaleFactor,
			)
		}
	} else {
		d.core.App.Logger.Info("No primary screen found")
	}

	for i, screen := range d.core.App.Screen.GetAll() {
		d.core.App.Logger.Info("Screen details",
			"index", i,
			"name", screen.Name,
			"primary", screen.IsPrimary,
			"bounds", screen.Bounds,
			"scaleFactor", screen.ScaleFactor,
		)
	}
}

func (d *API) monitorScreenChanges() {
	// Monitor for screen configuration changes
	d.core.App.Event.OnApplicationEvent(events.Common.ThemeChanged, func(event *application.ApplicationEvent) {
		d.core.App.Logger.Info("Screen configuration changed")

		// Re-analyze screens
		d.core.App.Logger.Info("Updated screen count", "count", len(d.core.App.Screen.GetAll()))

		// Could reposition windows here if needed
	})
}

func (d *API) ShowEnvironmentDialog() {
	envInfo := d.core.App.Env.Info()

	details := fmt.Sprintf(`Environment Information:

Operating System: %s
Architecture: %s
Debug Mode: %t

Dark Mode: %t

Platform Information:`,
		envInfo.OS,
		envInfo.Arch,
		envInfo.Debug,
		d.core.App.Env.IsDarkMode()) // Use d.core.App

	// Add platform-specific details
	for key, value := range envInfo.PlatformInfo {
		details += fmt.Sprintf("\n%s: %v", key, value)
	}

	if envInfo.OSInfo != nil {
		details += fmt.Sprintf("\n\nOS Details:\nName: %s\nVersion: %s",
			envInfo.OSInfo.Name,
			envInfo.OSInfo.Version)
	}

	dialog := d.core.App.Dialog.Info()
	dialog.SetTitle("Environment Information")
	dialog.SetMessage(details)
	dialog.Show()
}

func (d *API) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	d.analyzeScreens()
	d.monitorScreenChanges()
	d.buildMenu()
	d.systemTray()
	d.OpenWindow(
		OptName("main"),
		OptHeight(900),
		OptWidth(1280),
		OptURL("/"),
		OptTitle("Core"),
	)
	return nil
}
