// Package daemon provides the `core daemon` command for running as a background service.
package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/log"
	"github.com/host-uk/core/pkg/mcp"
)

func init() {
	cli.RegisterCommands(AddDaemonCommand)
}

// Transport types for MCP server.
const (
	TransportStdio  = "stdio"
	TransportTCP    = "tcp"
	TransportSocket = "socket"
)

// Config holds daemon configuration.
type Config struct {
	// MCPTransport is the MCP server transport type (stdio, tcp, socket).
	MCPTransport string
	// MCPAddr is the address/path for tcp or socket transports.
	MCPAddr string
	// HealthAddr is the address for health check endpoints.
	HealthAddr string
	// PIDFile is the path for the PID file.
	PIDFile string
}

// DefaultConfig returns the default daemon configuration.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		MCPTransport: TransportTCP,
		MCPAddr:      mcp.DefaultTCPAddr,
		HealthAddr:   "127.0.0.1:9101",
		PIDFile:      filepath.Join(home, ".core", "daemon.pid"),
	}
}

// ConfigFromEnv loads configuration from environment variables.
// Environment variables override default values.
func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	if v := os.Getenv("CORE_MCP_TRANSPORT"); v != "" {
		cfg.MCPTransport = v
	}
	if v := os.Getenv("CORE_MCP_ADDR"); v != "" {
		cfg.MCPAddr = v
	}
	if v := os.Getenv("CORE_HEALTH_ADDR"); v != "" {
		cfg.HealthAddr = v
	}
	if v := os.Getenv("CORE_PID_FILE"); v != "" {
		cfg.PIDFile = v
	}

	return cfg
}

// AddDaemonCommand adds the 'daemon' command to the root.
func AddDaemonCommand(root *cli.Command) {
	cfg := ConfigFromEnv()

	daemonCmd := cli.NewCommand(
		"daemon",
		"Start the core daemon",
		"Starts the core daemon which provides long-running services like MCP.\n\n"+
			"The daemon can be configured via environment variables or flags:\n"+
			"  CORE_MCP_TRANSPORT - MCP transport type (stdio, tcp, socket)\n"+
			"  CORE_MCP_ADDR      - MCP address/path (e.g., :9100, /tmp/mcp.sock)\n"+
			"  CORE_HEALTH_ADDR   - Health check endpoint address\n"+
			"  CORE_PID_FILE      - PID file path for single-instance enforcement",
		func(cmd *cli.Command, args []string) error {
			return runDaemon(cfg)
		},
	)

	// Flags override environment variables
	cli.StringFlag(daemonCmd, &cfg.MCPTransport, "mcp-transport", "t", cfg.MCPTransport,
		"MCP transport type (stdio, tcp, socket)")
	cli.StringFlag(daemonCmd, &cfg.MCPAddr, "mcp-addr", "a", cfg.MCPAddr,
		"MCP listen address (e.g., :9100 or /tmp/mcp.sock)")
	cli.StringFlag(daemonCmd, &cfg.HealthAddr, "health-addr", "", cfg.HealthAddr,
		"Health check endpoint address (empty to disable)")
	cli.StringFlag(daemonCmd, &cfg.PIDFile, "pid-file", "", cfg.PIDFile,
		"PID file path (empty to disable)")

	root.AddCommand(daemonCmd)
}

// runDaemon starts the daemon with the given configuration.
func runDaemon(cfg Config) error {
	// Set daemon mode environment for child processes
	os.Setenv("CORE_DAEMON", "1")

	log.Info("Starting daemon",
		"transport", cfg.MCPTransport,
		"addr", cfg.MCPAddr,
		"health", cfg.HealthAddr,
	)

	// Create MCP service
	mcpSvc, err := mcp.New()
	if err != nil {
		return fmt.Errorf("failed to create MCP service: %w", err)
	}

	// Create daemon with health checks
	daemon := cli.NewDaemon(cli.DaemonOptions{
		Medium:          io.Local,
		PIDFile:         cfg.PIDFile,
		HealthAddr:      cfg.HealthAddr,
		ShutdownTimeout: 30,
	})

	// Start daemon (acquires PID, starts health server)
	if err := daemon.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Get context that cancels on SIGINT/SIGTERM
	ctx := cli.Context()

	// Start MCP server in background
	mcpErrCh := make(chan error, 1)
	go func() {
		mcpErrCh <- startMCP(ctx, mcpSvc, cfg)
	}()

	// Mark as ready
	daemon.SetReady(true)
	log.Info("Daemon ready",
		"pid", os.Getpid(),
		"health", daemon.HealthAddr(),
	)

	// Wait for shutdown signal or MCP error
	select {
	case err := <-mcpErrCh:
		if err != nil && ctx.Err() == nil {
			log.Error("MCP server error", "err", err)
			return err
		}
	case <-ctx.Done():
		log.Info("Shutting down daemon")
	}

	return daemon.Stop()
}

// startMCP starts the MCP server with the configured transport.
func startMCP(ctx context.Context, svc *mcp.Service, cfg Config) error {
	switch cfg.MCPTransport {
	case TransportStdio:
		log.Info("Starting MCP server", "transport", "stdio")
		return svc.ServeStdio(ctx)

	case TransportTCP:
		log.Info("Starting MCP server", "transport", "tcp", "addr", cfg.MCPAddr)
		return svc.ServeTCP(ctx, cfg.MCPAddr)

	case TransportSocket:
		log.Info("Starting MCP server", "transport", "unix", "path", cfg.MCPAddr)
		return svc.ServeUnix(ctx, cfg.MCPAddr)

	default:
		return fmt.Errorf("unknown MCP transport: %s (valid: stdio, tcp, socket)", cfg.MCPTransport)
	}
}
