package plugin

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// GinPlugin is a plugin that registers routes on a Gin router group.
type GinPlugin interface {
	Plugin
	// RegisterRoutes registers the plugin's routes on the provided router group.
	// The group is already prefixed with /api/{namespace}/{name}
	RegisterRoutes(group *gin.RouterGroup)
}

// Router manages plugin registration and provides a Gin-based HTTP router.
// It implements http.Handler and can be used as the Wails asset handler middleware.
type Router struct {
	mu           sync.RWMutex
	plugins      map[string]Plugin // key: "namespace/name"
	byNS         map[string][]Plugin
	engine       *gin.Engine
	api          *gin.RouterGroup
	assetHandler http.Handler // fallback to Wails asset server
	route        string       // set by Wails on startup
}

// NewRouter creates a new plugin router with a Gin engine.
func NewRouter() *Router {
	// Use gin.New() for custom middleware control
	engine := gin.New()
	engine.Use(gin.Recovery())

	r := &Router{
		plugins: make(map[string]Plugin),
		byNS:    make(map[string][]Plugin),
		engine:  engine,
		api:     engine.Group("/api"),
	}

	// Register the plugins list endpoint
	r.api.GET("", r.handlePluginList)
	r.api.GET("/", r.handlePluginList)

	return r
}

// statusCapturingWriter wraps http.ResponseWriter to track if status was set.
type statusCapturingWriter struct {
	http.ResponseWriter
	statusSet bool
}

func (w *statusCapturingWriter) WriteHeader(code int) {
	w.statusSet = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusCapturingWriter) Write(b []byte) (int, error) {
	if !w.statusSet {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// SetAssetHandler sets the fallback handler for non-API routes (Wails assets).
func (r *Router) SetAssetHandler(h http.Handler) {
	r.assetHandler = h
	// Set up fallback to asset handler for non-API routes
	r.engine.NoRoute(func(c *gin.Context) {
		if r.assetHandler != nil {
			// Wrap the writer to ensure proper status handling
			// Gin's NoRoute may interfere with implicit status 200
			w := &statusCapturingWriter{ResponseWriter: c.Writer}
			r.assetHandler.ServeHTTP(w, c.Request)
		} else {
			c.Status(http.StatusNotFound)
		}
	})
}

// Engine returns the underlying Gin engine for advanced configuration.
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// ServiceStartup is called by Wails when the service starts.
func (r *Router) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	r.route = options.Route
	return nil
}

// Register adds a plugin to the router.
func (r *Router) Register(ctx context.Context, p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := p.Namespace() + "/" + p.Name()

	// Unregister existing plugin if present
	if old, exists := r.plugins[key]; exists {
		old.OnUnregister(ctx)
	}

	// Register the plugin
	if err := p.OnRegister(ctx); err != nil {
		return err
	}

	r.plugins[key] = p

	// Update namespace index
	if _, exists := r.plugins[key]; !exists {
		r.byNS[p.Namespace()] = append(r.byNS[p.Namespace()], p)
	}

	// If it's a GinPlugin, let it register its routes
	if gp, ok := p.(GinPlugin); ok {
		group := r.api.Group("/" + p.Namespace() + "/" + p.Name())
		gp.RegisterRoutes(group)
	} else {
		// For regular plugins, create a catch-all route that delegates to ServeHTTP
		basePath := "/" + p.Namespace() + "/" + p.Name()
		r.api.Any(basePath, r.wrapPlugin(p))
		r.api.Any(basePath+"/*path", r.wrapPlugin(p))
	}

	return nil
}

// wrapPlugin wraps a Plugin's ServeHTTP for use with Gin.
func (r *Router) wrapPlugin(p Plugin) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Strip the prefix to get the sub-path
		path := c.Param("path")
		if path == "" {
			path = "/"
		}
		c.Request.URL.Path = path
		p.ServeHTTP(c.Writer, c.Request)
	}
}

// Unregister removes a plugin from the router.
// Note: Gin doesn't support removing routes, so this only removes from our registry.
// A restart is required for route changes to take effect.
func (r *Router) Unregister(ctx context.Context, namespace, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := namespace + "/" + name
	p, exists := r.plugins[key]
	if !exists {
		return nil
	}

	if err := p.OnUnregister(ctx); err != nil {
		return err
	}

	delete(r.plugins, key)

	// Update namespace index
	plugins := r.byNS[namespace]
	for i, plugin := range plugins {
		if plugin.Name() == name {
			r.byNS[namespace] = append(plugins[:i], plugins[i+1:]...)
			break
		}
	}

	return nil
}

// Get returns a plugin by namespace and name.
func (r *Router) Get(namespace, name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[namespace+"/"+name]
	return p, ok
}

// ListByNamespace returns all plugins in a namespace.
func (r *Router) ListByNamespace(namespace string) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byNS[namespace]
}

// List returns info about all registered plugins.
func (r *Router) List() []PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]PluginInfo, 0, len(r.plugins))
	for _, p := range r.plugins {
		if bp, ok := p.(*BasePlugin); ok {
			infos = append(infos, bp.Info())
		} else {
			infos = append(infos, PluginInfo{
				Name:      p.Name(),
				Namespace: p.Namespace(),
			})
		}
	}
	return infos
}

// handlePluginList handles GET /api - returns list of plugins.
func (r *Router) handlePluginList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"plugins": r.List(),
	})
}

// ServeHTTP implements http.Handler - delegates to Gin engine.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

// ServiceOptions returns the Wails service options for the router.
func (r *Router) ServiceOptions() application.ServiceOptions {
	return application.ServiceOptions{
		Route: "/api",
	}
}
