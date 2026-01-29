package gocmd

import "github.com/leaanthony/clir"

// AddCommands registers the 'go' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddGoCommands(app)
}
