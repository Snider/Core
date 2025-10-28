package cmd

import (
	"github.com/leaanthony/clir"
)

// AddAPICommands adds the 'api' command and its subcommands to the given parent command.
func AddAPICommands(parent *clir.Command) {
	// Create the 'api' command
	apiCmd := parent.NewSubCommand("api", "Tools for managing service APIs")

	// Add the 'sync' command to 'api'
	AddSyncCommand(apiCmd)
}
