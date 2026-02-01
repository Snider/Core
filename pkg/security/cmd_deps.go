package security

import (
	"encoding/json"
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

func addDepsCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "deps",
		Short: i18n.T("cmd.security.deps.short"),
		Long:  i18n.T("cmd.security.deps.long"),
		RunE: func(c *cli.Command, args []string) error {
			return runDeps()
		},
	}

	cmd.Flags().StringVar(&securityRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	cmd.Flags().StringVar(&securityRepo, "repo", "", i18n.T("cmd.security.flag.repo"))
	cmd.Flags().StringVar(&securitySeverity, "severity", "", i18n.T("cmd.security.flag.severity"))
	cmd.Flags().BoolVar(&securityJSON, "json", false, i18n.T("common.flag.json"))

	parent.AddCommand(cmd)
}

// DepAlert represents a dependency vulnerability for output.
type DepAlert struct {
	Repo         string `json:"repo"`
	Severity     string `json:"severity"`
	CVE          string `json:"cve"`
	Package      string `json:"package"`
	Ecosystem    string `json:"ecosystem"`
	Vulnerable   string `json:"vulnerable_range"`
	PatchedVersion string `json:"patched_version,omitempty"`
	Manifest     string `json:"manifest"`
	Summary      string `json:"summary"`
}

func runDeps() error {
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

	var allAlerts []DepAlert
	summary := &AlertSummary{}

	for _, repo := range repoList {
		repoFullName := fmt.Sprintf("%s/%s", reg.Org, repo.Name)

		alerts, err := fetchDependabotAlerts(repoFullName)
		if err != nil {
			continue
		}

		for _, alert := range alerts {
			if alert.State != "open" {
				continue
			}

			severity := alert.Advisory.Severity
			if !filterBySeverity(severity, securitySeverity) {
				continue
			}

			summary.Add(severity)

			depAlert := DepAlert{
				Repo:           repo.Name,
				Severity:       severity,
				CVE:            alert.Advisory.CVEID,
				Package:        alert.Dependency.Package.Name,
				Ecosystem:      alert.Dependency.Package.Ecosystem,
				Vulnerable:     alert.SecurityVulnerability.VulnerableVersionRange,
				PatchedVersion: alert.SecurityVulnerability.FirstPatchedVersion.Identifier,
				Manifest:       alert.Dependency.ManifestPath,
				Summary:        alert.Advisory.Summary,
			}
			allAlerts = append(allAlerts, depAlert)
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
	cli.Print("%s %s\n", cli.DimStyle.Render("Dependabot:"), summary.String())
	cli.Blank()

	if len(allAlerts) == 0 {
		return nil
	}

	// Print table
	for _, alert := range allAlerts {
		sevStyle := severityStyle(alert.Severity)

		// Format upgrade suggestion
		upgrade := alert.Vulnerable
		if alert.PatchedVersion != "" {
			upgrade = fmt.Sprintf("%s -> %s", alert.Vulnerable, cli.SuccessStyle.Render(alert.PatchedVersion))
		}

		cli.Print("%-16s %s  %-16s %-30s %s\n",
			cli.ValueStyle.Render(alert.Repo),
			sevStyle.Render(fmt.Sprintf("%-8s", alert.Severity)),
			alert.CVE,
			alert.Package,
			upgrade,
		)
	}
	cli.Blank()

	return nil
}
