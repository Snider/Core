package dev

import (
	"github.com/leaanthony/clir"
)

// addAPICommands adds the 'api' command and its subcommands to the given parent command.
func addAPICommands(parent *clir.Command) {
	// Create the 'api' command
	apiCmd := parent.NewSubCommand("api", "Tools for managing service APIs")

	// Add the 'sync' command to 'api'
	addSyncCommand(apiCmd)

	// TODO: Add the 'test-gen' command to 'api'
	// addTestGenCommand(apiCmd)
}
