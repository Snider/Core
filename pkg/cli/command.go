package cli

import (
	"github.com/spf13/cobra"
)

// ─────────────────────────────────────────────────────────────────────────────
// Command Type Re-export
// ─────────────────────────────────────────────────────────────────────────────

// Command is the cobra command type.
// Re-exported for convenience so packages don't need to import cobra directly.
type Command = cobra.Command

// ─────────────────────────────────────────────────────────────────────────────
// Command Builders
// ─────────────────────────────────────────────────────────────────────────────

// NewCommand creates a new command with a RunE handler.
// This is the standard way to create commands that may return errors.
//
//	cmd := cli.NewCommand("build", "Build the project", "", func(cmd *cli.Command, args []string) error {
//	    // Build logic
//	    return nil
//	})
func NewCommand(use, short, long string, run func(cmd *Command, args []string) error) *Command {
	cmd := &Command{
		Use:   use,
		Short: short,
		RunE:  run,
	}
	if long != "" {
		cmd.Long = long
	}
	return cmd
}

// NewGroup creates a new command group (no RunE).
// Use this for parent commands that only contain subcommands.
//
//	devCmd := cli.NewGroup("dev", "Development commands", "")
//	devCmd.AddCommand(buildCmd, testCmd)
func NewGroup(use, short, long string) *Command {
	cmd := &Command{
		Use:   use,
		Short: short,
	}
	if long != "" {
		cmd.Long = long
	}
	return cmd
}

// NewRun creates a new command with a simple Run handler (no error return).
// Use when the command cannot fail.
//
//	cmd := cli.NewRun("version", "Show version", "", func(cmd *cli.Command, args []string) {
//	    cli.Println("v1.0.0")
//	})
func NewRun(use, short, long string, run func(cmd *Command, args []string)) *Command {
	cmd := &Command{
		Use:   use,
		Short: short,
		Run:   run,
	}
	if long != "" {
		cmd.Long = long
	}
	return cmd
}

// ─────────────────────────────────────────────────────────────────────────────
// Flag Helpers
// ─────────────────────────────────────────────────────────────────────────────

// StringFlag adds a string flag to a command.
// The value will be stored in the provided pointer.
//
//	var output string
//	cli.StringFlag(cmd, &output, "output", "o", "", "Output file path")
func StringFlag(cmd *Command, ptr *string, name, short, def, usage string) {
	if short != "" {
		cmd.Flags().StringVarP(ptr, name, short, def, usage)
	} else {
		cmd.Flags().StringVar(ptr, name, def, usage)
	}
}

// BoolFlag adds a boolean flag to a command.
// The value will be stored in the provided pointer.
//
//	var verbose bool
//	cli.BoolFlag(cmd, &verbose, "verbose", "v", false, "Enable verbose output")
func BoolFlag(cmd *Command, ptr *bool, name, short string, def bool, usage string) {
	if short != "" {
		cmd.Flags().BoolVarP(ptr, name, short, def, usage)
	} else {
		cmd.Flags().BoolVar(ptr, name, def, usage)
	}
}

// IntFlag adds an integer flag to a command.
// The value will be stored in the provided pointer.
//
//	var count int
//	cli.IntFlag(cmd, &count, "count", "n", 10, "Number of items")
func IntFlag(cmd *Command, ptr *int, name, short string, def int, usage string) {
	if short != "" {
		cmd.Flags().IntVarP(ptr, name, short, def, usage)
	} else {
		cmd.Flags().IntVar(ptr, name, def, usage)
	}
}

// StringSliceFlag adds a string slice flag to a command.
// The value will be stored in the provided pointer.
//
//	var tags []string
//	cli.StringSliceFlag(cmd, &tags, "tag", "t", nil, "Tags to apply")
func StringSliceFlag(cmd *Command, ptr *[]string, name, short string, def []string, usage string) {
	if short != "" {
		cmd.Flags().StringSliceVarP(ptr, name, short, def, usage)
	} else {
		cmd.Flags().StringSliceVar(ptr, name, def, usage)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Persistent Flag Helpers
// ─────────────────────────────────────────────────────────────────────────────

// PersistentStringFlag adds a persistent string flag (inherited by subcommands).
func PersistentStringFlag(cmd *Command, ptr *string, name, short, def, usage string) {
	if short != "" {
		cmd.PersistentFlags().StringVarP(ptr, name, short, def, usage)
	} else {
		cmd.PersistentFlags().StringVar(ptr, name, def, usage)
	}
}

// PersistentBoolFlag adds a persistent boolean flag (inherited by subcommands).
func PersistentBoolFlag(cmd *Command, ptr *bool, name, short string, def bool, usage string) {
	if short != "" {
		cmd.PersistentFlags().BoolVarP(ptr, name, short, def, usage)
	} else {
		cmd.PersistentFlags().BoolVar(ptr, name, def, usage)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Command Configuration
// ─────────────────────────────────────────────────────────────────────────────

// WithArgs sets the Args validation function for a command.
// Returns the command for chaining.
//
//	cmd := cli.NewCommand("build", "Build", "", run).WithArgs(cobra.ExactArgs(1))
func WithArgs(cmd *Command, args cobra.PositionalArgs) *Command {
	cmd.Args = args
	return cmd
}

// WithExample sets the Example field for a command.
// Returns the command for chaining.
func WithExample(cmd *Command, example string) *Command {
	cmd.Example = example
	return cmd
}

// ExactArgs returns a PositionalArgs that accepts exactly N arguments.
func ExactArgs(n int) cobra.PositionalArgs {
	return cobra.ExactArgs(n)
}

// MinimumNArgs returns a PositionalArgs that accepts minimum N arguments.
func MinimumNArgs(n int) cobra.PositionalArgs {
	return cobra.MinimumNArgs(n)
}

// MaximumNArgs returns a PositionalArgs that accepts maximum N arguments.
func MaximumNArgs(n int) cobra.PositionalArgs {
	return cobra.MaximumNArgs(n)
}

// NoArgs returns a PositionalArgs that accepts no arguments.
func NoArgs() cobra.PositionalArgs {
	return cobra.NoArgs
}

// ArbitraryArgs returns a PositionalArgs that accepts any arguments.
func ArbitraryArgs() cobra.PositionalArgs {
	return cobra.ArbitraryArgs
}
