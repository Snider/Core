// Package mcpcmd provides the MCP server command.
//
// Commands:
//   - mcp serve: Start the MCP server for AI tool integration
package mcpcmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/mcp"
)

func init() {
	cli.RegisterCommands(AddMCPCommands)
}

var workspaceFlag string

var mcpCmd = &cli.Command{
	Use:   "mcp",
	Short: "MCP server for AI tool integration",
	Long:  "Model Context Protocol (MCP) server providing file operations, RAG, and metrics tools.",
}

var serveCmd = &cli.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `Start the MCP server on stdio (default) or TCP.

The server provides file operations, RAG tools, and metrics tools for AI assistants.

Environment variables:
  MCP_ADDR    TCP address to listen on (e.g., "localhost:9999")
              If not set, uses stdio transport.

Examples:
  # Start with stdio transport (for Claude Code integration)
  core mcp serve

  # Start with workspace restriction
  core mcp serve --workspace /path/to/project

  # Start TCP server
  MCP_ADDR=localhost:9999 core mcp serve`,
	RunE: func(cmd *cli.Command, args []string) error {
		return runServe()
	},
}

func initFlags() {
	cli.StringFlag(serveCmd, &workspaceFlag, "workspace", "w", "", "Restrict file operations to this directory (empty = unrestricted)")
}

// AddMCPCommands registers the 'mcp' command and all subcommands.
func AddMCPCommands(root *cli.Command) {
	initFlags()
	mcpCmd.AddCommand(serveCmd)
	root.AddCommand(mcpCmd)
}

func runServe() error {
	// Build MCP service options
	var opts []mcp.Option

	if workspaceFlag != "" {
		opts = append(opts, mcp.WithWorkspaceRoot(workspaceFlag))
	} else {
		// Explicitly unrestricted when no workspace specified
		opts = append(opts, mcp.WithWorkspaceRoot(""))
	}

	// Create the MCP service
	svc, err := mcp.New(opts...)
	if err != nil {
		return cli.Wrap(err, "create MCP service")
	}

	// Set up signal handling for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	// Run the server (blocks until context cancelled or error)
	return svc.Run(ctx)
}
