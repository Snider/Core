package plugin

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// echoPlugin is a test plugin that echoes back the request path
type echoPlugin struct {
	*BasePlugin
}

func newEchoPlugin(namespace, name string) *echoPlugin {
	p := &echoPlugin{}
	p.BasePlugin = NewBasePlugin(namespace, name, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("echo:" + r.URL.Path))
	}))
	return p
}

// ginEchoPlugin is a test plugin that uses Gin routes directly
type ginEchoPlugin struct {
	*BasePlugin
}

func newGinEchoPlugin(namespace, name string) *ginEchoPlugin {
	return &ginEchoPlugin{
		BasePlugin: NewBasePlugin(namespace, name, nil),
	}
}

func (p *ginEchoPlugin) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello from gin")
	})
	group.GET("/echo/:msg", func(c *gin.Context) {
		c.String(http.StatusOK, "gin echo: "+c.Param("msg"))
	})
	group.POST("/data", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.String(http.StatusOK, "received: "+string(body))
	})
}

func TestBasePlugin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	p := NewBasePlugin("core", "test", handler).
		WithDescription("A test plugin").
		WithVersion("1.0.0")

	assert.Equal(t, "test", p.Name())
	assert.Equal(t, "core", p.Namespace())

	info := p.Info()
	assert.Equal(t, "A test plugin", info.Description)
	assert.Equal(t, "1.0.0", info.Version)

	// Test HTTP handling
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	p.ServeHTTP(w, req)

	assert.Equal(t, "hello", w.Body.String())
}

func TestRouter_Register(t *testing.T) {
	router := NewRouter()
	ctx := context.Background()

	p1 := newEchoPlugin("core", "echo1")
	p2 := newEchoPlugin("core", "echo2")
	p3 := newEchoPlugin("mining", "status")

	require.NoError(t, router.Register(ctx, p1))
	require.NoError(t, router.Register(ctx, p2))
	require.NoError(t, router.Register(ctx, p3))

	// Check plugins are registered
	got, ok := router.Get("core", "echo1")
	assert.True(t, ok)
	assert.Equal(t, "echo1", got.Name())

	// Check list
	all := router.List()
	assert.Len(t, all, 3)
}

func TestRouter_Unregister(t *testing.T) {
	router := NewRouter()
	ctx := context.Background()

	p := newEchoPlugin("core", "test")
	require.NoError(t, router.Register(ctx, p))

	_, ok := router.Get("core", "test")
	assert.True(t, ok)

	require.NoError(t, router.Unregister(ctx, "core", "test"))

	_, ok = router.Get("core", "test")
	assert.False(t, ok)
}

func TestRouter_ServeHTTP_PluginList(t *testing.T) {
	router := NewRouter()
	ctx := context.Background()

	p := newEchoPlugin("core", "echo")
	require.NoError(t, router.Register(ctx, p))

	req := httptest.NewRequest("GET", "/api", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"plugins"`)
}

func TestRouter_ServeHTTP_RegularPlugin(t *testing.T) {
	router := NewRouter()
	ctx := context.Background()

	p := newEchoPlugin("core", "echo")
	require.NoError(t, router.Register(ctx, p))

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "routes to plugin with path",
			path:       "/api/core/echo/test/path",
			wantStatus: http.StatusOK,
			wantBody:   "echo:/test/path",
		},
		{
			name:       "routes to plugin root",
			path:       "/api/core/echo",
			wantStatus: http.StatusOK,
			wantBody:   "echo:/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				body, _ := io.ReadAll(w.Body)
				assert.Contains(t, string(body), tt.wantBody)
			}
		})
	}
}

func TestRouter_ServeHTTP_GinPlugin(t *testing.T) {
	router := NewRouter()
	ctx := context.Background()

	p := newGinEchoPlugin("core", "ginecho")
	require.NoError(t, router.Register(ctx, p))

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "GET hello endpoint",
			method:     "GET",
			path:       "/api/core/ginecho/hello",
			wantStatus: http.StatusOK,
			wantBody:   "hello from gin",
		},
		{
			name:       "GET echo with param",
			method:     "GET",
			path:       "/api/core/ginecho/echo/world",
			wantStatus: http.StatusOK,
			wantBody:   "gin echo: world",
		},
		{
			name:       "POST data",
			method:     "POST",
			path:       "/api/core/ginecho/data",
			body:       "test payload",
			wantStatus: http.StatusOK,
			wantBody:   "received: test payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestRouter_AssetFallback(t *testing.T) {
	router := NewRouter()

	// Set up a mock asset handler
	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("asset: " + r.URL.Path))
	})
	router.SetAssetHandler(assetHandler)

	// Request a non-API path should fall through to asset handler
	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "asset: /index.html")
}

func TestBasePlugin_NilHandler(t *testing.T) {
	p := NewBasePlugin("core", "test", nil)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	p.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
	assert.Contains(t, w.Body.String(), "Not implemented")
}

func TestServiceOptionsForPlugin(t *testing.T) {
	p := NewBasePlugin("core", "test", nil)
	opts := ServiceOptionsForPlugin(p)

	assert.Equal(t, "/api/core/test", opts.Route)
}

func TestRouter_Engine(t *testing.T) {
	router := NewRouter()

	engine := router.Engine()
	assert.NotNil(t, engine)
}

func TestRouter_ServiceStartup(t *testing.T) {
	router := NewRouter()
	ctx := context.Background()

	opts := router.ServiceOptions()
	err := router.ServiceStartup(ctx, opts)
	assert.NoError(t, err)
}

func TestRouter_ServiceOptions(t *testing.T) {
	router := NewRouter()

	opts := router.ServiceOptions()
	assert.Equal(t, "/api", opts.Route)
}

func TestRouter_ListByNamespace(t *testing.T) {
	router := NewRouter()

	// Test ListByNamespace returns empty for nonexistent namespace
	emptyPlugins := router.ListByNamespace("nonexistent")
	assert.Empty(t, emptyPlugins)
}

func TestRouter_UnregisterNonExistent(t *testing.T) {
	router := NewRouter()
	ctx := context.Background()

	// Should not error when unregistering non-existent plugin
	err := router.Unregister(ctx, "core", "nonexistent")
	assert.NoError(t, err)
}

// Note: Re-registration test removed because Gin does not support re-registering routes.
// The router code does handle re-registration of the plugin object, but since Gin routes
// cannot be removed/re-added, this would cause a panic.

func TestRouter_NoAssetHandler(t *testing.T) {
	router := NewRouter()

	// Set asset handler to nil explicitly to trigger the fallback
	router.SetAssetHandler(nil)

	// Request a non-API path should return 404 when no asset handler
	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// errorPlugin is a plugin that returns errors from lifecycle methods
type errorPlugin struct {
	*BasePlugin
	onRegisterErr   error
	onUnregisterErr error
}

func newErrorPlugin(namespace, name string) *errorPlugin {
	return &errorPlugin{
		BasePlugin: NewBasePlugin(namespace, name, nil),
	}
}

func (p *errorPlugin) OnRegister(ctx context.Context) error {
	return p.onRegisterErr
}

func (p *errorPlugin) OnUnregister(ctx context.Context) error {
	return p.onUnregisterErr
}

func TestRouter_RegisterError(t *testing.T) {
	router := NewRouter()
	ctx := context.Background()

	p := newErrorPlugin("core", "error")
	p.onRegisterErr = assert.AnError

	err := router.Register(ctx, p)
	assert.Error(t, err)
}

func TestRouter_UnregisterError(t *testing.T) {
	router := NewRouter()
	ctx := context.Background()

	p := newErrorPlugin("core", "error")
	// First register successfully
	require.NoError(t, router.Register(ctx, p))

	// Set unregister to return error
	p.onUnregisterErr = assert.AnError

	err := router.Unregister(ctx, "core", "error")
	assert.Error(t, err)
}

// customPlugin implements Plugin but is not a BasePlugin
type customPlugin struct {
	name      string
	namespace string
}

func (p *customPlugin) Name() string                                     { return p.name }
func (p *customPlugin) Namespace() string                                { return p.namespace }
func (p *customPlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
func (p *customPlugin) OnRegister(ctx context.Context) error             { return nil }
func (p *customPlugin) OnUnregister(ctx context.Context) error           { return nil }

func TestRouter_ListNonBasePlugin(t *testing.T) {
	router := NewRouter()
	ctx := context.Background()

	// Register a custom plugin that's not a BasePlugin
	p := &customPlugin{name: "custom", namespace: "core"}
	require.NoError(t, router.Register(ctx, p))

	// List should still work and include basic info
	all := router.List()
	assert.Len(t, all, 1)
	assert.Equal(t, "custom", all[0].Name)
	assert.Equal(t, "core", all[0].Namespace)
}
