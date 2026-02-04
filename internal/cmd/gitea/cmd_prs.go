package gitea

import (
	"fmt"
	"strings"

	sdk "code.gitea.io/sdk/gitea"

	"github.com/host-uk/core/pkg/cli"
	gt "github.com/host-uk/core/pkg/gitea"
)

// PRs command flags.
var (
	prsState string
)

// addPRsCommand adds the 'prs' subcommand for listing pull requests.
func addPRsCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "prs <owner/repo>",
		Short: "List pull requests",
		Long:  "List pull requests for a repository.",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			owner, repo, err := splitOwnerRepo(args[0])
			if err != nil {
				return err
			}
			return runListPRs(owner, repo)
		},
	}

	cmd.Flags().StringVar(&prsState, "state", "open", "Filter by state (open, closed, all)")

	parent.AddCommand(cmd)
}

func runListPRs(owner, repo string) error {
	client, err := gt.NewFromConfig("", "")
	if err != nil {
		return err
	}

	prs, err := client.ListPullRequests(owner, repo, prsState)
	if err != nil {
		return err
	}

	if len(prs) == 0 {
		cli.Text(fmt.Sprintf("No %s pull requests in %s/%s.", prsState, owner, repo))
		return nil
	}

	cli.Blank()
	cli.Print("  %s\n\n", fmt.Sprintf("%d %s pull requests in %s/%s", len(prs), prsState, owner, repo))

	for _, pr := range prs {
		printGiteaPR(pr)
	}

	return nil
}

func printGiteaPR(pr *sdk.PullRequest) {
	num := numberStyle.Render(fmt.Sprintf("#%d", pr.Index))
	title := valueStyle.Render(cli.Truncate(pr.Title, 50))

	var author string
	if pr.Poster != nil {
		author = infoStyle.Render("@" + pr.Poster.UserName)
	}

	// Branch info
	branch := dimStyle.Render(pr.Head.Ref + " -> " + pr.Base.Ref)

	// Merge status
	var status string
	if pr.HasMerged {
		status = successStyle.Render("merged")
	} else if pr.State == sdk.StateClosed {
		status = errorStyle.Render("closed")
	} else {
		status = warningStyle.Render("open")
	}

	// Labels
	var labelStr string
	if len(pr.Labels) > 0 {
		var labels []string
		for _, l := range pr.Labels {
			labels = append(labels, l.Name)
		}
		labelStr = " " + warningStyle.Render("["+strings.Join(labels, ", ")+"]")
	}

	cli.Print("  %s %s %s %s  %s%s\n", num, title, author, status, branch, labelStr)
}
