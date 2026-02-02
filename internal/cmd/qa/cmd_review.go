// cmd_review.go implements the 'qa review' command for PR review status.
//
// Usage:
//   core qa review              # Show all PRs needing attention
//   core qa review --mine       # Show status of your open PRs
//   core qa review --requested  # Show PRs you need to review

package qa

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/log"
)

// Review command flags
var (
	reviewMine      bool
	reviewRequested bool
	reviewRepo      string
)

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number         int                `json:"number"`
	Title          string             `json:"title"`
	Author         Author             `json:"author"`
	State          string             `json:"state"`
	IsDraft        bool               `json:"isDraft"`
	Mergeable      string             `json:"mergeable"`
	ReviewDecision string             `json:"reviewDecision"`
	URL            string             `json:"url"`
	HeadRefName    string             `json:"headRefName"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	Additions      int                `json:"additions"`
	Deletions      int                `json:"deletions"`
	ChangedFiles   int                `json:"changedFiles"`
	StatusChecks   *StatusCheckRollup `json:"statusCheckRollup"`
	ReviewRequests ReviewRequests     `json:"reviewRequests"`
	Reviews        []Review           `json:"reviews"`
}

// Author represents a GitHub user
type Author struct {
	Login string `json:"login"`
}

// StatusCheckRollup contains CI check status
type StatusCheckRollup struct {
	Contexts []StatusContext `json:"contexts"`
}

// StatusContext represents a single check
type StatusContext struct {
	State      string `json:"state"`
	Conclusion string `json:"conclusion"`
	Name       string `json:"name"`
}

// ReviewRequests contains pending review requests
type ReviewRequests struct {
	Nodes []ReviewRequest `json:"nodes"`
}

// ReviewRequest represents a review request
type ReviewRequest struct {
	RequestedReviewer Author `json:"requestedReviewer"`
}

// Review represents a PR review
type Review struct {
	Author Author `json:"author"`
	State  string `json:"state"`
}

// addReviewCommand adds the 'review' subcommand to the qa command.
func addReviewCommand(parent *cli.Command) {
	reviewCmd := &cli.Command{
		Use:   "review",
		Short: i18n.T("cmd.qa.review.short"),
		Long:  i18n.T("cmd.qa.review.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runReview()
		},
	}

	reviewCmd.Flags().BoolVarP(&reviewMine, "mine", "m", false, i18n.T("cmd.qa.review.flag.mine"))
	reviewCmd.Flags().BoolVarP(&reviewRequested, "requested", "r", false, i18n.T("cmd.qa.review.flag.requested"))
	reviewCmd.Flags().StringVar(&reviewRepo, "repo", "", i18n.T("cmd.qa.review.flag.repo"))

	parent.AddCommand(reviewCmd)
}

func runReview() error {
	// Check gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return log.E("qa.review", i18n.T("error.gh_not_found"), nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Determine repo
	repoFullName := reviewRepo
	if repoFullName == "" {
		var err error
		repoFullName, err = detectRepoFromGit()
		if err != nil {
			return log.E("qa.review", i18n.T("cmd.qa.review.error.no_repo"), nil)
		}
	}

	// Default: show both mine and requested if neither flag is set
	showMine := reviewMine || (!reviewMine && !reviewRequested)
	showRequested := reviewRequested || (!reviewMine && !reviewRequested)

	if showMine {
		if err := showMyPRs(ctx, repoFullName); err != nil {
			return err
		}
	}

	if showRequested {
		if showMine {
			cli.Blank()
		}
		if err := showRequestedReviews(ctx, repoFullName); err != nil {
			return err
		}
	}

	return nil
}

// showMyPRs shows the user's open PRs with status
func showMyPRs(ctx context.Context, repo string) error {
	prs, err := fetchPRs(ctx, repo, "author:@me")
	if err != nil {
		return log.E("qa.review", "failed to fetch your PRs", err)
	}

	if len(prs) == 0 {
		cli.Print("%s\n", dimStyle.Render(i18n.T("cmd.qa.review.no_prs")))
		return nil
	}

	cli.Print("%s (%d):\n", i18n.T("cmd.qa.review.your_prs"), len(prs))

	for _, pr := range prs {
		printPRStatus(pr)
	}

	return nil
}

// showRequestedReviews shows PRs where user's review is requested
func showRequestedReviews(ctx context.Context, repo string) error {
	prs, err := fetchPRs(ctx, repo, "review-requested:@me")
	if err != nil {
		return log.E("qa.review", "failed to fetch review requests", err)
	}

	if len(prs) == 0 {
		cli.Print("%s\n", dimStyle.Render(i18n.T("cmd.qa.review.no_reviews")))
		return nil
	}

	cli.Print("%s (%d):\n", i18n.T("cmd.qa.review.review_requested"), len(prs))

	for _, pr := range prs {
		printPRForReview(pr)
	}

	return nil
}

// fetchPRs fetches PRs matching the search query
func fetchPRs(ctx context.Context, repo, search string) ([]PullRequest, error) {
	args := []string{
		"pr", "list",
		"--state", "open",
		"--search", search,
		"--json", "number,title,author,state,isDraft,mergeable,reviewDecision,url,headRefName,createdAt,updatedAt,additions,deletions,changedFiles,statusCheckRollup,reviewRequests,reviews",
	}

	if repo != "" {
		args = append(args, "--repo", repo)
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}

	var prs []PullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, err
	}

	return prs, nil
}

// printPRStatus prints a PR with its merge status
func printPRStatus(pr PullRequest) {
	// Determine status icon and color
	status, style, action := analyzePRStatus(pr)

	cli.Print("  %s #%d %s\n",
		style.Render(status),
		pr.Number,
		truncate(pr.Title, 50))

	if action != "" {
		cli.Print("      %s %s\n", dimStyle.Render("->"), action)
	}
}

// printPRForReview prints a PR that needs review
func printPRForReview(pr PullRequest) {
	// Show PR info with stats
	stats := fmt.Sprintf("+%d/-%d, %d files",
		pr.Additions, pr.Deletions, pr.ChangedFiles)

	cli.Print("  %s #%d %s\n",
		warningStyle.Render("◯"),
		pr.Number,
		truncate(pr.Title, 50))
	cli.Print("      %s @%s, %s\n",
		dimStyle.Render("->"),
		pr.Author.Login,
		stats)
	cli.Print("      %s gh pr checkout %d\n",
		dimStyle.Render("->"),
		pr.Number)
}

// analyzePRStatus determines the status, style, and action for a PR
func analyzePRStatus(pr PullRequest) (status string, style *cli.AnsiStyle, action string) {
	// Check if draft
	if pr.IsDraft {
		return "◯", dimStyle, "Draft - convert to ready when done"
	}

	// Check CI status
	ciPassed := true
	ciFailed := false
	ciPending := false
	var failedCheck string

	if pr.StatusChecks != nil {
		for _, check := range pr.StatusChecks.Contexts {
			switch check.Conclusion {
			case "FAILURE", "failure":
				ciFailed = true
				ciPassed = false
				if failedCheck == "" {
					failedCheck = check.Name
				}
			case "PENDING", "pending", "":
				if check.State == "PENDING" || check.State == "" {
					ciPending = true
					ciPassed = false
				}
			}
		}
	}

	// Check review status
	approved := pr.ReviewDecision == "APPROVED"
	changesRequested := pr.ReviewDecision == "CHANGES_REQUESTED"

	// Check mergeable status
	hasConflicts := pr.Mergeable == "CONFLICTING"

	// Determine overall status
	if hasConflicts {
		return "✗", errorStyle, "Needs rebase - has merge conflicts"
	}

	if ciFailed {
		return "✗", errorStyle, fmt.Sprintf("CI failed: %s", failedCheck)
	}

	if changesRequested {
		return "✗", warningStyle, "Changes requested - address review feedback"
	}

	if ciPending {
		return "◯", warningStyle, "CI running..."
	}

	if !approved && pr.ReviewDecision != "" {
		return "◯", warningStyle, "Awaiting review"
	}

	if approved && ciPassed {
		return "✓", successStyle, "Ready to merge"
	}

	return "◯", dimStyle, ""
}

// truncate shortens a string to max length (rune-safe for UTF-8)
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}
