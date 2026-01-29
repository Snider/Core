package doctor

import "github.com/leaanthony/clir"

// AddCommands registers the 'doctor' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddDoctorCommand(app)
}
