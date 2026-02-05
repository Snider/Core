package mcp

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/host-uk/core/pkg/log"
	"github.com/host-uk/core/pkg/ws"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WSStartInput contains parameters for starting the WebSocket server.
type WSStartInput struct {
	Addr string `json:"addr,omitempty"` // Address to listen on (default: ":8080")
}

// WSStartOutput contains the result of starting the WebSocket server.
type WSStartOutput struct {
	Success bool   `json:"success"`
	Addr    string `json:"addr"`
	Message string `json:"message,omitempty"`
}

// WSInfoInput contains parameters for getting WebSocket hub info.
type WSInfoInput struct{}

// WSInfoOutput contains WebSocket hub statistics.
type WSInfoOutput struct {
	Clients  int `json:"clients"`
	Channels int `json:"channels"`
}

// registerWSTools adds WebSocket tools to the MCP server.
// Returns false if WebSocket hub is not available.
func (s *Service) registerWSTools(server *mcp.Server) bool {
	if s.wsHub == nil {
		return false
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ws_start",
		Description: "Start the WebSocket server for real-time process output streaming.",
	}, s.wsStart)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ws_info",
		Description: "Get WebSocket hub statistics (connected clients and active channels).",
	}, s.wsInfo)

	return true
}

// wsStart handles the ws_start tool call.
func (s *Service) wsStart(ctx context.Context, req *mcp.CallToolRequest, input WSStartInput) (*mcp.CallToolResult, WSStartOutput, error) {
	addr := input.Addr
	if addr == "" {
		addr = ":8080"
	}

	s.logger.Security("MCP tool execution", "tool", "ws_start", "addr", addr, "user", log.Username())

	// Check if server is already running
	if s.wsServer != nil {
		return nil, WSStartOutput{
			Success: true,
			Addr:    s.wsAddr,
			Message: "WebSocket server already running",
		}, nil
	}

	// Create HTTP server with WebSocket handler
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.wsHub.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start listener to get actual address
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("mcp: ws start listen failed", "addr", addr, "err", err)
		return nil, WSStartOutput{}, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	actualAddr := ln.Addr().String()
	s.wsServer = server
	s.wsAddr = actualAddr

	// Start server in background
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Error("mcp: ws server error", "err", err)
		}
	}()

	return nil, WSStartOutput{
		Success: true,
		Addr:    actualAddr,
		Message: fmt.Sprintf("WebSocket server started at ws://%s/ws", actualAddr),
	}, nil
}

// wsInfo handles the ws_info tool call.
func (s *Service) wsInfo(ctx context.Context, req *mcp.CallToolRequest, input WSInfoInput) (*mcp.CallToolResult, WSInfoOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "ws_info", "user", log.Username())

	stats := s.wsHub.Stats()

	return nil, WSInfoOutput{
		Clients:  stats.Clients,
		Channels: stats.Channels,
	}, nil
}

// ProcessEventCallback is a callback function for process events.
// It can be registered with the process service to forward events to WebSocket.
type ProcessEventCallback struct {
	hub *ws.Hub
}

// NewProcessEventCallback creates a callback that forwards process events to WebSocket.
func NewProcessEventCallback(hub *ws.Hub) *ProcessEventCallback {
	return &ProcessEventCallback{hub: hub}
}

// OnProcessOutput forwards process output to WebSocket subscribers.
func (c *ProcessEventCallback) OnProcessOutput(processID string, line string) {
	if c.hub != nil {
		_ = c.hub.SendProcessOutput(processID, line)
	}
}

// OnProcessStatus forwards process status changes to WebSocket subscribers.
func (c *ProcessEventCallback) OnProcessStatus(processID string, status string, exitCode int) {
	if c.hub != nil {
		_ = c.hub.SendProcessStatus(processID, status, exitCode)
	}
}
