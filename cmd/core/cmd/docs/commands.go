package docs

import "github.com/leaanthony/clir"

// AddCommands registers the 'docs' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddDocsCommand(app)
}
