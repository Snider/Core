// Package vm provides LinuxKit VM management commands.
package vm

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

func init() {
	cli.RegisterCommands(AddVMCommands)
}

// Style aliases from shared
var (
	repoNameStyle = cli.RepoStyle
	successStyle  = cli.SuccessStyle
	errorStyle    = cli.ErrorStyle
	dimStyle      = cli.DimStyle
)

// VM-specific styles
var (
	varStyle     = cli.NewStyle().Foreground(cli.ColourAmber500)
	defaultStyle = cli.NewStyle().Foreground(cli.ColourGray500).Italic()
)

// AddVMCommands adds container-related commands under 'vm' to the CLI.
func AddVMCommands(root *cobra.Command) {
	vmCmd := &cobra.Command{
		Use:   "vm",
		Short: i18n.T("cmd.vm.short"),
		Long:  i18n.T("cmd.vm.long"),
	}

	root.AddCommand(vmCmd)
	addVMRunCommand(vmCmd)
	addVMPsCommand(vmCmd)
	addVMStopCommand(vmCmd)
	addVMLogsCommand(vmCmd)
	addVMExecCommand(vmCmd)
	addVMTemplatesCommand(vmCmd)
}