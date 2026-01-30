package dev

import (
	"github.com/spf13/cobra"
)

// addAPICommands adds the 'api' command and its subcommands to the given parent command.
func addAPICommands(parent *cobra.Command) {
	// Create the 'api' command
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Tools for managing service APIs",
	}
	parent.AddCommand(apiCmd)

	// Add the 'sync' command to 'api'
	addSyncCommand(apiCmd)

	// TODO: Add the 'test-gen' command to 'api'
	// addTestGenCommand(apiCmd)
}
