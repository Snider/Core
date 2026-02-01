package workspace

import "github.com/host-uk/core/pkg/cli"

func init() {
	cli.RegisterCommands(AddWorkspaceCommands)
}
