package dev

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
)

// PR-specific styles (aliases to shared)
var (
	prNumberStyle   = cli.PrNumberStyle
	prTitleStyle    = cli.ValueStyle
	prAuthorStyle   = cli.InfoStyle
	prApprovedStyle = cli.SuccessStyle
	prChangesStyle  = cli.WarningStyle
	prPendingStyle  = cli.DimStyle
	prDraftStyle    = cli.DimStyle
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
func addReviewsCommand(parent *cli.Command) {
	reviewsCmd := &cli.Command{
		Use:   "reviews",
		Short: i18n.T("cmd.dev.reviews.short"),
		Long:  i18n.T("cmd.dev.reviews.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runReviews(reviewsRegistryPath, reviewsAuthor, reviewsShowAll)
		},
	}

	reviewsCmd.Flags().StringVar(&reviewsRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	reviewsCmd.Flags().StringVar(&reviewsAuthor, "author", "", i18n.T("cmd.dev.reviews.flag.author"))
	reviewsCmd.Flags().BoolVar(&reviewsShowAll, "all", false, i18n.T("cmd.dev.reviews.flag.all"))

	parent.AddCommand(reviewsCmd)
}

func runReviews(registryPath string, author string, showAll bool) error {
	// Check gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return errors.New(i18n.T("error.gh_not_found"))
	}

	// Find or use provided registry, fall back to directory scan
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
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return cli.Wrap(err, "failed to scan directory")
			}
		}
	}

	// Fetch PRs sequentially (avoid GitHub rate limits)
	var allPRs []GitHubPR
	var fetchErrors []error

	repoList := reg.List()
	for i, repo := range repoList {
		repoFullName := cli.Sprintf("%s/%s", reg.Org, repo.Name)
		cli.Print("\033[2K\r%s %d/%d %s", dimStyle.Render(i18n.T("i18n.progress.fetch")), i+1, len(repoList), repo.Name)

		prs, err := fetchPRs(repoFullName, repo.Name, author)
		if err != nil {
			fetchErrors = append(fetchErrors, cli.Wrap(err, repo.Name))
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
	cli.Print("\033[2K\r") // Clear progress line

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
		cli.Text(i18n.T("cmd.dev.reviews.no_prs"))
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

	cli.Line("")
	cli.Print("%s", i18n.T("cmd.dev.reviews.open_prs", map[string]interface{}{"Count": len(allPRs)}))
	if pending > 0 {
		cli.Print(" * %s", prPendingStyle.Render(i18n.T("common.count.pending", map[string]interface{}{"Count": pending})))
	}
	if approved > 0 {
		cli.Print(" * %s", prApprovedStyle.Render(i18n.T("cmd.dev.reviews.approved", map[string]interface{}{"Count": approved})))
	}
	if changesRequested > 0 {
		cli.Print(" * %s", prChangesStyle.Render(i18n.T("cmd.dev.reviews.changes_requested", map[string]interface{}{"Count": changesRequested})))
	}
	cli.Line("")
	cli.Line("")

	for _, pr := range allPRs {
		printPR(pr)
	}

	// Print any errors
	if len(fetchErrors) > 0 {
		cli.Line("")
		for _, err := range fetchErrors {
			cli.Print("%s %s\n", errorStyle.Render(i18n.Label("error")), err)
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
			return nil, cli.Err("%s", stderr)
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
	num := prNumberStyle.Render(cli.Sprintf("#%d", pr.Number))
	repo := issueRepoStyle.Render(cli.Sprintf("[%s]", pr.RepoName))
	title := prTitleStyle.Render(cli.Truncate(pr.Title, 50))
	author := prAuthorStyle.Render("@" + pr.Author.Login)

	// Review status
	var status string
	switch pr.ReviewDecision {
	case "APPROVED":
		status = prApprovedStyle.Render(i18n.T("cmd.dev.reviews.status_approved"))
	case "CHANGES_REQUESTED":
		status = prChangesStyle.Render(i18n.T("cmd.dev.reviews.status_changes"))
	default:
		status = prPendingStyle.Render(i18n.T("cmd.dev.reviews.status_pending"))
	}

	// Draft indicator
	draft := ""
	if pr.IsDraft {
		draft = prDraftStyle.Render(" " + i18n.T("cmd.dev.reviews.draft"))
	}

	age := cli.FormatAge(pr.CreatedAt)

	cli.Print("  %s %s %s%s %s  %s  %s\n", num, repo, title, draft, author, status, issueAgeStyle.Render(age))
}
