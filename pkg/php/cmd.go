// Package php provides Laravel/PHP development commands.
package php

import (
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
	phpFrankenPHPStyle = cli.NewStyle().Foreground(cli.ColourIndigo500)
	phpViteStyle       = cli.NewStyle().Foreground(cli.ColourYellow500)
	phpHorizonStyle    = cli.NewStyle().Foreground(cli.ColourOrange500)
	phpReverbStyle     = cli.NewStyle().Foreground(cli.ColourViolet500)
	phpRedisStyle      = cli.NewStyle().Foreground(cli.ColourRed500)
)

// Status styles (from shared)
var (
	phpStatusRunning = cli.SuccessStyle
	phpStatusStopped = cli.DimStyle
	phpStatusError   = cli.ErrorStyle
)

// QA command styles (from shared)
var (
	phpQAPassedStyle  = cli.SuccessStyle
	phpQAFailedStyle  = cli.ErrorStyle
	phpQAWarningStyle = cli.WarningStyle
	phpQAStageStyle   = cli.HeaderStyle
)

// Security severity styles (from shared)
var (
	phpSecurityCriticalStyle = cli.NewStyle().Bold().Foreground(cli.ColourRed500)
	phpSecurityHighStyle     = cli.NewStyle().Bold().Foreground(cli.ColourOrange500)
	phpSecurityMediumStyle   = cli.NewStyle().Foreground(cli.ColourAmber500)
	phpSecurityLowStyle      = cli.NewStyle().Foreground(cli.ColourGray500)
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