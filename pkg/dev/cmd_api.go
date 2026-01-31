package dev

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

// addAPICommands adds the 'api' command and its subcommands to the given parent command.
func addAPICommands(parent *cli.Command) {
	// Create the 'api' command
	apiCmd := &cli.Command{
		Use:   "api",
		Short: i18n.T("cmd.dev.api.short"),
	}
	parent.AddCommand(apiCmd)

	// Add the 'sync' command to 'api'
	addSyncCommand(apiCmd)

	// TODO: Add the 'test-gen' command to 'api'
	// addTestGenCommand(apiCmd)
}
