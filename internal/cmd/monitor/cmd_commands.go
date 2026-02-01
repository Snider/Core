// Package monitor provides security monitoring commands.
//
// Commands:
//   - monitor: Aggregate security findings from GitHub Security Tab, workflow artifacts, and PR comments
//
// Data sources (all free tier):
//   - Code scanning: Semgrep, Trivy, Gitleaks, OSV-Scanner, Checkov, CodeQL
//   - Dependabot: Dependency vulnerability alerts
//   - Secret scanning: Exposed secrets/credentials
package monitor

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

func init() {
	cli.RegisterCommands(AddMonitorCommands)
}

// Style aliases from shared package
var (
	successStyle = cli.SuccessStyle
	errorStyle   = cli.ErrorStyle
	warningStyle = cli.WarningStyle
	dimStyle     = cli.DimStyle
)

// AddMonitorCommands registers the 'monitor' command.
func AddMonitorCommands(root *cli.Command) {
	monitorCmd := &cli.Command{
		Use:   "monitor",
		Short: i18n.T("cmd.monitor.short"),
		Long:  i18n.T("cmd.monitor.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runMonitor()
		},
	}

	// Flags
	monitorCmd.Flags().StringVarP(&monitorRepo, "repo", "r", "", i18n.T("cmd.monitor.flag.repo"))
	monitorCmd.Flags().StringSliceVarP(&monitorSeverity, "severity", "s", []string{}, i18n.T("cmd.monitor.flag.severity"))
	monitorCmd.Flags().BoolVar(&monitorJSON, "json", false, i18n.T("cmd.monitor.flag.json"))
	monitorCmd.Flags().BoolVar(&monitorAll, "all", false, i18n.T("cmd.monitor.flag.all"))

	root.AddCommand(monitorCmd)
}
