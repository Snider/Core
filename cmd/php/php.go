// Package php provides Laravel/PHP development commands.
package php

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
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
	phpViteStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#eab308")) // yellow-500
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
	phpQAStageStyle   = lipgloss.NewStyle().Bold(true).Foreground(shared.ColourIndigo500)
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
		Short: "Laravel/PHP development tools",
		Long: "Manage Laravel development environment with FrankenPHP.\n\n" +
			"Services orchestrated:\n" +
			"  - FrankenPHP/Octane (port 8000, HTTPS on 443)\n" +
			"  - Vite dev server (port 5173)\n" +
			"  - Laravel Horizon (queue workers)\n" +
			"  - Laravel Reverb (WebSocket, port 8080)\n" +
			"  - Redis (port 6379)",
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
