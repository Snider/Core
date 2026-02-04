package security

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/ai"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

var (
	jobsTargets   []string
	jobsIssueRepo string
	jobsDryRun    bool
	jobsCopies    int
)

func addJobsCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "jobs",
		Short: i18n.T("cmd.security.jobs.short"),
		Long:  i18n.T("cmd.security.jobs.long"),
		RunE: func(c *cli.Command, args []string) error {
			return runJobs()
		},
	}

	cmd.Flags().StringSliceVar(&jobsTargets, "targets", nil, i18n.T("cmd.security.jobs.flag.targets"))
	cmd.Flags().StringVar(&jobsIssueRepo, "issue-repo", "host-uk/core", i18n.T("cmd.security.jobs.flag.issue_repo"))
	cmd.Flags().BoolVar(&jobsDryRun, "dry-run", false, i18n.T("cmd.security.jobs.flag.dry_run"))
	cmd.Flags().IntVar(&jobsCopies, "copies", 1, i18n.T("cmd.security.jobs.flag.copies"))

	parent.AddCommand(cmd)
}

func runJobs() error {
	if err := checkGH(); err != nil {
		return err
	}

	if len(jobsTargets) == 0 {
		return cli.Err("at least one --targets value required (e.g. --targets wailsapp/wails)")
	}

	if jobsCopies < 1 {
		return cli.Err("--copies must be at least 1")
	}

	var failedCount int
	for _, target := range jobsTargets {
		if err := createJobForTarget(target); err != nil {
			cli.Print("%s %s: %v\n", cli.ErrorStyle.Render(">>"), target, err)
			failedCount++
			continue
		}
	}

	if failedCount == len(jobsTargets) {
		return cli.Err("all targets failed to process")
	}

	return nil
}

func createJobForTarget(target string) error {
	parts := strings.SplitN(target, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: use owner/repo")
	}

	// Gather findings
	summary := &AlertSummary{}
	var findings []string
	var fetchErrors int

	// Code scanning
	codeAlerts, err := fetchCodeScanningAlerts(target)
	if err != nil {
		cli.Print("%s %s: failed to fetch code scanning alerts: %v\n", cli.WarningStyle.Render(">>"), target, err)
		fetchErrors++
	}
	if err == nil {
		for _, alert := range codeAlerts {
			if alert.State != "open" {
				continue
			}
			severity := alert.Rule.Severity
			if severity == "" {
				severity = "medium"
			}
			summary.Add(severity)
			findings = append(findings, fmt.Sprintf("- [%s] %s: %s (%s:%d)",
				strings.ToUpper(severity), alert.Tool.Name, alert.Rule.Description,
				alert.MostRecentInstance.Location.Path, alert.MostRecentInstance.Location.StartLine))
		}
	}

	// Dependabot
	depAlerts, err := fetchDependabotAlerts(target)
	if err != nil {
		cli.Print("%s %s: failed to fetch dependabot alerts: %v\n", cli.WarningStyle.Render(">>"), target, err)
		fetchErrors++
	}
	if err == nil {
		for _, alert := range depAlerts {
			if alert.State != "open" {
				continue
			}
			summary.Add(alert.Advisory.Severity)
			findings = append(findings, fmt.Sprintf("- [%s] %s: %s (%s)",
				strings.ToUpper(alert.Advisory.Severity), alert.Dependency.Package.Name,
				alert.Advisory.Summary, alert.Advisory.CVEID))
		}
	}

	// Secret scanning
	secretAlerts, err := fetchSecretScanningAlerts(target)
	if err != nil {
		cli.Print("%s %s: failed to fetch secret scanning alerts: %v\n", cli.WarningStyle.Render(">>"), target, err)
		fetchErrors++
	}
	if err == nil {
		for _, alert := range secretAlerts {
			if alert.State != "open" {
				continue
			}
			summary.Add("high")
			findings = append(findings, fmt.Sprintf("- [HIGH] Secret: %s (#%d)", alert.SecretType, alert.Number))
		}
	}

	if fetchErrors == 3 {
		return fmt.Errorf("failed to fetch any alerts for %s", target)
	}

	if summary.Total == 0 {
		cli.Print("%s %s: %s\n", cli.SuccessStyle.Render(">>"), target, "No open findings")
		return nil
	}

	// Build issue body
	title := fmt.Sprintf("Security scan: %s", target)
	body := buildJobIssueBody(target, summary, findings)

	for i := range jobsCopies {
		issueTitle := title
		if jobsCopies > 1 {
			issueTitle = fmt.Sprintf("%s (#%d)", title, i+1)
		}

		if jobsDryRun {
			cli.Blank()
			cli.Print("%s %s\n", cli.DimStyle.Render("[dry-run] Would create issue:"), issueTitle)
			cli.Print("%s %s\n", cli.DimStyle.Render("  Repo:"), jobsIssueRepo)
			cli.Print("%s %s\n", cli.DimStyle.Render("  Labels:"), "type:security-scan,repo:"+target)
			cli.Print("%s %d findings\n", cli.DimStyle.Render("  Findings:"), summary.Total)
			continue
		}

		// Create issue via gh CLI
		cmd := exec.Command("gh", "issue", "create",
			"--repo", jobsIssueRepo,
			"--title", issueTitle,
			"--body", body,
			"--label", "type:security-scan,repo:"+target,
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return cli.Wrap(err, fmt.Sprintf("create issue for %s: %s", target, string(output)))
		}

		issueURL := strings.TrimSpace(string(output))
		cli.Print("%s %s: %s\n", cli.SuccessStyle.Render(">>"), issueTitle, issueURL)

		// Record metrics
		_ = ai.Record(ai.Event{
			Type:      "security.job_created",
			Timestamp: time.Now(),
			Repo:      target,
			Data: map[string]any{
				"issue_repo": jobsIssueRepo,
				"issue_url":  issueURL,
				"total":      summary.Total,
				"critical":   summary.Critical,
				"high":       summary.High,
			},
		})
	}

	return nil
}

func buildJobIssueBody(target string, summary *AlertSummary, findings []string) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Security Scan: %s\n\n", target)
	fmt.Fprintf(&sb, "**Summary:** %s\n\n", summary.String())

	sb.WriteString("### Findings\n\n")
	if len(findings) > 50 {
		// Truncate long lists
		for _, f := range findings[:50] {
			sb.WriteString(f + "\n")
		}
		fmt.Fprintf(&sb, "\n... and %d more\n", len(findings)-50)
	} else {
		for _, f := range findings {
			sb.WriteString(f + "\n")
		}
	}

	sb.WriteString("\n### Checklist\n\n")
	sb.WriteString("- [ ] Review findings above\n")
	sb.WriteString("- [ ] Triage by severity (critical/high first)\n")
	sb.WriteString("- [ ] Create PRs for fixes\n")
	sb.WriteString("- [ ] Verify fixes resolve alerts\n")

	sb.WriteString("\n### Instructions\n\n")
	sb.WriteString("1. Claim this issue by assigning yourself\n")
	fmt.Fprintf(&sb, "2. Run `core security alerts --target %s` for the latest findings\n", target)
	sb.WriteString("3. Work through the checklist above\n")
	sb.WriteString("4. Close this issue when all findings are addressed\n")

	return sb.String()
}
