package workspace

import (
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/spf13/cobra"
)

// AddWorkspaceCommands registers workspace management commands.
func AddWorkspaceCommands(root *cobra.Command) {
	wsCmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage workspace configuration",
		RunE:  runWorkspaceInfo,
	}

	wsCmd.AddCommand(&cobra.Command{
		Use:   "active [package]",
		Short: "Show or set the active package",
		RunE:  runWorkspaceActive,
	})

	addTaskCommands(wsCmd)

	root.AddCommand(wsCmd)
}

func runWorkspaceInfo(cmd *cobra.Command, args []string) error {
	root, err := FindWorkspaceRoot()
	if err != nil {
		return cli.Err("not in a workspace")
	}

	config, err := LoadConfig(root)
	if err != nil {
		return err
	}
	if config == nil {
		return cli.Err("workspace config not found")
	}

	cli.Print("Active:   %s\n", cli.ValueStyle.Render(config.Active))
	cli.Print("Packages: %s\n", cli.DimStyle.Render(config.PackagesDir))
	if len(config.DefaultOnly) > 0 {
		cli.Print("Types:    %s\n", cli.DimStyle.Render(strings.Join(config.DefaultOnly, ", ")))
	}

	return nil
}

func runWorkspaceActive(cmd *cobra.Command, args []string) error {
	root, err := FindWorkspaceRoot()
	if err != nil {
		return cli.Err("not in a workspace")
	}

	config, err := LoadConfig(root)
	if err != nil {
		return err
	}
	if config == nil {
		config = DefaultConfig()
	}

	// If no args, show active
	if len(args) == 0 {
		if config.Active == "" {
			cli.Println("No active package set")
			return nil
		}
		cli.Text(config.Active)
		return nil
	}

	// Set active
	target := args[0]
	if target == config.Active {
		cli.Print("Active package is already %s\n", cli.ValueStyle.Render(target))
		return nil
	}

	config.Active = target
	if err := SaveConfig(root, config); err != nil {
		return err
	}

	cli.Print("Active package set to %s\n", cli.SuccessStyle.Render(target))
	return nil
}
