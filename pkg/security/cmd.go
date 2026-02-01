package security

import "github.com/host-uk/core/pkg/cli"

func init() {
	cli.RegisterCommands(AddSecurityCommands)
}
