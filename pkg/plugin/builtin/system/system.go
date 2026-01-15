// Package system provides the built-in system plugin for Core.
// It exposes runtime information and basic system operations via HTTP API.
package system

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Snider/Core/pkg/plugin"
)

// Plugin is the built-in system information plugin.
type Plugin struct {
	*plugin.BasePlugin
	startTime time.Time
}

// New creates a new system plugin.
func New() *Plugin {
	p := &Plugin{
		startTime: time.Now(),
	}
	p.BasePlugin = plugin.NewBasePlugin("core", "system", nil).
		WithDescription("Core system information and operations").
		WithVersion("1.0.0")
	return p
}

// RegisterRoutes registers the plugin's Gin routes.
func (p *Plugin) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/info", p.handleInfo)
	group.GET("/health", p.handleHealth)
	group.GET("/runtime", p.handleRuntime)
}

// InfoResponse contains system information.
type InfoResponse struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	GoVersion string `json:"goVersion"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

func (p *Plugin) handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, InfoResponse{
		Name:      "Core",
		Version:   "0.1.0",
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	})
}

// HealthResponse contains health check information.
type HealthResponse struct {
	Status string `json:"status"`
	Uptime string `json:"uptime"`
}

func (p *Plugin) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "healthy",
		Uptime: time.Since(p.startTime).String(),
	})
}

// RuntimeResponse contains Go runtime statistics.
type RuntimeResponse struct {
	NumGoroutine int    `json:"numGoroutine"`
	NumCPU       int    `json:"numCPU"`
	MemAlloc     uint64 `json:"memAlloc"`
	MemTotal     uint64 `json:"memTotalAlloc"`
	MemSys       uint64 `json:"memSys"`
	NumGC        uint32 `json:"numGC"`
}

func (p *Plugin) handleRuntime(c *gin.Context) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	c.JSON(http.StatusOK, RuntimeResponse{
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		MemAlloc:     mem.Alloc,
		MemTotal:     mem.TotalAlloc,
		MemSys:       mem.Sys,
		NumGC:        mem.NumGC,
	})
}
