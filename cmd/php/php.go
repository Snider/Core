// Package php provides Laravel/PHP development commands.
package php

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// Style aliases from shared
var (
	successStyle = shared.SuccessStyle
	errorStyle   = shared.ErrorStyle
	dimStyle     = shared.DimStyle
	linkStyle    = shared.LinkStyle
)

// Service colors for log output (domain-specific, keep local)
var (
	phpFrankenPHPStyle = lipgloss.NewStyle().Foreground(shared.ColourIndigo500)
	phpViteStyle       = lipgloss.NewStyle().Foreground(shared.ColourYellow500)
	phpHorizonStyle    = lipgloss.NewStyle().Foreground(shared.ColourOrange500)
	phpReverbStyle     = lipgloss.NewStyle().Foreground(shared.ColourViolet500)
	phpRedisStyle      = lipgloss.NewStyle().Foreground(shared.ColourRed500)
)

// Status styles (from shared)
var (
	phpStatusRunning = shared.SuccessStyle
	phpStatusStopped = shared.StatusPendingStyle
	phpStatusError   = shared.ErrorStyle
)

// QA command styles (from shared)
var (
	phpQAPassedStyle  = shared.SuccessStyle
	phpQAFailedStyle  = shared.ErrorStyle
	phpQAWarningStyle = shared.WarningStyle
	phpQAStageStyle   = shared.StageStyle
)

// Security severity styles (from shared)
var (
	phpSecurityCriticalStyle = shared.SeverityCriticalStyle
	phpSecurityHighStyle     = shared.SeverityHighStyle
	phpSecurityMediumStyle   = shared.SeverityMediumStyle
	phpSecurityLowStyle      = shared.SeverityLowStyle
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
	addPHPAnalyseCommand(phpCmd)

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
