package cmd

import (
	"fmt"
	"os"

	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tviewCmd)
}

var tviewCmd = &cobra.Command{
	Use:   "tview-example",
	Short: "Runs a tview example to demonstrate its capabilities",
	Long:  `This command launches a simple tview application to showcase its full-screen terminal UI features.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Removed: fmt.Println("Starting tview example...") to prevent visual artifacts
		app := tview.NewApplication()
		box := tview.NewBox().
			SetBorder(true).
			SetTitle("Hello from tview!")
		if err := app.SetRoot(box, true).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running tview app: %v\n", err)
			os.Exit(1)
		}
	},
}
