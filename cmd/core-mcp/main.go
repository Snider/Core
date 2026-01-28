// core-mcp is a standalone MCP server for Core.
// It allows Claude Code and other MCP clients to interact with Core's
// file system and IDE functionality.
//
// Usage:
//
//	Add to Claude Code's MCP settings:
//	{
//	  "mcpServers": {
//	    "core": {
//	      "command": "/path/to/core-mcp"
//	    }
//	  }
//	}
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/host-uk/core/pkg/mcp"
)

func main() {
	// Create standalone MCP service (no Core instance needed)
	svc := mcp.NewStandalone()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Run the MCP server on stdio
	if err := svc.Run(ctx); err != nil {
		log.Printf("MCP server error: %v", err)
		os.Exit(1)
	}
}
