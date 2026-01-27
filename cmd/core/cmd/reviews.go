package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/Snider/Core/pkg/repos"
	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
)

var (
	prNumberStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#a855f7")) // purple-500

	prTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")) // gray-200

	prAuthorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")) // blue-500

	prApprovedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22c55e")) // green-500

	prChangesStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f59e0b")) // amber-500

	prPendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")) // gray-500

	prDraftStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")) // gray-500
)

// GitHubPR represents a GitHub pull request.
type GitHubPR struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	IsDraft   bool      `json:"isDraft"`
	CreatedAt time.Time `json:"createdAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	ReviewDecision string `json:"reviewDecision"`
	Reviews        struct {
		Nodes []struct {
			State  string `json:"state"`
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
		} `json:"nodes"`
	} `json:"reviews"`
	URL string `json:"url"`

	// Added by us
	RepoName string `json:"-"`
}

// AddReviewsCommand adds the 'reviews' command to the given parent command.
func AddReviewsCommand(parent *clir.Cli) {
	var registryPath string
	var author string
	var showAll bool

	reviewsCmd := parent.NewSubCommand("reviews", "List PRs needing review across all repos")
	reviewsCmd.LongDescription("Fetches open PRs from GitHub for all repos in the registry.\n" +
		"Shows review status (approved, changes requested, pending).\n" +
		"Requires the 'gh' CLI to be installed and authenticated.")

	reviewsCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)
	reviewsCmd.StringFlag("author", "Filter by PR author", &author)
	reviewsCmd.BoolFlag("all", "Show all PRs including drafts", &showAll)

	reviewsCmd.Action(func() error {
		return runReviews(registryPath, author, showAll)
	})
}

func runReviews(registryPath string, author string, showAll bool) error {
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

	// Fetch PRs sequentially (avoid GitHub rate limits)
	var allPRs []GitHubPR
	var fetchErrors []error

	repoList := reg.List()
	for i, repo := range repoList {
		repoFullName := fmt.Sprintf("%s/%s", reg.Org, repo.Name)
		fmt.Printf("\033[2K\r%s %d/%d %s", dimStyle.Render("Fetching"), i+1, len(repoList), repo.Name)

		prs, err := fetchPRs(repoFullName, repo.Name, author)
		if err != nil {
			fetchErrors = append(fetchErrors, fmt.Errorf("%s: %w", repo.Name, err))
			continue
		}

		for _, pr := range prs {
			// Filter drafts unless --all
			if !showAll && pr.IsDraft {
				continue
			}
			allPRs = append(allPRs, pr)
		}
	}
	fmt.Print("\033[2K\r") // Clear progress line

	// Sort: pending review first, then by date
	sort.Slice(allPRs, func(i, j int) bool {
		// Pending reviews come first
		iPending := allPRs[i].ReviewDecision == "" || allPRs[i].ReviewDecision == "REVIEW_REQUIRED"
		jPending := allPRs[j].ReviewDecision == "" || allPRs[j].ReviewDecision == "REVIEW_REQUIRED"
		if iPending != jPending {
			return iPending
		}
		return allPRs[i].CreatedAt.After(allPRs[j].CreatedAt)
	})

	// Print PRs
	if len(allPRs) == 0 {
		fmt.Println("No open PRs found.")
		return nil
	}

	// Count by status
	var pending, approved, changesRequested int
	for _, pr := range allPRs {
		switch pr.ReviewDecision {
		case "APPROVED":
			approved++
		case "CHANGES_REQUESTED":
			changesRequested++
		default:
			pending++
		}
	}

	fmt.Println()
	fmt.Printf("%d open PR(s)", len(allPRs))
	if pending > 0 {
		fmt.Printf(" · %s", prPendingStyle.Render(fmt.Sprintf("%d pending", pending)))
	}
	if approved > 0 {
		fmt.Printf(" · %s", prApprovedStyle.Render(fmt.Sprintf("%d approved", approved)))
	}
	if changesRequested > 0 {
		fmt.Printf(" · %s", prChangesStyle.Render(fmt.Sprintf("%d changes requested", changesRequested)))
	}
	fmt.Println()
	fmt.Println()

	for _, pr := range allPRs {
		printPR(pr)
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

func fetchPRs(repoFullName, repoName string, author string) ([]GitHubPR, error) {
	args := []string{
		"pr", "list",
		"--repo", repoFullName,
		"--state", "open",
		"--json", "number,title,state,isDraft,createdAt,author,reviewDecision,reviews,url",
	}

	if author != "" {
		args = append(args, "--author", author)
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "no pull requests") || strings.Contains(stderr, "Could not resolve") {
				return nil, nil
			}
			return nil, fmt.Errorf("%s", stderr)
		}
		return nil, err
	}

	var prs []GitHubPR
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, err
	}

	// Tag with repo name
	for i := range prs {
		prs[i].RepoName = repoName
	}

	return prs, nil
}

func printPR(pr GitHubPR) {
	// #12 [core-php] Webhook validation
	num := prNumberStyle.Render(fmt.Sprintf("#%d", pr.Number))
	repo := issueRepoStyle.Render(fmt.Sprintf("[%s]", pr.RepoName))
	title := prTitleStyle.Render(truncate(pr.Title, 50))
	author := prAuthorStyle.Render("@" + pr.Author.Login)

	// Review status
	var status string
	switch pr.ReviewDecision {
	case "APPROVED":
		status = prApprovedStyle.Render("✓ approved")
	case "CHANGES_REQUESTED":
		status = prChangesStyle.Render("● changes requested")
	default:
		status = prPendingStyle.Render("○ pending review")
	}

	// Draft indicator
	draft := ""
	if pr.IsDraft {
		draft = prDraftStyle.Render(" [draft]")
	}

	age := formatAge(pr.CreatedAt)

	fmt.Printf("  %s %s %s%s %s  %s  %s\n", num, repo, title, draft, author, status, issueAgeStyle.Render(age))
}
