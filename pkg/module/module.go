// Package module provides a unified module system for Core applications.
// Modules can register API routes, UI menus, and configuration using the .itw3.json format.
package module

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Context represents the UI context (developer, retail, miner, etc.)
type Context string

const (
	ContextDefault   Context = "default"
	ContextDeveloper Context = "developer"
	ContextRetail    Context = "retail"
	ContextMiner     Context = "miner"
)

// ModuleType defines the type of module.
type ModuleType string

const (
	TypeCore ModuleType = "core" // Built-in core module
	TypeApp  ModuleType = "app"  // External application
	TypeBin  ModuleType = "bin"  // Binary/daemon wrapper
)

// Config is the .itw3.json format for module registration.
// This is the boundary format between Core and external modules.
type Config struct {
	Code        string         `json:"code"`                  // Unique identifier
	Type        ModuleType     `json:"type"`                  // core, app, bin
	Name        string         `json:"name"`                  // Display name
	Version     string         `json:"version"`               // Semantic version
	Namespace   string         `json:"namespace"`             // API/config namespace
	Description string         `json:"description,omitempty"` // Human description
	Author      string         `json:"author,omitempty"`
	Menu        []MenuItem     `json:"menu,omitempty"`      // UI menu contributions
	Routes      []Route        `json:"routes,omitempty"`    // UI route contributions
	Contexts    []Context      `json:"contexts,omitempty"`  // Which contexts this module supports
	Downloads   *Downloads     `json:"downloads,omitempty"` // Platform binaries
	App         *AppConfig     `json:"app,omitempty"`       // Web app config
	Depends     []string       `json:"depends,omitempty"`   // Module dependencies
	API         []APIEndpoint  `json:"api,omitempty"`       // API endpoint declarations
	Config      map[string]any `json:"config,omitempty"`    // Default configuration
}

// MenuItem represents a menu item contribution.
type MenuItem struct {
	ID          string     `json:"id"`
	Label       string     `json:"label"`
	Icon        string     `json:"icon,omitempty"`
	Action      string     `json:"action,omitempty"`      // Event name to emit
	Route       string     `json:"route,omitempty"`       // Frontend route
	Accelerator string     `json:"accelerator,omitempty"` // Keyboard shortcut
	Contexts    []Context  `json:"contexts,omitempty"`    // Show in these contexts
	Children    []MenuItem `json:"children,omitempty"`    // Submenu items
	Order       int        `json:"order,omitempty"`       // Sort order
	Separator   bool       `json:"separator,omitempty"`
}

// Route represents a UI route contribution.
type Route struct {
	Path      string    `json:"path"`
	Component string    `json:"component"` // Custom element or component
	Title     string    `json:"title,omitempty"`
	Icon      string    `json:"icon,omitempty"`
	Contexts  []Context `json:"contexts,omitempty"`
}

// APIEndpoint declares an API endpoint the module provides.
type APIEndpoint struct {
	Method      string `json:"method"` // GET, POST, etc.
	Path        string `json:"path"`   // Relative to /api/{namespace}/{code}
	Description string `json:"description,omitempty"`
}

// Downloads defines platform-specific binary downloads.
type Downloads struct {
	App     string            `json:"app,omitempty"` // Web app archive
	X86_64  *PlatformBinaries `json:"x86_64,omitempty"`
	Aarch64 *PlatformBinaries `json:"aarch64,omitempty"`
}

// PlatformBinaries defines OS-specific binary URLs.
type PlatformBinaries struct {
	Darwin  *BinaryInfo `json:"darwin,omitempty"`
	Linux   *BinaryInfo `json:"linux,omitempty"`
	Windows *BinaryInfo `json:"windows,omitempty"`
}

// BinaryInfo contains download info for a binary.
type BinaryInfo struct {
	URL      string `json:"url"`
	Checksum string `json:"checksum,omitempty"`
}

// AppConfig defines web app specific configuration.
type AppConfig struct {
	URL   string    `json:"url,omitempty"`
	Type  string    `json:"type,omitempty"` // spa, iframe, etc.
	Hooks []AppHook `json:"hooks,omitempty"`
}

// AppHook defines app lifecycle hooks.
type AppHook struct {
	Type string         `json:"type"` // rename, copy, etc.
	From string         `json:"from,omitempty"`
	To   string         `json:"to,omitempty"`
	Data map[string]any `json:"data,omitempty"`
}

// Module is a registered module with its config and optional handler.
type Module struct {
	Config  Config
	Handler http.Handler // Optional HTTP handler for API routes
}

// GinModule is a module that registers Gin routes.
type GinModule interface {
	RegisterRoutes(group *gin.RouterGroup)
}
