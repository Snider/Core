package deploy

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/spf13/cobra"
)

func init() {
	cli.RegisterCommands(AddDeployCommands)
}

// AddDeployCommands registers the 'deploy' command and all subcommands.
func AddDeployCommands(root *cobra.Command) {
	root.AddCommand(Cmd)
}
