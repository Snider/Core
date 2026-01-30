package dev

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

// CI-specific styles (aliases to shared)
var (
	ciSuccessStyle = cli.SuccessStyle
	ciFailureStyle = cli.ErrorStyle
	ciPendingStyle = cli.StatusWarningStyle
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
func addCICommand(parent *cobra.Command) {
	ciCmd := &cobra.Command{
		Use:   "ci",
		Short: i18n.T("cmd.dev.ci.short"),
		Long:  i18n.T("cmd.dev.ci.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf(i18n.T("error.gh_not_found"))
	}

	// Find or use provided registry
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return fmt.Errorf("failed to load registry: %w", err)
			}
		} else {
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
		}
	}

	// Fetch CI status sequentially
	var allRuns []WorkflowRun
	var fetchErrors []error
	var noCI []string

	repoList := reg.List()
	for i, repo := range repoList {
		repoFullName := fmt.Sprintf("%s/%s", reg.Org, repo.Name)
		fmt.Printf("\033[2K\r%s %d/%d %s", dimStyle.Render(i18n.T("cli.progress.checking")), i+1, len(repoList), repo.Name)

		runs, err := fetchWorkflowRuns(repoFullName, repo.Name, branch)
		if err != nil {
			if strings.Contains(err.Error(), "no workflows") {
				noCI = append(noCI, repo.Name)
			} else {
				fetchErrors = append(fetchErrors, fmt.Errorf("%s: %w", repo.Name, err))
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
	fmt.Print("\033[2K\r") // Clear progress line

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
	fmt.Println()
	fmt.Printf("%s", i18n.T("cmd.dev.ci.repos_checked", map[string]interface{}{"Count": len(repoList)}))
	if success > 0 {
		fmt.Printf(" * %s", ciSuccessStyle.Render(i18n.T("cmd.dev.ci.passing", map[string]interface{}{"Count": success})))
	}
	if failed > 0 {
		fmt.Printf(" * %s", ciFailureStyle.Render(i18n.T("cmd.dev.ci.failing", map[string]interface{}{"Count": failed})))
	}
	if pending > 0 {
		fmt.Printf(" * %s", ciPendingStyle.Render(i18n.T("common.count.pending", map[string]interface{}{"Count": pending})))
	}
	if len(noCI) > 0 {
		fmt.Printf(" * %s", ciSkippedStyle.Render(i18n.T("cmd.dev.ci.no_ci", map[string]interface{}{"Count": len(noCI)})))
	}
	fmt.Println()
	fmt.Println()

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
		fmt.Println()
		for _, err := range fetchErrors {
			fmt.Printf("%s %s\n", errorStyle.Render(i18n.T("common.label.error")), err)
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
			return nil, fmt.Errorf("%s", strings.TrimSpace(stderr))
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

	fmt.Printf("  %s %-18s %-22s %s\n",
		status,
		repoNameStyle.Render(run.RepoName),
		dimStyle.Render(workflowName),
		issueAgeStyle.Render(age),
	)
}
