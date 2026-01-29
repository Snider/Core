package php

import "github.com/leaanthony/clir"

// AddCommands registers the 'php' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddPHPCommands(app)
}
