package php

import (
	"os"
	"path/filepath"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/workspace"
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check if we are in a workspace root
			wsRoot, err := workspace.FindWorkspaceRoot()
			if err != nil {
				return nil // Not in a workspace, regular behavior
			}

			// Load workspace config
			config, err := workspace.LoadConfig(wsRoot)
			if err != nil || config == nil {
				return nil // Failed to load or no config, ignore
			}

			if config.Active == "" {
				return nil // No active package
			}

			// Calculate package path
			pkgDir := config.PackagesDir
			if pkgDir == "" {
				pkgDir = "./packages"
			}
			if !filepath.IsAbs(pkgDir) {
				pkgDir = filepath.Join(wsRoot, pkgDir)
			}

			targetDir := filepath.Join(pkgDir, config.Active)

			// Check if target directory exists
			if _, err := os.Stat(targetDir); err != nil {
				cli.Warnf("Active package directory not found: %s", targetDir)
				return nil
			}

			// Change working directory
			if err := os.Chdir(targetDir); err != nil {
				return cli.Err("failed to change directory to active package: %w", err)
			}

			cli.Print("%s %s\n", dimStyle.Render("Workspace:"), config.Active)
			return nil
		},
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