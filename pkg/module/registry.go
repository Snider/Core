package module

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// Registry manages module registration and provides unified API routing + UI assembly.
type Registry struct {
	mu            sync.RWMutex
	modules       map[string]*Module // key: code
	activeContext Context
	engine        *gin.Engine
	api           *gin.RouterGroup
	assetHandler  http.Handler
	appsDir       string // Directory to scan for dynamic modules
}

// NewRegistry creates a new module registry.
func NewRegistry() *Registry {
	engine := gin.New()
	engine.Use(gin.Recovery())

	r := &Registry{
		modules:       make(map[string]*Module),
		activeContext: ContextDefault,
		engine:        engine,
		api:           engine.Group("/api"),
		appsDir:       "apps",
	}

	// Root API endpoint lists modules
	r.api.GET("", r.handleModuleList)
	r.api.GET("/", r.handleModuleList)

	return r
}

// SetAppsDir sets the directory to scan for dynamic modules.
func (r *Registry) SetAppsDir(dir string) {
	r.appsDir = dir
}

// SetAssetHandler sets the fallback handler for non-API routes.
func (r *Registry) SetAssetHandler(h http.Handler) {
	r.assetHandler = h
	r.engine.NoRoute(func(c *gin.Context) {
		if r.assetHandler != nil {
			r.assetHandler.ServeHTTP(c.Writer, c.Request)
		} else {
			c.Status(http.StatusNotFound)
		}
	})
}

// Engine returns the underlying Gin engine.
func (r *Registry) Engine() *gin.Engine {
	return r.engine
}

// ServeHTTP implements http.Handler.
func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

// Register registers a module from its config.
func (r *Registry) Register(cfg Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	mod := &Module{Config: cfg}
	r.modules[cfg.Code] = mod

	return nil
}

// RegisterWithHandler registers a module with an HTTP handler for API routes.
func (r *Registry) RegisterWithHandler(cfg Config, handler http.Handler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	mod := &Module{Config: cfg, Handler: handler}
	r.modules[cfg.Code] = mod

	// Register API routes
	basePath := "/" + cfg.Namespace + "/" + cfg.Code
	r.api.Any(basePath, r.wrapHandler(handler))
	r.api.Any(basePath+"/*path", r.wrapHandler(handler))

	return nil
}

// RegisterGinModule registers a module that provides Gin routes.
func (r *Registry) RegisterGinModule(cfg Config, gm GinModule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	mod := &Module{Config: cfg}
	r.modules[cfg.Code] = mod

	// Let the module register its routes
	group := r.api.Group("/" + cfg.Namespace + "/" + cfg.Code)
	gm.RegisterRoutes(group)

	return nil
}

// RegisterFromJSON registers a module from JSON config.
func (r *Registry) RegisterFromJSON(data []byte) error {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("invalid module config: %w", err)
	}
	return r.Register(cfg)
}

// RegisterFromFile registers a module from a .itw3.json file.
func (r *Registry) RegisterFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading module file: %w", err)
	}
	return r.RegisterFromJSON(data)
}

// LoadApps scans the apps directory and loads all .itw3.json configs.
func (r *Registry) LoadApps(ctx context.Context) error {
	if r.appsDir == "" {
		return nil
	}

	// Check if apps directory exists
	if _, err := os.Stat(r.appsDir); os.IsNotExist(err) {
		return nil // No apps directory, that's fine
	}

	return filepath.Walk(r.appsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".itw3.json") {
			if err := r.RegisterFromFile(path); err != nil {
				// Log but don't fail on individual module errors
				fmt.Printf("Warning: failed to load module %s: %v\n", path, err)
			}
		}
		return nil
	})
}

// Unregister removes a module.
func (r *Registry) Unregister(code string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.modules, code)
}

// Get returns a module by code.
func (r *Registry) Get(code string) (*Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.modules[code]
	return m, ok
}

// SetContext changes the active UI context.
func (r *Registry) SetContext(ctx Context) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.activeContext = ctx
}

// GetContext returns the current context.
func (r *Registry) GetContext() Context {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.activeContext
}

// GetModules returns all registered module configs.
func (r *Registry) GetModules() []Config {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Config, 0, len(r.modules))
	for _, m := range r.modules {
		result = append(result, m.Config)
	}
	return result
}

// GetMenus returns aggregated menu items filtered by the active context.
func (r *Registry) GetMenus() []MenuItem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []MenuItem
	for _, mod := range r.modules {
		for _, menu := range mod.Config.Menu {
			if r.matchesContext(menu.Contexts) {
				filtered := r.filterMenuChildren(menu)
				result = append(result, filtered)
			}
		}
	}

	// Sort by order
	sort.Slice(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})

	return result
}

// GetRoutes returns aggregated routes filtered by the active context.
func (r *Registry) GetRoutes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Route
	for _, mod := range r.modules {
		for _, route := range mod.Config.Routes {
			if r.matchesContext(route.Contexts) {
				result = append(result, route)
			}
		}
	}
	return result
}

// GetUIConfig returns complete UI configuration for the current context.
func (r *Registry) GetUIConfig() UIConfig {
	return UIConfig{
		Context: r.GetContext(),
		Menus:   r.GetMenus(),
		Routes:  r.GetRoutes(),
		Modules: r.GetModules(),
	}
}

// UIConfig is the complete UI configuration for frontends.
type UIConfig struct {
	Context Context    `json:"context"`
	Menus   []MenuItem `json:"menus"`
	Routes  []Route    `json:"routes"`
	Modules []Config   `json:"modules"`
}

// matchesContext checks if item should show in current context.
func (r *Registry) matchesContext(contexts []Context) bool {
	if len(contexts) == 0 {
		return true // No restriction = show everywhere
	}
	for _, ctx := range contexts {
		if ctx == r.activeContext || ctx == ContextDefault {
			return true
		}
	}
	return false
}

// filterMenuChildren recursively filters menu children by context.
func (r *Registry) filterMenuChildren(menu MenuItem) MenuItem {
	if len(menu.Children) == 0 {
		return menu
	}
	filtered := menu
	filtered.Children = nil
	for _, child := range menu.Children {
		if r.matchesContext(child.Contexts) {
			filtered.Children = append(filtered.Children, r.filterMenuChildren(child))
		}
	}
	return filtered
}

// wrapHandler wraps an http.Handler for Gin.
func (r *Registry) wrapHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("path")
		if path == "" {
			path = "/"
		}
		c.Request.URL.Path = path
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// handleModuleList handles GET /api - returns list of modules.
func (r *Registry) handleModuleList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"modules": r.GetModules(),
		"context": r.GetContext(),
	})
}
