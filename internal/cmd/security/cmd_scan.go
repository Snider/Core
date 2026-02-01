package security

import (
	"encoding/json"
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

var (
	scanTool string
)

func addScanCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "scan",
		Short: i18n.T("cmd.security.scan.short"),
		Long:  i18n.T("cmd.security.scan.long"),
		RunE: func(c *cli.Command, args []string) error {
			return runScan()
		},
	}

	cmd.Flags().StringVar(&securityRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	cmd.Flags().StringVar(&securityRepo, "repo", "", i18n.T("cmd.security.flag.repo"))
	cmd.Flags().StringVar(&securitySeverity, "severity", "", i18n.T("cmd.security.flag.severity"))
	cmd.Flags().StringVar(&scanTool, "tool", "", i18n.T("cmd.security.scan.flag.tool"))
	cmd.Flags().BoolVar(&securityJSON, "json", false, i18n.T("common.flag.json"))

	parent.AddCommand(cmd)
}

// ScanAlert represents a code scanning alert for output.
type ScanAlert struct {
	Repo        string `json:"repo"`
	Severity    string `json:"severity"`
	RuleID      string `json:"rule_id"`
	Tool        string `json:"tool"`
	Path        string `json:"path"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	Message     string `json:"message"`
}

func runScan() error {
	if err := checkGH(); err != nil {
		return err
	}

	reg, err := loadRegistry(securityRegistryPath)
	if err != nil {
		return err
	}

	repoList := getReposToCheck(reg, securityRepo)
	if len(repoList) == 0 {
		return cli.Err("repo not found: %s", securityRepo)
	}

	var allAlerts []ScanAlert
	summary := &AlertSummary{}

	for _, repo := range repoList {
		repoFullName := fmt.Sprintf("%s/%s", reg.Org, repo.Name)

		alerts, err := fetchCodeScanningAlerts(repoFullName)
		if err != nil {
			continue
		}

		for _, alert := range alerts {
			if alert.State != "open" {
				continue
			}

			// Filter by tool if specified
			if scanTool != "" && alert.Tool.Name != scanTool {
				continue
			}

			severity := alert.Rule.Severity
			if severity == "" {
				severity = "medium" // Default if not specified
			}

			if !filterBySeverity(severity, securitySeverity) {
				continue
			}

			summary.Add(severity)

			scanAlert := ScanAlert{
				Repo:        repo.Name,
				Severity:    severity,
				RuleID:      alert.Rule.ID,
				Tool:        alert.Tool.Name,
				Path:        alert.MostRecentInstance.Location.Path,
				Line:        alert.MostRecentInstance.Location.StartLine,
				Description: alert.Rule.Description,
				Message:     alert.MostRecentInstance.Message.Text,
			}
			allAlerts = append(allAlerts, scanAlert)
		}
	}

	if securityJSON {
		output, err := json.MarshalIndent(allAlerts, "", "  ")
		if err != nil {
			return cli.Wrap(err, "marshal JSON output")
		}
		cli.Text(string(output))
		return nil
	}

	// Print summary
	cli.Blank()
	cli.Print("%s %s\n", cli.DimStyle.Render("Code Scanning:"), summary.String())
	cli.Blank()

	if len(allAlerts) == 0 {
		return nil
	}

	// Print table
	for _, alert := range allAlerts {
		sevStyle := severityStyle(alert.Severity)

		location := fmt.Sprintf("%s:%d", alert.Path, alert.Line)

		cli.Print("%-16s %s  %-20s %-40s %s\n",
			cli.ValueStyle.Render(alert.Repo),
			sevStyle.Render(fmt.Sprintf("%-8s", alert.Severity)),
			alert.RuleID,
			location,
			cli.DimStyle.Render(alert.Tool),
		)
	}
	cli.Blank()

	return nil
}
