package testcmd

import "github.com/leaanthony/clir"

// AddCommands registers the 'test' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddTestCommand(app)
}
