// Package vm provides LinuxKit VM management commands.
package vm

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// Style aliases from shared
var (
	repoNameStyle = shared.RepoNameStyle
	successStyle  = shared.SuccessStyle
	errorStyle    = shared.ErrorStyle
	dimStyle      = shared.DimStyle
)

// VM-specific styles
var (
	varStyle     = lipgloss.NewStyle().Foreground(shared.ColourAmber500)
	defaultStyle = lipgloss.NewStyle().Foreground(shared.ColourGray500).Italic(true)
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
