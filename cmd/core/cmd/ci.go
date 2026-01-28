package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/repos"
	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
)

var (
	ciSuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22c55e")) // green-500

	ciFailureStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ef4444")) // red-500

	ciPendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f59e0b")) // amber-500

	ciSkippedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")) // gray-500
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

// AddCICommand adds the 'ci' command to the given parent command.
func AddCICommand(parent *clir.Cli) {
	var registryPath string
	var branch string
	var failedOnly bool

	ciCmd := parent.NewSubCommand("ci", "Check CI status across all repos")
	ciCmd.LongDescription("Fetches GitHub Actions workflow status for all repos.\n" +
		"Shows latest run status for each repo.\n" +
		"Requires the 'gh' CLI to be installed and authenticated.")

	ciCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)
	ciCmd.StringFlag("branch", "Filter by branch (default: main)", &branch)
	ciCmd.BoolFlag("failed", "Show only failed runs", &failedOnly)

	ciCmd.Action(func() error {
		if branch == "" {
			branch = "main"
		}
		return runCI(registryPath, branch, failedOnly)
	})
}

func runCI(registryPath string, branch string, failedOnly bool) error {
	// Check gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("'gh' CLI not found. Install from https://cli.github.com/")
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
		fmt.Printf("\033[2K\r%s %d/%d %s", dimStyle.Render("Checking"), i+1, len(repoList), repo.Name)

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
	fmt.Printf("%d repos checked", len(repoList))
	if success > 0 {
		fmt.Printf(" · %s", ciSuccessStyle.Render(fmt.Sprintf("%d passing", success)))
	}
	if failed > 0 {
		fmt.Printf(" · %s", ciFailureStyle.Render(fmt.Sprintf("%d failing", failed)))
	}
	if pending > 0 {
		fmt.Printf(" · %s", ciPendingStyle.Render(fmt.Sprintf("%d pending", pending)))
	}
	if len(noCI) > 0 {
		fmt.Printf(" · %s", ciSkippedStyle.Render(fmt.Sprintf("%d no CI", len(noCI))))
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
			fmt.Printf("%s %s\n", errorStyle.Render("Error:"), err)
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
		status = ciSuccessStyle.Render("✓")
	case "failure":
		status = ciFailureStyle.Render("✗")
	case "":
		if run.Status == "in_progress" {
			status = ciPendingStyle.Render("●")
		} else if run.Status == "queued" {
			status = ciPendingStyle.Render("○")
		} else {
			status = ciSkippedStyle.Render("—")
		}
	case "skipped":
		status = ciSkippedStyle.Render("—")
	case "cancelled":
		status = ciSkippedStyle.Render("○")
	default:
		status = ciSkippedStyle.Render("?")
	}

	// Workflow name (truncated)
	workflowName := truncate(run.Name, 20)

	// Age
	age := formatAge(run.UpdatedAt)

	fmt.Printf("  %s %-18s %-22s %s\n",
		status,
		repoNameStyle.Render(run.RepoName),
		dimStyle.Render(workflowName),
		issueAgeStyle.Render(age),
	)
}
