package cmd

import (
	"fmt"
	"os"

	"github.com/leaanthony/clir"
	"github.com/rivo/tview"
)

// AddTviewCommand adds the tview-example command to the clir app.
func AddTviewCommand(app *clir.Cli) {
	tviewCmd := app.NewSubCommand("tview-example", "Runs a tview example to demonstrate its capabilities")
	tviewCmd.LongDescription("This command launches a simple tview application to showcase its full-screen terminal UI features.")
	tviewCmd.Action(func() error {
		app := tview.NewApplication()
		box := tview.NewBox().
			SetBorder(true).
			SetTitle("Hello from tview!")
		if err := app.SetRoot(box, true).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running tview app: %v\n", err)
			os.Exit(1)
		}
		return nil
	})
}
