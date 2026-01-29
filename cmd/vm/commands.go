package vm

import "github.com/leaanthony/clir"

// AddCommands registers the 'vm' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddContainerCommands(app)
}
