package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "demo-cli",
	Short: "A demo CLI for the display module",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from the demo CLI!")
	},
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once
// to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
