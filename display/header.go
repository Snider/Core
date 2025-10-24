package display

import (
	"github.com/Snider/Core"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Brand defines the type for different application brands.
type Brand string

const (
	AdminHub     Brand = "admin-hub"
	ServerHub    Brand = "server-hub"
	GatewayHub   Brand = "gateway-hub"
	DeveloperHub Brand = "developer-hub"
	ClientHub    Brand = "client-hub"
)

// Service manages all OS-level UI interactions (menus, windows, tray).
// It is the main entry point for all display-related operations.
type API struct {
	// --- Injected Dependencies ---
	core *core.Core

	windowHandles map[string]*application.WebviewWindow
}
