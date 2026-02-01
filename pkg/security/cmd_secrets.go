package security

import (
	"encoding/json"
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

func addSecretsCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "secrets",
		Short: i18n.T("cmd.security.secrets.short"),
		Long:  i18n.T("cmd.security.secrets.long"),
		RunE: func(c *cli.Command, args []string) error {
			return runSecrets()
		},
	}

	cmd.Flags().StringVar(&securityRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	cmd.Flags().StringVar(&securityRepo, "repo", "", i18n.T("cmd.security.flag.repo"))
	cmd.Flags().BoolVar(&securityJSON, "json", false, i18n.T("common.flag.json"))

	parent.AddCommand(cmd)
}

// SecretAlert represents a secret scanning alert for output.
type SecretAlert struct {
	Repo           string `json:"repo"`
	Number         int    `json:"number"`
	SecretType     string `json:"secret_type"`
	State          string `json:"state"`
	Resolution     string `json:"resolution,omitempty"`
	PushProtection bool   `json:"push_protection_bypassed"`
}

func runSecrets() error {
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

	var allAlerts []SecretAlert
	openCount := 0

	for _, repo := range repoList {
		repoFullName := fmt.Sprintf("%s/%s", reg.Org, repo.Name)

		alerts, err := fetchSecretScanningAlerts(repoFullName)
		if err != nil {
			continue
		}

		for _, alert := range alerts {
			if alert.State != "open" {
				continue
			}
			openCount++

			secretAlert := SecretAlert{
				Repo:           repo.Name,
				Number:         alert.Number,
				SecretType:     alert.SecretType,
				State:          alert.State,
				Resolution:     alert.Resolution,
				PushProtection: alert.PushProtection,
			}
			allAlerts = append(allAlerts, secretAlert)
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
	if openCount > 0 {
		cli.Print("%s %s\n", cli.DimStyle.Render("Secrets:"), cli.ErrorStyle.Render(fmt.Sprintf("%d open", openCount)))
	} else {
		cli.Print("%s %s\n", cli.DimStyle.Render("Secrets:"), cli.SuccessStyle.Render("No exposed secrets"))
	}
	cli.Blank()

	if len(allAlerts) == 0 {
		return nil
	}

	// Print table
	for _, alert := range allAlerts {
		bypassed := ""
		if alert.PushProtection {
			bypassed = cli.WarningStyle.Render(" (push protection bypassed)")
		}

		cli.Print("%-16s %-6d %-30s%s\n",
			cli.ValueStyle.Render(alert.Repo),
			alert.Number,
			cli.ErrorStyle.Render(alert.SecretType),
			bypassed,
		)
	}
	cli.Blank()

	return nil
}
