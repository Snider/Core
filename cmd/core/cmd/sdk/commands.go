package sdk

import "github.com/leaanthony/clir"

// AddCommands registers the 'sdk' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddSDKCommand(app)
}
