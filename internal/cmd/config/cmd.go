package config

import "github.com/host-uk/core/pkg/cli"

func init() {
	cli.RegisterCommands(AddConfigCommands)
}

// AddConfigCommands registers the 'config' command group and all subcommands.
func AddConfigCommands(root *cli.Command) {
	configCmd := cli.NewGroup("config", "Manage configuration", "")
	root.AddCommand(configCmd)

	addGetCommand(configCmd)
	addSetCommand(configCmd)
	addListCommand(configCmd)
	addPathCommand(configCmd)
}
