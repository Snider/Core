// Package vm provides LinuxKit VM management commands.
package vm

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

func init() {
	cli.RegisterCommands(AddVMCommands)
}

// Style aliases from shared
var (
	repoNameStyle = cli.RepoNameStyle
	successStyle  = cli.SuccessStyle
	errorStyle    = cli.ErrorStyle
	dimStyle      = cli.DimStyle
)

// VM-specific styles
var (
	varStyle     = lipgloss.NewStyle().Foreground(cli.ColourAmber500)
	defaultStyle = lipgloss.NewStyle().Foreground(cli.ColourGray500).Italic(true)
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
