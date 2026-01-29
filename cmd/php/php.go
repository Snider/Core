// Package php provides Laravel/PHP development commands.
package php

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
	"github.com/leaanthony/clir"
)

// Style aliases from shared
var (
	successStyle = shared.SuccessStyle
	errorStyle   = shared.ErrorStyle
	dimStyle     = shared.DimStyle
	linkStyle    = shared.LinkStyle
)

// Service colors for log output
var (
	phpFrankenPHPStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6366f1")) // indigo-500

	phpViteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eab308")) // yellow-500

	phpHorizonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f97316")) // orange-500

	phpReverbStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8b5cf6")) // violet-500

	phpRedisStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")) // red-500

	phpStatusRunning = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22c55e")). // green-500
				Bold(true)

	phpStatusStopped = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6b7280")) // gray-500

	phpStatusError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")). // red-500
			Bold(true)
)

// QA command styles
var (
	phpQAPassedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22c55e")). // green-500
				Bold(true)

	phpQAFailedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ef4444")). // red-500
				Bold(true)

	phpQAWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f59e0b")). // amber-500
				Bold(true)

	phpQAStageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6366f1")). // indigo-500
			Bold(true)

	phpSecurityCriticalStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#ef4444")). // red-500
					Bold(true)

	phpSecurityHighStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f97316")). // orange-500
				Bold(true)

	phpSecurityMediumStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f59e0b")) // amber-500

	phpSecurityLowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6b7280")) // gray-500
)

// AddPHPCommands adds PHP/Laravel development commands.
func AddPHPCommands(parent *clir.Cli) {
	phpCmd := parent.NewSubCommand("php", "Laravel/PHP development tools")
	phpCmd.LongDescription("Manage Laravel development environment with FrankenPHP.\n\n" +
		"Services orchestrated:\n" +
		"  - FrankenPHP/Octane (port 8000, HTTPS on 443)\n" +
		"  - Vite dev server (port 5173)\n" +
		"  - Laravel Horizon (queue workers)\n" +
		"  - Laravel Reverb (WebSocket, port 8080)\n" +
		"  - Redis (port 6379)")

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
