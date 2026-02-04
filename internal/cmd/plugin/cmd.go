// Package plugin provides CLI commands for managing core plugins.
//
// Commands:
//   - install: Install a plugin from GitHub
//   - list: List installed plugins
//   - info: Show detailed plugin information
//   - update: Update a plugin or all plugins
//   - remove: Remove an installed plugin
package plugin

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

func init() {
	cli.RegisterCommands(AddPluginCommands)
}

// AddPluginCommands registers the 'plugin' command and all subcommands.
func AddPluginCommands(root *cli.Command) {
	pluginCmd := &cli.Command{
		Use:   "plugin",
		Short: i18n.T("Manage plugins"),
	}
	root.AddCommand(pluginCmd)

	addInstallCommand(pluginCmd)
	addListCommand(pluginCmd)
	addInfoCommand(pluginCmd)
	addUpdateCommand(pluginCmd)
	addRemoveCommand(pluginCmd)
}
