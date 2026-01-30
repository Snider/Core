package dev

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

// Issue-specific styles (aliases to shared)
var (
	issueRepoStyle     = shared.DimStyle
	issueNumberStyle   = shared.TitleStyle
	issueTitleStyle    = shared.ValueStyle
	issueLabelStyle    = shared.WarningStyle
	issueAssigneeStyle = shared.StatusSuccessStyle
	issueAgeStyle      = shared.DimStyle
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

// Issues command flags
var (
	issuesRegistryPath string
	issuesLimit        int
	issuesAssignee     string
)

// addIssuesCommand adds the 'issues' command to the given parent command.
func addIssuesCommand(parent *cobra.Command) {
	issuesCmd := &cobra.Command{
		Use:   "issues",
		Short: i18n.T("cmd.dev.issues.short"),
		Long:  i18n.T("cmd.dev.issues.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			limit := issuesLimit
			if limit == 0 {
				limit = 10
			}
			return runIssues(issuesRegistryPath, limit, issuesAssignee)
		},
	}

	issuesCmd.Flags().StringVar(&issuesRegistryPath, "registry", "", i18n.T("cmd.dev.issues.flag.registry"))
	issuesCmd.Flags().IntVarP(&issuesLimit, "limit", "l", 10, i18n.T("cmd.dev.issues.flag.limit"))
	issuesCmd.Flags().StringVarP(&issuesAssignee, "assignee", "a", "", i18n.T("cmd.dev.issues.flag.assignee"))

	parent.AddCommand(issuesCmd)
}

func runIssues(registryPath string, limit int, assignee string) error {
	// Check gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf(i18n.T("error.gh_not_found"))
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
		fmt.Printf("\033[2K\r%s %d/%d %s", dimStyle.Render(i18n.T("cli.progress.fetching")), i+1, len(repoList), repo.Name)

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
		fmt.Println(i18n.T("cmd.dev.issues.no_issues"))
		return nil
	}

	fmt.Printf("\n%s\n\n", i18n.T("cmd.dev.issues.open_issues", map[string]interface{}{"Count": len(allIssues)}))

	for _, issue := range allIssues {
		printIssue(issue)
	}

	// Print any errors
	if len(fetchErrors) > 0 {
		fmt.Println()
		for _, err := range fetchErrors {
			fmt.Printf("%s %s\n", errorStyle.Render(i18n.T("cmd.dev.issues.error_label")), err)
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
	title := issueTitleStyle.Render(shared.Truncate(issue.Title, 60))

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
	age := shared.FormatAge(issue.CreatedAt)
	line += " " + issueAgeStyle.Render(age)

	fmt.Println(line)
}
