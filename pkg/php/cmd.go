// Package php provides Laravel/PHP development commands.
package php

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

func init() {
	cli.RegisterCommands(AddPHPCommands)
}

// Style aliases from shared
var (
	successStyle = cli.SuccessStyle
	errorStyle   = cli.ErrorStyle
	dimStyle     = cli.DimStyle
	linkStyle    = cli.LinkStyle
)

// Service colors for log output (domain-specific, keep local)
var (
	phpFrankenPHPStyle = lipgloss.NewStyle().Foreground(cli.ColourIndigo500)
	phpViteStyle       = lipgloss.NewStyle().Foreground(cli.ColourYellow500)
	phpHorizonStyle    = lipgloss.NewStyle().Foreground(cli.ColourOrange500)
	phpReverbStyle     = lipgloss.NewStyle().Foreground(cli.ColourViolet500)
	phpRedisStyle      = lipgloss.NewStyle().Foreground(cli.ColourRed500)
)

// Status styles (from shared)
var (
	phpStatusRunning = cli.SuccessStyle
	phpStatusStopped = cli.StatusPendingStyle
	phpStatusError   = cli.ErrorStyle
)

// QA command styles (from shared)
var (
	phpQAPassedStyle  = cli.SuccessStyle
	phpQAFailedStyle  = cli.ErrorStyle
	phpQAWarningStyle = cli.WarningStyle
	phpQAStageStyle   = cli.StageStyle
)

// Security severity styles (from shared)
var (
	phpSecurityCriticalStyle = cli.SeverityCriticalStyle
	phpSecurityHighStyle     = cli.SeverityHighStyle
	phpSecurityMediumStyle   = cli.SeverityMediumStyle
	phpSecurityLowStyle      = cli.SeverityLowStyle
)

// AddPHPCommands adds PHP/Laravel development commands.
func AddPHPCommands(root *cobra.Command) {
	phpCmd := &cobra.Command{
		Use:   "php",
		Short: i18n.T("cmd.php.short"),
		Long:  i18n.T("cmd.php.long"),
	}
	root.AddCommand(phpCmd)

	// Development
	addPHPDevCommand(phpCmd)
	addPHPLogsCommand(phpCmd)
	addPHPStopCommand(phpCmd)
	addPHPStatusCommand(phpCmd)
	addPHPSSLCommand(phpCmd)

	// Build & Deploy
	addPHPBuildCommand(phpCmd)
	addPHPServeCommand(phpCmd)
	addPHPShellCommand(phpCmd)

	// Quality (existing)
	addPHPTestCommand(phpCmd)
	addPHPFmtCommand(phpCmd)
	addPHPStanCommand(phpCmd)

	// Quality (new)
	addPHPPsalmCommand(phpCmd)
	addPHPAuditCommand(phpCmd)
	addPHPSecurityCommand(phpCmd)
	addPHPQACommand(phpCmd)
	addPHPRectorCommand(phpCmd)
	addPHPInfectionCommand(phpCmd)

	// Package Management
	addPHPPackagesCommands(phpCmd)

	// Deployment
	addPHPDeployCommands(phpCmd)
}
