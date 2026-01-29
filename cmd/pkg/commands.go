package pkg

import "github.com/leaanthony/clir"

// AddCommands registers the 'pkg' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddPkgCommands(app)
}
