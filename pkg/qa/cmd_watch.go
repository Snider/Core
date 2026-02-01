// cmd_watch.go implements the 'qa watch' command for monitoring GitHub Actions.
//
// Usage:
//   core qa watch              # Watch current repo's latest push
//   core qa watch --repo X     # Watch specific repo
//   core qa watch --commit SHA # Watch specific commit
//   core qa watch --timeout 5m # Custom timeout (default: 10m)

package qa

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

// Watch command flags
var (
	watchRepo    string
	watchCommit  string
	watchTimeout time.Duration
)

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	ID           int64     `json:"databaseId"`
	Name         string    `json:"name"`
	DisplayTitle string    `json:"displayTitle"`
	Status       string    `json:"status"`
	Conclusion   string    `json:"conclusion"`
	HeadSha      string    `json:"headSha"`
	URL          string    `json:"url"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// WorkflowJob represents a job within a workflow run
type WorkflowJob struct {
	ID         int64  `json:"databaseId"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	URL        string `json:"url"`
}

// JobStep represents a step within a job
type JobStep struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Number     int    `json:"number"`
}

// addWatchCommand adds the 'watch' subcommand to the qa command.
func addWatchCommand(parent *cli.Command) {
	watchCmd := &cli.Command{
		Use:   "watch",
		Short: i18n.T("cmd.qa.watch.short"),
		Long:  i18n.T("cmd.qa.watch.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runWatch()
		},
	}

	watchCmd.Flags().StringVarP(&watchRepo, "repo", "r", "", i18n.T("cmd.qa.watch.flag.repo"))
	watchCmd.Flags().StringVarP(&watchCommit, "commit", "c", "", i18n.T("cmd.qa.watch.flag.commit"))
	watchCmd.Flags().DurationVarP(&watchTimeout, "timeout", "t", 10*time.Minute, i18n.T("cmd.qa.watch.flag.timeout"))

	parent.AddCommand(watchCmd)
}

func runWatch() error {
	// Check gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return errors.New(i18n.T("error.gh_not_found"))
	}

	// Determine repo
	repoFullName, err := resolveRepo(watchRepo)
	if err != nil {
		return err
	}

	// Determine commit
	commitSha, err := resolveCommit(watchCommit)
	if err != nil {
		return err
	}

	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("repo")), repoFullName)
	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.qa.watch.commit")), commitSha[:8])
	cli.Blank()

	// Poll for workflow runs
	deadline := time.Now().Add(watchTimeout)
	pollInterval := 3 * time.Second
	var lastStatus string

	for time.Now().Before(deadline) {
		runs, err := fetchWorkflowRunsForCommit(repoFullName, commitSha)
		if err != nil {
			return cli.Wrap(err, "failed to fetch workflow runs")
		}

		if len(runs) == 0 {
			// No workflows triggered yet, keep waiting
			cli.Print("\033[2K\r%s", dimStyle.Render(i18n.T("cmd.qa.watch.waiting_for_workflows")))
			time.Sleep(pollInterval)
			continue
		}

		// Check status of all runs
		allComplete := true
		var pending, success, failed int
		for _, run := range runs {
			switch run.Status {
			case "completed":
				if run.Conclusion == "success" {
					success++
				} else if run.Conclusion == "failure" {
					failed++
				}
			default:
				allComplete = false
				pending++
			}
		}

		// Build status line
		status := fmt.Sprintf("%d workflow(s): ", len(runs))
		if pending > 0 {
			status += warningStyle.Render(fmt.Sprintf("%d running", pending))
			if success > 0 || failed > 0 {
				status += ", "
			}
		}
		if success > 0 {
			status += successStyle.Render(fmt.Sprintf("%d passed", success))
			if failed > 0 {
				status += ", "
			}
		}
		if failed > 0 {
			status += errorStyle.Render(fmt.Sprintf("%d failed", failed))
		}

		// Only print if status changed
		if status != lastStatus {
			cli.Print("\033[2K\r%s", status)
			lastStatus = status
		}

		if allComplete {
			cli.Blank()
			cli.Blank()
			return printResults(repoFullName, runs)
		}

		time.Sleep(pollInterval)
	}

	cli.Blank()
	return errors.New(i18n.T("cmd.qa.watch.timeout", map[string]interface{}{"Duration": watchTimeout}))
}

// resolveRepo determines the repo to watch
func resolveRepo(specified string) (string, error) {
	if specified != "" {
		// If it contains /, assume it's already full name
		if strings.Contains(specified, "/") {
			return specified, nil
		}
		// Try to get org from current directory
		org := detectOrgFromGit()
		if org != "" {
			return org + "/" + specified, nil
		}
		return "", errors.New(i18n.T("cmd.qa.watch.error.repo_format"))
	}

	// Detect from current directory
	return detectRepoFromGit()
}

// resolveCommit determines the commit to watch
func resolveCommit(specified string) (string, error) {
	if specified != "" {
		return specified, nil
	}

	// Get HEAD commit
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", cli.Wrap(err, "failed to get HEAD commit")
	}

	return strings.TrimSpace(string(output)), nil
}

// detectRepoFromGit detects the repo from git remote
func detectRepoFromGit() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.New(i18n.T("cmd.qa.watch.error.not_git_repo"))
	}

	url := strings.TrimSpace(string(output))
	return parseGitHubRepo(url)
}

// detectOrgFromGit tries to detect the org from git remote
func detectOrgFromGit() string {
	repo, err := detectRepoFromGit()
	if err != nil {
		return ""
	}
	parts := strings.Split(repo, "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

// parseGitHubRepo extracts org/repo from a git URL
func parseGitHubRepo(url string) (string, error) {
	// Handle SSH URLs: git@github.com:org/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		path := strings.TrimPrefix(url, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		return path, nil
	}

	// Handle HTTPS URLs: https://github.com/org/repo.git
	if strings.Contains(url, "github.com/") {
		parts := strings.Split(url, "github.com/")
		if len(parts) >= 2 {
			path := strings.TrimSuffix(parts[1], ".git")
			return path, nil
		}
	}

	return "", fmt.Errorf("could not parse GitHub repo from URL: %s", url)
}

// fetchWorkflowRunsForCommit fetches workflow runs for a specific commit
func fetchWorkflowRunsForCommit(repoFullName, commitSha string) ([]WorkflowRun, error) {
	args := []string{
		"run", "list",
		"--repo", repoFullName,
		"--commit", commitSha,
		"--json", "databaseId,name,displayTitle,status,conclusion,headSha,url,createdAt,updatedAt",
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, cli.Err("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}

	var runs []WorkflowRun
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, err
	}

	return runs, nil
}

// printResults prints the final results with actionable information
func printResults(repoFullName string, runs []WorkflowRun) error {
	var failures []WorkflowRun
	var successes []WorkflowRun

	for _, run := range runs {
		if run.Conclusion == "failure" {
			failures = append(failures, run)
		} else if run.Conclusion == "success" {
			successes = append(successes, run)
		}
	}

	// Print successes briefly
	for _, run := range successes {
		cli.Print("%s %s\n", successStyle.Render(cli.Glyph(":check:")), run.Name)
	}

	// Print failures with details
	for _, run := range failures {
		cli.Print("%s %s\n", errorStyle.Render(cli.Glyph(":cross:")), run.Name)

		// Fetch failed job details
		failedJob, failedStep, errorLine := fetchFailureDetails(repoFullName, run.ID)
		if failedJob != "" {
			cli.Print("  %s Job: %s", dimStyle.Render("->"), failedJob)
			if failedStep != "" {
				cli.Print(" (step: %s)", failedStep)
			}
			cli.Blank()
		}
		if errorLine != "" {
			cli.Print("  %s Error: %s\n", dimStyle.Render("->"), errorLine)
		}
		cli.Print("  %s %s\n", dimStyle.Render("->"), run.URL)
	}

	// Exit with error if any failures
	if len(failures) > 0 {
		cli.Blank()
		return cli.Err(i18n.T("cmd.qa.watch.workflows_failed", map[string]interface{}{"Count": len(failures)}))
	}

	cli.Blank()
	cli.Print("%s\n", successStyle.Render(i18n.T("cmd.qa.watch.all_passed")))
	return nil
}

// fetchFailureDetails fetches details about why a workflow failed
func fetchFailureDetails(repoFullName string, runID int64) (jobName, stepName, errorLine string) {
	// Fetch jobs for this run
	args := []string{
		"run", "view", fmt.Sprintf("%d", runID),
		"--repo", repoFullName,
		"--json", "jobs",
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", "", ""
	}

	var result struct {
		Jobs []struct {
			Name       string `json:"name"`
			Conclusion string `json:"conclusion"`
			Steps      []struct {
				Name       string `json:"name"`
				Conclusion string `json:"conclusion"`
				Number     int    `json:"number"`
			} `json:"steps"`
		} `json:"jobs"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return "", "", ""
	}

	// Find the failed job and step
	for _, job := range result.Jobs {
		if job.Conclusion == "failure" {
			jobName = job.Name
			for _, step := range job.Steps {
				if step.Conclusion == "failure" {
					stepName = fmt.Sprintf("%d: %s", step.Number, step.Name)
					break
				}
			}
			break
		}
	}

	// Try to get the error line from logs (if available)
	errorLine = fetchErrorFromLogs(repoFullName, runID, jobName)

	return jobName, stepName, errorLine
}

// fetchErrorFromLogs attempts to extract the first error line from workflow logs
func fetchErrorFromLogs(repoFullName string, runID int64, jobName string) string {
	// This would require downloading and parsing logs, which can be large
	// For now, return empty - users can follow the URL for details
	// Future: could use `gh run view --log-failed` and parse output
	return ""
}
