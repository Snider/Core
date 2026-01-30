// Package cli provides the CLI runtime and utilities.
package cli

import (
	"sync"

	"github.com/spf13/cobra"
)

// CommandRegistration is a function that adds commands to the root.
type CommandRegistration func(root *cobra.Command)

var (
	registeredCommands   []CommandRegistration
	registeredCommandsMu sync.Mutex
	commandsAttached     bool
)

// RegisterCommands registers a function that adds commands to the CLI.
// Call this in your package's init() to register commands.
//
//	func init() {
//	    cli.RegisterCommands(AddCommands)
//	}
//
//	func AddCommands(root *cobra.Command) {
//	    root.AddCommand(myCmd)
//	}
func RegisterCommands(fn CommandRegistration) {
	registeredCommandsMu.Lock()
	defer registeredCommandsMu.Unlock()
	registeredCommands = append(registeredCommands, fn)

	// If commands already attached (CLI already running), attach immediately
	if commandsAttached && instance != nil && instance.root != nil {
		fn(instance.root)
	}
}

// attachRegisteredCommands calls all registered command functions.
// Called by Init() after creating the root command.
func attachRegisteredCommands(root *cobra.Command) {
	registeredCommandsMu.Lock()
	defer registeredCommandsMu.Unlock()

	for _, fn := range registeredCommands {
		fn(root)
	}
	commandsAttached = true
}
