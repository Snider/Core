package setup

import "github.com/leaanthony/clir"

// AddCommands registers the 'setup' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddSetupCommand(app)
}
