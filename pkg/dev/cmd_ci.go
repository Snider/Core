package dev

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
)

// CI-specific styles (aliases to shared)
var (
	ciSuccessStyle = cli.SuccessStyle
	ciFailureStyle = cli.ErrorStyle
	ciPendingStyle = cli.WarningStyle
	ciSkippedStyle = cli.DimStyle
)

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	HeadBranch string    `json:"headBranch"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	URL        string    `json:"url"`

	// Added by us
	RepoName string `json:"-"`
}

// CI command flags
var (
	ciRegistryPath string
	ciBranch       string
	ciFailedOnly   bool
)

// addCICommand adds the 'ci' command to the given parent command.
func addCICommand(parent *cli.Command) {
	ciCmd := &cli.Command{
		Use:   "ci",
		Short: i18n.T("cmd.dev.ci.short"),
		Long:  i18n.T("cmd.dev.ci.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			branch := ciBranch
			if branch == "" {
				branch = "main"
			}
			return runCI(ciRegistryPath, branch, ciFailedOnly)
		},
	}

	ciCmd.Flags().StringVar(&ciRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	ciCmd.Flags().StringVarP(&ciBranch, "branch", "b", "main", i18n.T("cmd.dev.ci.flag.branch"))
	ciCmd.Flags().BoolVar(&ciFailedOnly, "failed", false, i18n.T("cmd.dev.ci.flag.failed"))

	parent.AddCommand(ciCmd)
}

func runCI(registryPath string, branch string, failedOnly bool) error {
	// Check gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return errors.New(i18n.T("error.gh_not_found"))
	}

	// Find or use provided registry
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return cli.Wrap(err, "failed to load registry")
		}
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return cli.Wrap(err, "failed to load registry")
			}
		} else {
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return cli.Wrap(err, "failed to scan directory")
			}
		}
	}

	// Fetch CI status sequentially
	var allRuns []WorkflowRun
	var fetchErrors []error
	var noCI []string

	repoList := reg.List()
	for i, repo := range repoList {
		repoFullName := cli.Sprintf("%s/%s", reg.Org, repo.Name)
		cli.Print("\033[2K\r%s %d/%d %s", dimStyle.Render(i18n.T("i18n.progress.check")), i+1, len(repoList), repo.Name)

		runs, err := fetchWorkflowRuns(repoFullName, repo.Name, branch)
		if err != nil {
			if strings.Contains(err.Error(), "no workflows") {
				noCI = append(noCI, repo.Name)
			} else {
				fetchErrors = append(fetchErrors, cli.Wrap(err, repo.Name))
			}
			continue
		}

		if len(runs) > 0 {
			// Just get the latest run
			allRuns = append(allRuns, runs[0])
		} else {
			noCI = append(noCI, repo.Name)
		}
	}
	cli.Print("\033[2K\r") // Clear progress line

	// Count by status
	var success, failed, pending, other int
	for _, run := range allRuns {
		switch run.Conclusion {
		case "success":
			success++
		case "failure":
			failed++
		case "":
			if run.Status == "in_progress" || run.Status == "queued" {
				pending++
			} else {
				other++
			}
		default:
			other++
		}
	}

	// Print summary
	cli.Blank()
	cli.Print("%s", i18n.T("cmd.dev.ci.repos_checked", map[string]interface{}{"Count": len(repoList)}))
	if success > 0 {
		cli.Print(" * %s", ciSuccessStyle.Render(i18n.T("cmd.dev.ci.passing", map[string]interface{}{"Count": success})))
	}
	if failed > 0 {
		cli.Print(" * %s", ciFailureStyle.Render(i18n.T("cmd.dev.ci.failing", map[string]interface{}{"Count": failed})))
	}
	if pending > 0 {
		cli.Print(" * %s", ciPendingStyle.Render(i18n.T("common.count.pending", map[string]interface{}{"Count": pending})))
	}
	if len(noCI) > 0 {
		cli.Print(" * %s", ciSkippedStyle.Render(i18n.T("cmd.dev.ci.no_ci", map[string]interface{}{"Count": len(noCI)})))
	}
	cli.Blank()
	cli.Blank()

	// Filter if needed
	displayRuns := allRuns
	if failedOnly {
		displayRuns = nil
		for _, run := range allRuns {
			if run.Conclusion == "failure" {
				displayRuns = append(displayRuns, run)
			}
		}
	}

	// Print details
	for _, run := range displayRuns {
		printWorkflowRun(run)
	}

	// Print errors
	if len(fetchErrors) > 0 {
		cli.Blank()
		for _, err := range fetchErrors {
			cli.Print("%s %s\n", errorStyle.Render(i18n.Label("error")), err)
		}
	}

	return nil
}

func fetchWorkflowRuns(repoFullName, repoName string, branch string) ([]WorkflowRun, error) {
	args := []string{
		"run", "list",
		"--repo", repoFullName,
		"--branch", branch,
		"--limit", "1",
		"--json", "name,status,conclusion,headBranch,createdAt,updatedAt,url",
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			return nil, cli.Err("%s", strings.TrimSpace(stderr))
		}
		return nil, err
	}

	var runs []WorkflowRun
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, err
	}

	// Tag with repo name
	for i := range runs {
		runs[i].RepoName = repoName
	}

	return runs, nil
}

func printWorkflowRun(run WorkflowRun) {
	// Status icon
	var status string
	switch run.Conclusion {
	case "success":
		status = ciSuccessStyle.Render("v")
	case "failure":
		status = ciFailureStyle.Render("x")
	case "":
		if run.Status == "in_progress" {
			status = ciPendingStyle.Render("*")
		} else if run.Status == "queued" {
			status = ciPendingStyle.Render("o")
		} else {
			status = ciSkippedStyle.Render("-")
		}
	case "skipped":
		status = ciSkippedStyle.Render("-")
	case "cancelled":
		status = ciSkippedStyle.Render("o")
	default:
		status = ciSkippedStyle.Render("?")
	}

	// Workflow name (truncated)
	workflowName := cli.Truncate(run.Name, 20)

	// Age
	age := cli.FormatAge(run.UpdatedAt)

	cli.Print("  %s %-18s %-22s %s\n",
		status,
		repoNameStyle.Render(run.RepoName),
		dimStyle.Render(workflowName),
		issueAgeStyle.Render(age),
	)
}
