package main

import (
	"context"

	"github.com/Snider/Core/pkg/module"
	"github.com/Snider/Mining/pkg/mining"
	"github.com/gin-gonic/gin"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MiningBridge wraps the Mining service for integration with lthn-desktop.
// It implements module.GinModule to register routes with the module registry.
type MiningBridge struct {
	manager *mining.Manager
	service *mining.Service
}

// NewMiningBridge creates a new mining bridge with its own manager.
func NewMiningBridge() *MiningBridge {
	return &MiningBridge{}
}

// ServiceName returns the canonical service name for Wails.
func (m *MiningBridge) ServiceName() string {
	return "lthn-desktop/mining"
}

// ServiceStartup initializes the mining manager and service.
func (m *MiningBridge) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	// Create the mining manager
	m.manager = mining.NewManager()

	// Create the service - empty addresses since we're embedding, not serving standalone
	service, err := mining.NewService(m.manager, "", "", "/api/mining")
	if err != nil {
		return err
	}
	m.service = service

	// Start the service which initializes the router and sets up routes
	// We use a background context since we don't want it to be canceled with the startup context
	go func() {
		// Note: ServiceStartup runs the HTTP server which we don't need when embedded
		// The routes are already registered on m.service.Router
		_ = m.service.ServiceStartup(context.Background())
	}()

	return nil
}

// ServiceShutdown cleans up the mining manager.
func (m *MiningBridge) ServiceShutdown() error {
	if m.manager != nil {
		m.manager.Stop()
	}
	return nil
}

// RegisterRoutes implements module.GinModule to register mining API routes.
func (m *MiningBridge) RegisterRoutes(group *gin.RouterGroup) {
	// The mining service router handles all /api/mining/* routes
	// Forward requests from the module registry's group to the mining router
	if m.service != nil && m.service.Router != nil {
		group.Any("/*path", func(c *gin.Context) {
			// Rewrite the path to match what the mining router expects
			originalPath := c.Request.URL.Path
			c.Request.URL.Path = "/api/mining" + c.Param("path")
			m.service.Router.ServeHTTP(c.Writer, c.Request)
			c.Request.URL.Path = originalPath
		})
	}
}

// ModuleConfig returns the module configuration for registration.
func (m *MiningBridge) ModuleConfig() module.Config {
	return module.Config{
		Code:        "mining",
		Type:        module.TypeCore,
		Name:        "Mining Module",
		Version:     "0.1.0",
		Namespace:   "mining",
		Description: "Cryptocurrency mining management powered by Snider/Mining",
		Author:      "Lethean",
		Contexts:    []module.Context{module.ContextMiner, module.ContextDefault},
		Menu: []module.MenuItem{
			{
				ID:       "mining",
				Label:    "Mining",
				Order:    200,
				Contexts: []module.Context{module.ContextMiner},
				Children: []module.MenuItem{
					{ID: "mining-dashboard", Label: "Dashboard", Route: "/mining/dashboard", Order: 1},
					{ID: "mining-pools", Label: "Pools", Route: "/mining/pools", Order: 2},
					{ID: "mining-miners", Label: "Miners", Route: "/mining/miners", Order: 3},
					{ID: "mining-sep1", Separator: true, Order: 4},
					{ID: "mining-start", Label: "Start Mining", Action: "mining:start", Order: 5},
					{ID: "mining-stop", Label: "Stop Mining", Action: "mining:stop", Order: 6},
				},
			},
		},
		Routes: []module.Route{
			{Path: "/mining/dashboard", Component: "mining-dashboard", Title: "Mining Dashboard", Contexts: []module.Context{module.ContextMiner}},
			{Path: "/mining/pools", Component: "mining-pools", Title: "Mining Pools", Contexts: []module.Context{module.ContextMiner}},
			{Path: "/mining/miners", Component: "mining-miners", Title: "Miners", Contexts: []module.Context{module.ContextMiner}},
		},
		API: []module.APIEndpoint{
			{Method: "GET", Path: "/health", Description: "Health check"},
			{Method: "GET", Path: "/ready", Description: "Readiness check"},
			{Method: "GET", Path: "/info", Description: "Get mining service info"},
			{Method: "GET", Path: "/metrics", Description: "Get mining metrics"},
			{Method: "GET", Path: "/miners", Description: "List running miners"},
			{Method: "GET", Path: "/miners/available", Description: "List available miners"},
			{Method: "POST", Path: "/miners/:miner_name/install", Description: "Install a miner"},
			{Method: "DELETE", Path: "/miners/:miner_name/uninstall", Description: "Uninstall a miner"},
			{Method: "DELETE", Path: "/miners/:miner_name", Description: "Stop a miner"},
			{Method: "GET", Path: "/miners/:miner_name/stats", Description: "Get miner stats"},
			{Method: "GET", Path: "/miners/:miner_name/logs", Description: "Get miner logs"},
			{Method: "POST", Path: "/profiles", Description: "Create mining profile"},
			{Method: "GET", Path: "/profiles", Description: "List mining profiles"},
			{Method: "POST", Path: "/profiles/:id/start", Description: "Start mining with profile"},
		},
	}
}

// --- Wails-bound methods for frontend access ---

// MinerInfo represents basic miner information for the frontend.
type MinerInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsRunning bool   `json:"isRunning"`
}

// MinerStats represents miner statistics for the frontend.
type MinerStats struct {
	Hashrate  int    `json:"hashrate"`
	Shares    int    `json:"shares"`
	Rejected  int    `json:"rejected"`
	Uptime    int    `json:"uptime"`
	LastShare int64  `json:"lastShare"`
	Algorithm string `json:"algorithm"`
}

// AvailableMiner represents an available miner for installation.
type AvailableMiner struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetManager returns the mining manager for direct access.
func (m *MiningBridge) GetManager() *mining.Manager {
	return m.manager
}

// ListMiners returns a list of running miners.
func (m *MiningBridge) ListMiners() []MinerInfo {
	if m.manager == nil {
		return nil
	}
	miners := m.manager.ListMiners()
	result := make([]MinerInfo, len(miners))
	for i, miner := range miners {
		result[i] = MinerInfo{
			Name:      miner.GetName(),
			Type:      miner.GetName(), // Type is typically the same as name for now
			IsRunning: true,            // If it's in ListMiners, it's running
		}
	}
	return result
}

// GetMinerStats returns stats for a specific miner.
func (m *MiningBridge) GetMinerStats(name string) (*MinerStats, error) {
	if m.manager == nil {
		return nil, nil
	}
	miner, err := m.manager.GetMiner(name)
	if err != nil {
		return nil, err
	}
	stats, err := miner.GetStats()
	if err != nil {
		return nil, err
	}
	return &MinerStats{
		Hashrate:  stats.Hashrate,
		Shares:    stats.Shares,
		Rejected:  stats.Rejected,
		Uptime:    stats.Uptime,
		LastShare: stats.LastShare,
		Algorithm: stats.Algorithm,
	}, nil
}

// StartMiner starts a miner with the given configuration.
func (m *MiningBridge) StartMiner(minerType string, pool string, wallet string) error {
	if m.manager == nil {
		return nil
	}
	config := &mining.Config{
		Pool:      pool,
		Wallet:    wallet,
		LogOutput: true,
	}
	_, err := m.manager.StartMiner(minerType, config)
	return err
}

// StopMiner stops a running miner.
func (m *MiningBridge) StopMiner(name string) error {
	if m.manager == nil {
		return nil
	}
	return m.manager.StopMiner(name)
}

// GetAvailableMiners returns available miner types with their installation status.
func (m *MiningBridge) GetAvailableMiners() []AvailableMiner {
	if m.manager == nil {
		return nil
	}
	available := m.manager.ListAvailableMiners()
	result := make([]AvailableMiner, len(available))
	for i, am := range available {
		result[i] = AvailableMiner{
			Name:        am.Name,
			Description: am.Description,
		}
	}
	return result
}

// InstallMiner installs a miner by type.
func (m *MiningBridge) InstallMiner(minerType string) error {
	if m.manager == nil {
		return nil
	}
	// Start the miner with no config to trigger installation
	// The Mining package handles installation when starting an uninstalled miner
	miner, err := m.manager.GetMiner(minerType)
	if err != nil {
		// Miner not found, create a temporary config to start (which will install)
		config := &mining.Config{
			Pool:   "stratum+tcp://placeholder:3333",
			Wallet: "placeholder",
		}
		_, err = m.manager.StartMiner(minerType, config)
		if err != nil {
			return err
		}
		// Stop it after installation
		return m.manager.StopMiner(minerType)
	}
	// Miner exists, install it directly
	return miner.Install()
}

// UninstallMiner uninstalls a miner by type.
func (m *MiningBridge) UninstallMiner(minerType string) error {
	if m.manager == nil {
		return nil
	}
	return m.manager.UninstallMiner(minerType)
}

// GetHashrateHistory returns hashrate history for a miner.
func (m *MiningBridge) GetHashrateHistory(name string) ([]mining.HashratePoint, error) {
	if m.manager == nil {
		return nil, nil
	}
	return m.manager.GetMinerHashrateHistory(name)
}
