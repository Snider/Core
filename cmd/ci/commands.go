package ci

import "github.com/leaanthony/clir"

// AddCommands registers the 'ci' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddCIReleaseCommand(app)
}
