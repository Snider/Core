package daemon

import (
	"github.com/host-uk/core/pkg/cli"
)

func init() {
	cli.RegisterCommands(AddDaemonCommand)
}

// AddDaemonCommand adds the 'daemon' command to the root.
func AddDaemonCommand(root *cli.Command) {
	var mcpTransport string
	var mcpAddr string

	daemonCmd := cli.NewCommand("daemon", "Start the core daemon", "Starts the core daemon which provides long-running services like MCP.", func(cmd *cli.Command, args []string) error {
		cli.LogInfo("Daemon started")
		// cli.Run blocks until context is cancelled (SIGINT/SIGTERM)
		return cli.Run(cli.Context())
	})

	// Flags for MCP configuration (also parsed manually in cli.Main)
	cli.StringFlag(daemonCmd, &mcpTransport, "mcp-transport", "", "", "MCP transport type (stdio, tcp, socket)")
	cli.StringFlag(daemonCmd, &mcpAddr, "mcp-addr", "", "", "MCP listen address (e.g. :9100 or /tmp/mcp.sock)")

	root.AddCommand(daemonCmd)
}
