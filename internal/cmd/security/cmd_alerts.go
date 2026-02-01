package security

import (
	"encoding/json"
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

func addAlertsCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "alerts",
		Short: i18n.T("cmd.security.alerts.short"),
		Long:  i18n.T("cmd.security.alerts.long"),
		RunE: func(c *cli.Command, args []string) error {
			return runAlerts()
		},
	}

	cmd.Flags().StringVar(&securityRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	cmd.Flags().StringVar(&securityRepo, "repo", "", i18n.T("cmd.security.flag.repo"))
	cmd.Flags().StringVar(&securitySeverity, "severity", "", i18n.T("cmd.security.flag.severity"))
	cmd.Flags().BoolVar(&securityJSON, "json", false, i18n.T("common.flag.json"))

	parent.AddCommand(cmd)
}

// AlertOutput represents a unified alert for output.
type AlertOutput struct {
	Repo     string `json:"repo"`
	Severity string `json:"severity"`
	ID       string `json:"id"`
	Package  string `json:"package,omitempty"`
	Version  string `json:"version,omitempty"`
	Location string `json:"location,omitempty"`
	Type     string `json:"type"`
	Message  string `json:"message"`
}

func runAlerts() error {
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

	var allAlerts []AlertOutput
	summary := &AlertSummary{}

	for _, repo := range repoList {
		repoFullName := fmt.Sprintf("%s/%s", reg.Org, repo.Name)

		// Fetch Dependabot alerts
		depAlerts, err := fetchDependabotAlerts(repoFullName)
		if err == nil {
			for _, alert := range depAlerts {
				if alert.State != "open" {
					continue
				}
				severity := alert.Advisory.Severity
				if !filterBySeverity(severity, securitySeverity) {
					continue
				}
				summary.Add(severity)
				allAlerts = append(allAlerts, AlertOutput{
					Repo:     repo.Name,
					Severity: severity,
					ID:       alert.Advisory.CVEID,
					Package:  alert.Dependency.Package.Name,
					Version:  alert.SecurityVulnerability.VulnerableVersionRange,
					Type:     "dependabot",
					Message:  alert.Advisory.Summary,
				})
			}
		}

		// Fetch code scanning alerts
		codeAlerts, err := fetchCodeScanningAlerts(repoFullName)
		if err == nil {
			for _, alert := range codeAlerts {
				if alert.State != "open" {
					continue
				}
				severity := alert.Rule.Severity
				if !filterBySeverity(severity, securitySeverity) {
					continue
				}
				summary.Add(severity)
				location := fmt.Sprintf("%s:%d", alert.MostRecentInstance.Location.Path, alert.MostRecentInstance.Location.StartLine)
				allAlerts = append(allAlerts, AlertOutput{
					Repo:     repo.Name,
					Severity: severity,
					ID:       alert.Rule.ID,
					Location: location,
					Type:     alert.Tool.Name,
					Message:  alert.Rule.Description,
				})
			}
		}

		// Fetch secret scanning alerts
		secretAlerts, err := fetchSecretScanningAlerts(repoFullName)
		if err == nil {
			for _, alert := range secretAlerts {
				if alert.State != "open" {
					continue
				}
				if !filterBySeverity("high", securitySeverity) {
					continue
				}
				summary.Add("high") // Secrets are always high severity
				allAlerts = append(allAlerts, AlertOutput{
					Repo:     repo.Name,
					Severity: "high",
					ID:       fmt.Sprintf("secret-%d", alert.Number),
					Type:     "secret-scanning",
					Message:  alert.SecretType,
				})
			}
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
	cli.Print("%s %s\n", cli.DimStyle.Render("Alerts:"), summary.String())
	cli.Blank()

	if len(allAlerts) == 0 {
		return nil
	}

	// Print table
	for _, alert := range allAlerts {
		sevStyle := severityStyle(alert.Severity)

		// Format: repo  SEVERITY  ID  package/location  type
		location := alert.Package
		if location == "" {
			location = alert.Location
		}
		if alert.Version != "" {
			location = fmt.Sprintf("%s %s", location, cli.DimStyle.Render(alert.Version))
		}

		cli.Print("%-20s %s  %-16s %-40s %s\n",
			cli.ValueStyle.Render(alert.Repo),
			sevStyle.Render(fmt.Sprintf("%-8s", alert.Severity)),
			alert.ID,
			location,
			cli.DimStyle.Render(alert.Type),
		)
	}
	cli.Blank()

	return nil
}

func fetchDependabotAlerts(repoFullName string) ([]DependabotAlert, error) {
	endpoint := fmt.Sprintf("repos/%s/dependabot/alerts?state=open", repoFullName)
	output, err := runGHAPI(endpoint)
	if err != nil {
		return nil, cli.Wrap(err, fmt.Sprintf("fetch dependabot alerts for %s", repoFullName))
	}

	var alerts []DependabotAlert
	if err := json.Unmarshal(output, &alerts); err != nil {
		return nil, cli.Wrap(err, fmt.Sprintf("parse dependabot alerts for %s", repoFullName))
	}
	return alerts, nil
}

func fetchCodeScanningAlerts(repoFullName string) ([]CodeScanningAlert, error) {
	endpoint := fmt.Sprintf("repos/%s/code-scanning/alerts?state=open", repoFullName)
	output, err := runGHAPI(endpoint)
	if err != nil {
		return nil, cli.Wrap(err, fmt.Sprintf("fetch code-scanning alerts for %s", repoFullName))
	}

	var alerts []CodeScanningAlert
	if err := json.Unmarshal(output, &alerts); err != nil {
		return nil, cli.Wrap(err, fmt.Sprintf("parse code-scanning alerts for %s", repoFullName))
	}
	return alerts, nil
}

func fetchSecretScanningAlerts(repoFullName string) ([]SecretScanningAlert, error) {
	endpoint := fmt.Sprintf("repos/%s/secret-scanning/alerts?state=open", repoFullName)
	output, err := runGHAPI(endpoint)
	if err != nil {
		return nil, cli.Wrap(err, fmt.Sprintf("fetch secret-scanning alerts for %s", repoFullName))
	}

	var alerts []SecretScanningAlert
	if err := json.Unmarshal(output, &alerts); err != nil {
		return nil, cli.Wrap(err, fmt.Sprintf("parse secret-scanning alerts for %s", repoFullName))
	}
	return alerts, nil
}
