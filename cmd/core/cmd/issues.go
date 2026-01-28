package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/repos"
	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
)

var (
	issueRepoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")) // gray-500

	issueNumberStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#3b82f6")) // blue-500

	issueTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")) // gray-200

	issueLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f59e0b")) // amber-500

	issueAssigneeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22c55e")) // green-500

	issueAgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")) // gray-500
)

// GitHubIssue represents a GitHub issue from the API.
type GitHubIssue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"createdAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Assignees struct {
		Nodes []struct {
			Login string `json:"login"`
		} `json:"nodes"`
	} `json:"assignees"`
	Labels struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
	URL string `json:"url"`

	// Added by us
	RepoName string `json:"-"`
}

// AddIssuesCommand adds the 'issues' command to the given parent command.
func AddIssuesCommand(parent *clir.Cli) {
	var registryPath string
	var limit int
	var assignee string

	issuesCmd := parent.NewSubCommand("issues", "List open issues across all repos")
	issuesCmd.LongDescription("Fetches open issues from GitHub for all repos in the registry.\n" +
		"Requires the 'gh' CLI to be installed and authenticated.")

	issuesCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)
	issuesCmd.IntFlag("limit", "Max issues per repo (default 10)", &limit)
	issuesCmd.StringFlag("assignee", "Filter by assignee (use @me for yourself)", &assignee)

	issuesCmd.Action(func() error {
		if limit == 0 {
			limit = 10
		}
		return runIssues(registryPath, limit, assignee)
	})
}

func runIssues(registryPath string, limit int, assignee string) error {
	// Check gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("'gh' CLI not found. Install from https://cli.github.com/")
	}

	// Find or use provided registry, fall back to directory scan
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
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
		}
	}

	// Fetch issues sequentially (avoid GitHub rate limits)
	var allIssues []GitHubIssue
	var fetchErrors []error

	repoList := reg.List()
	for i, repo := range repoList {
		repoFullName := fmt.Sprintf("%s/%s", reg.Org, repo.Name)
		fmt.Printf("\033[2K\r%s %d/%d %s", dimStyle.Render("Fetching"), i+1, len(repoList), repo.Name)

		issues, err := fetchIssues(repoFullName, repo.Name, limit, assignee)
		if err != nil {
			fetchErrors = append(fetchErrors, fmt.Errorf("%s: %w", repo.Name, err))
			continue
		}
		allIssues = append(allIssues, issues...)
	}
	fmt.Print("\033[2K\r") // Clear progress line

	// Sort by created date (newest first)
	sort.Slice(allIssues, func(i, j int) bool {
		return allIssues[i].CreatedAt.After(allIssues[j].CreatedAt)
	})

	// Print issues
	if len(allIssues) == 0 {
		fmt.Println("No open issues found.")
		return nil
	}

	fmt.Printf("\n%d open issue(s):\n\n", len(allIssues))

	for _, issue := range allIssues {
		printIssue(issue)
	}

	// Print any errors
	if len(fetchErrors) > 0 {
		fmt.Println()
		for _, err := range fetchErrors {
			fmt.Printf("%s %s\n", errorStyle.Render("Error:"), err)
		}
	}

	return nil
}

func fetchIssues(repoFullName, repoName string, limit int, assignee string) ([]GitHubIssue, error) {
	args := []string{
		"issue", "list",
		"--repo", repoFullName,
		"--state", "open",
		"--limit", fmt.Sprintf("%d", limit),
		"--json", "number,title,state,createdAt,author,assignees,labels,url",
	}

	if assignee != "" {
		args = append(args, "--assignee", assignee)
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		// Check if it's just "no issues" vs actual error
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "no issues") || strings.Contains(stderr, "Could not resolve") {
				return nil, nil
			}
			return nil, fmt.Errorf("%s", stderr)
		}
		return nil, err
	}

	var issues []GitHubIssue
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, err
	}

	// Tag with repo name
	for i := range issues {
		issues[i].RepoName = repoName
	}

	return issues, nil
}

func printIssue(issue GitHubIssue) {
	// #42 [core-bio] Fix avatar upload
	num := issueNumberStyle.Render(fmt.Sprintf("#%d", issue.Number))
	repo := issueRepoStyle.Render(fmt.Sprintf("[%s]", issue.RepoName))
	title := issueTitleStyle.Render(truncate(issue.Title, 60))

	line := fmt.Sprintf("  %s %s %s", num, repo, title)

	// Add labels if any
	if len(issue.Labels.Nodes) > 0 {
		var labels []string
		for _, l := range issue.Labels.Nodes {
			labels = append(labels, l.Name)
		}
		line += " " + issueLabelStyle.Render("["+strings.Join(labels, ", ")+"]")
	}

	// Add assignee if any
	if len(issue.Assignees.Nodes) > 0 {
		var assignees []string
		for _, a := range issue.Assignees.Nodes {
			assignees = append(assignees, "@"+a.Login)
		}
		line += " " + issueAssigneeStyle.Render(strings.Join(assignees, ", "))
	}

	// Add age
	age := formatAge(issue.CreatedAt)
	line += " " + issueAgeStyle.Render(age)

	fmt.Println(line)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatAge(t time.Time) string {
	d := time.Since(t)

	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	if d < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	}
	return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
}
