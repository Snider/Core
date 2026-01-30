// Package vm provides LinuxKit VM management commands.
package vm

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
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
	varStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b"))
	defaultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Italic(true)
)

// AddVMCommands adds container-related commands under 'vm' to the CLI.
func AddVMCommands(root *cobra.Command) {
	vmCmd := &cobra.Command{
		Use:   "vm",
		Short: "LinuxKit VM management",
		Long: "Manage LinuxKit virtual machines.\n\n" +
			"LinuxKit VMs are lightweight, immutable VMs built from YAML templates.\n" +
			"They run using qemu or hyperkit depending on your system.\n\n" +
			"Commands:\n" +
			"  run        Run a VM from image or template\n" +
			"  ps         List running VMs\n" +
			"  stop       Stop a running VM\n" +
			"  logs       View VM logs\n" +
			"  exec       Execute command in VM\n" +
			"  templates  Manage LinuxKit templates",
	}

	root.AddCommand(vmCmd)
	addVMRunCommand(vmCmd)
	addVMPsCommand(vmCmd)
	addVMStopCommand(vmCmd)
	addVMLogsCommand(vmCmd)
	addVMExecCommand(vmCmd)
	addVMTemplatesCommand(vmCmd)
}
