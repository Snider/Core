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
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

// PR-specific styles (aliases to shared)
var (
	prNumberStyle   = shared.PrNumberStyle
	prTitleStyle    = shared.ValueStyle
	prAuthorStyle   = shared.InfoStyle
	prApprovedStyle = shared.SuccessStyle
	prChangesStyle  = shared.WarningStyle
	prPendingStyle  = shared.DimStyle
	prDraftStyle    = shared.DimStyle
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

// Reviews command flags
var (
	reviewsRegistryPath string
	reviewsAuthor       string
	reviewsShowAll      bool
)

// addReviewsCommand adds the 'reviews' command to the given parent command.
func addReviewsCommand(parent *cobra.Command) {
	reviewsCmd := &cobra.Command{
		Use:   "reviews",
		Short: "List PRs needing review across all repos",
		Long: `Fetches open PRs from GitHub for all repos in the registry.
Shows review status (approved, changes requested, pending).
Requires the 'gh' CLI to be installed and authenticated.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReviews(reviewsRegistryPath, reviewsAuthor, reviewsShowAll)
		},
	}

	reviewsCmd.Flags().StringVar(&reviewsRegistryPath, "registry", "", "Path to repos.yaml (auto-detected if not specified)")
	reviewsCmd.Flags().StringVar(&reviewsAuthor, "author", "", "Filter by PR author")
	reviewsCmd.Flags().BoolVar(&reviewsShowAll, "all", false, "Show all PRs including drafts")

	parent.AddCommand(reviewsCmd)
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
		fmt.Printf(" * %s", prPendingStyle.Render(fmt.Sprintf("%d pending", pending)))
	}
	if approved > 0 {
		fmt.Printf(" * %s", prApprovedStyle.Render(fmt.Sprintf("%d approved", approved)))
	}
	if changesRequested > 0 {
		fmt.Printf(" * %s", prChangesStyle.Render(fmt.Sprintf("%d changes requested", changesRequested)))
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
	title := prTitleStyle.Render(shared.Truncate(pr.Title, 50))
	author := prAuthorStyle.Render("@" + pr.Author.Login)

	// Review status
	var status string
	switch pr.ReviewDecision {
	case "APPROVED":
		status = prApprovedStyle.Render("v approved")
	case "CHANGES_REQUESTED":
		status = prChangesStyle.Render("* changes requested")
	default:
		status = prPendingStyle.Render("o pending review")
	}

	// Draft indicator
	draft := ""
	if pr.IsDraft {
		draft = prDraftStyle.Render(" [draft]")
	}

	age := shared.FormatAge(pr.CreatedAt)

	fmt.Printf("  %s %s %s%s %s  %s  %s\n", num, repo, title, draft, author, status, issueAgeStyle.Render(age))
}
