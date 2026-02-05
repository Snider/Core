// cmd_issues.go implements the 'qa issues' command for intelligent issue triage.
//
// Usage:
//   core qa issues              # Show prioritised, actionable issues
//   core qa issues --mine       # Show issues assigned to you
//   core qa issues --triage     # Show issues needing triage (no labels/assignee)
//   core qa issues --blocked    # Show blocked issues

package qa

import (
	"encoding/json"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/log"
	"github.com/host-uk/core/pkg/repos"
)

// Issue command flags
var (
	issuesMine     bool
	issuesTriage   bool
	issuesBlocked  bool
	issuesRegistry string
	issuesLimit    int
)

// Issue represents a GitHub issue with triage metadata
type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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
	Comments struct {
		TotalCount int `json:"totalCount"`
		Nodes      []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			CreatedAt time.Time `json:"createdAt"`
		} `json:"nodes"`
	} `json:"comments"`
	URL string `json:"url"`

	// Computed fields
	RepoName   string
	Priority   int    // Lower = higher priority
	Category   string // "needs_response", "ready", "blocked", "triage"
	ActionHint string
}

// addIssuesCommand adds the 'issues' subcommand to qa.
func addIssuesCommand(parent *cli.Command) {
	issuesCmd := &cli.Command{
		Use:   "issues",
		Short: i18n.T("cmd.qa.issues.short"),
		Long:  i18n.T("cmd.qa.issues.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runQAIssues()
		},
	}

	issuesCmd.Flags().BoolVarP(&issuesMine, "mine", "m", false, i18n.T("cmd.qa.issues.flag.mine"))
	issuesCmd.Flags().BoolVarP(&issuesTriage, "triage", "t", false, i18n.T("cmd.qa.issues.flag.triage"))
	issuesCmd.Flags().BoolVarP(&issuesBlocked, "blocked", "b", false, i18n.T("cmd.qa.issues.flag.blocked"))
	issuesCmd.Flags().StringVar(&issuesRegistry, "registry", "", i18n.T("common.flag.registry"))
	issuesCmd.Flags().IntVarP(&issuesLimit, "limit", "l", 50, i18n.T("cmd.qa.issues.flag.limit"))

	parent.AddCommand(issuesCmd)
}

func runQAIssues() error {
	// Check gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return log.E("qa.issues", i18n.T("error.gh_not_found"), nil)
	}

	// Load registry
	var reg *repos.Registry
	var err error

	if issuesRegistry != "" {
		reg, err = repos.LoadRegistry(io.Local, issuesRegistry)
	} else {
		registryPath, findErr := repos.FindRegistry(io.Local)
		if findErr != nil {
			return log.E("qa.issues", i18n.T("error.registry_not_found"), nil)
		}
		reg, err = repos.LoadRegistry(io.Local, registryPath)
	}
	if err != nil {
		return log.E("qa.issues", "failed to load registry", err)
	}

	// Fetch issues from all repos
	var allIssues []Issue
	repoList := reg.List()

	for i, repo := range repoList {
		cli.Print("\033[2K\r%s %d/%d %s",
			dimStyle.Render(i18n.T("cmd.qa.issues.fetching")),
			i+1, len(repoList), repo.Name)

		issues, err := fetchQAIssues(reg.Org, repo.Name, issuesLimit)
		if err != nil {
			continue // Skip repos with errors
		}
		allIssues = append(allIssues, issues...)
	}
	cli.Print("\033[2K\r") // Clear progress

	if len(allIssues) == 0 {
		cli.Text(i18n.T("cmd.qa.issues.no_issues"))
		return nil
	}

	// Categorise and prioritise issues
	categorised := categoriseIssues(allIssues)

	// Filter based on flags
	if issuesMine {
		categorised = filterMine(categorised)
	}
	if issuesTriage {
		categorised = filterCategory(categorised, "triage")
	}
	if issuesBlocked {
		categorised = filterCategory(categorised, "blocked")
	}

	// Print categorised issues
	printCategorisedIssues(categorised)

	return nil
}

func fetchQAIssues(org, repoName string, limit int) ([]Issue, error) {
	repoFullName := cli.Sprintf("%s/%s", org, repoName)

	args := []string{
		"issue", "list",
		"--repo", repoFullName,
		"--state", "open",
		"--limit", cli.Sprintf("%d", limit),
		"--json", "number,title,state,body,createdAt,updatedAt,author,assignees,labels,comments,url",
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var issues []Issue
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, err
	}

	// Tag with repo name
	for i := range issues {
		issues[i].RepoName = repoName
	}

	return issues, nil
}

func categoriseIssues(issues []Issue) map[string][]Issue {
	result := map[string][]Issue{
		"needs_response": {},
		"ready":          {},
		"blocked":        {},
		"triage":         {},
	}

	currentUser := getCurrentUser()

	for i := range issues {
		issue := &issues[i]
		categoriseIssue(issue, currentUser)
		result[issue.Category] = append(result[issue.Category], *issue)
	}

	// Sort each category by priority
	for cat := range result {
		sort.Slice(result[cat], func(i, j int) bool {
			return result[cat][i].Priority < result[cat][j].Priority
		})
	}

	return result
}

func categoriseIssue(issue *Issue, currentUser string) {
	labels := getLabels(issue)

	// Check if blocked
	for _, l := range labels {
		if strings.HasPrefix(l, "blocked") || l == "waiting" {
			issue.Category = "blocked"
			issue.Priority = 30
			issue.ActionHint = i18n.T("cmd.qa.issues.hint.blocked")
			return
		}
	}

	// Check if needs triage (no labels, no assignee)
	if len(issue.Labels.Nodes) == 0 && len(issue.Assignees.Nodes) == 0 {
		issue.Category = "triage"
		issue.Priority = 20
		issue.ActionHint = i18n.T("cmd.qa.issues.hint.triage")
		return
	}

	// Check if needs response (recent comment from someone else)
	if issue.Comments.TotalCount > 0 && len(issue.Comments.Nodes) > 0 {
		lastComment := issue.Comments.Nodes[len(issue.Comments.Nodes)-1]
		// If last comment is not from current user and is recent
		if lastComment.Author.Login != currentUser {
			age := time.Since(lastComment.CreatedAt)
			if age < 48*time.Hour {
				issue.Category = "needs_response"
				issue.Priority = 10
				issue.ActionHint = cli.Sprintf("@%s %s", lastComment.Author.Login, i18n.T("cmd.qa.issues.hint.needs_response"))
				return
			}
		}
	}

	// Default: ready to work
	issue.Category = "ready"
	issue.Priority = calculatePriority(issue, labels)
	issue.ActionHint = ""
}

func calculatePriority(issue *Issue, labels []string) int {
	priority := 50

	// Priority labels
	for _, l := range labels {
		switch {
		case strings.Contains(l, "critical") || strings.Contains(l, "urgent"):
			priority = 1
		case strings.Contains(l, "high"):
			priority = 10
		case strings.Contains(l, "medium"):
			priority = 30
		case strings.Contains(l, "low"):
			priority = 70
		case l == "good-first-issue" || l == "good first issue":
			priority = min(priority, 15) // Boost good first issues
		case l == "help-wanted" || l == "help wanted":
			priority = min(priority, 20)
		case l == "agent:ready" || l == "agentic":
			priority = min(priority, 5) // AI-ready issues are high priority
		}
	}

	return priority
}

func getLabels(issue *Issue) []string {
	var labels []string
	for _, l := range issue.Labels.Nodes {
		labels = append(labels, strings.ToLower(l.Name))
	}
	return labels
}

func getCurrentUser() string {
	cmd := exec.Command("gh", "api", "user", "--jq", ".login")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func filterMine(categorised map[string][]Issue) map[string][]Issue {
	currentUser := getCurrentUser()
	result := make(map[string][]Issue)

	for cat, issues := range categorised {
		var filtered []Issue
		for _, issue := range issues {
			for _, a := range issue.Assignees.Nodes {
				if a.Login == currentUser {
					filtered = append(filtered, issue)
					break
				}
			}
		}
		if len(filtered) > 0 {
			result[cat] = filtered
		}
	}

	return result
}

func filterCategory(categorised map[string][]Issue, category string) map[string][]Issue {
	if issues, ok := categorised[category]; ok && len(issues) > 0 {
		return map[string][]Issue{category: issues}
	}
	return map[string][]Issue{}
}

func printCategorisedIssues(categorised map[string][]Issue) {
	// Print in order: needs_response, ready, blocked, triage
	categories := []struct {
		key   string
		title string
		style *cli.AnsiStyle
	}{
		{"needs_response", i18n.T("cmd.qa.issues.category.needs_response"), warningStyle},
		{"ready", i18n.T("cmd.qa.issues.category.ready"), successStyle},
		{"blocked", i18n.T("cmd.qa.issues.category.blocked"), errorStyle},
		{"triage", i18n.T("cmd.qa.issues.category.triage"), dimStyle},
	}

	first := true
	for _, cat := range categories {
		issues := categorised[cat.key]
		if len(issues) == 0 {
			continue
		}

		if !first {
			cli.Blank()
		}
		first = false

		cli.Print("%s (%d):\n", cat.style.Render(cat.title), len(issues))

		for _, issue := range issues {
			printTriagedIssue(issue)
		}
	}

	if first {
		cli.Text(i18n.T("cmd.qa.issues.no_issues"))
	}
}

func printTriagedIssue(issue Issue) {
	// #42 [core-bio] Fix avatar upload
	num := cli.TitleStyle.Render(cli.Sprintf("#%d", issue.Number))
	repo := dimStyle.Render(cli.Sprintf("[%s]", issue.RepoName))
	title := cli.ValueStyle.Render(truncate(issue.Title, 50))

	cli.Print("  %s %s %s", num, repo, title)

	// Add labels if priority-related
	var importantLabels []string
	for _, l := range issue.Labels.Nodes {
		name := strings.ToLower(l.Name)
		if strings.Contains(name, "priority") || strings.Contains(name, "critical") ||
			name == "good-first-issue" || name == "agent:ready" || name == "agentic" {
			importantLabels = append(importantLabels, l.Name)
		}
	}
	if len(importantLabels) > 0 {
		cli.Print(" %s", warningStyle.Render("["+strings.Join(importantLabels, ", ")+"]"))
	}

	// Add age
	age := cli.FormatAge(issue.UpdatedAt)
	cli.Print(" %s\n", dimStyle.Render(age))

	// Add action hint if present
	if issue.ActionHint != "" {
		cli.Print("      %s %s\n", dimStyle.Render("->"), issue.ActionHint)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
