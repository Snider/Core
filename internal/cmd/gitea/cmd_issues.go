package gitea

import (
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"

	"github.com/host-uk/core/pkg/cli"
	gt "github.com/host-uk/core/pkg/gitea"
)

// Issues command flags.
var (
	issuesState string
	issuesTitle string
	issuesBody  string
)

// addIssuesCommand adds the 'issues' subcommand for listing and creating issues.
func addIssuesCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "issues <owner/repo>",
		Short: "List and manage issues",
		Long:  "List issues for a repository, or create a new issue.",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			owner, repo, err := splitOwnerRepo(args[0])
			if err != nil {
				return err
			}

			// If title is set, create an issue instead
			if issuesTitle != "" {
				return runCreateIssue(owner, repo)
			}

			return runListIssues(owner, repo)
		},
	}

	cmd.Flags().StringVar(&issuesState, "state", "open", "Filter by state (open, closed, all)")
	cmd.Flags().StringVar(&issuesTitle, "title", "", "Create issue with this title")
	cmd.Flags().StringVar(&issuesBody, "body", "", "Issue body (used with --title)")

	parent.AddCommand(cmd)
}

func runListIssues(owner, repo string) error {
	client, err := gt.NewFromConfig("", "")
	if err != nil {
		return err
	}

	issues, err := client.ListIssues(owner, repo, gt.ListIssuesOpts{
		State: issuesState,
	})
	if err != nil {
		return err
	}

	if len(issues) == 0 {
		cli.Text(fmt.Sprintf("No %s issues in %s/%s.", issuesState, owner, repo))
		return nil
	}

	cli.Blank()
	cli.Print("  %s\n\n", fmt.Sprintf("%d %s issues in %s/%s", len(issues), issuesState, owner, repo))

	for _, issue := range issues {
		printGiteaIssue(issue, owner, repo)
	}

	return nil
}

func runCreateIssue(owner, repo string) error {
	client, err := gt.NewFromConfig("", "")
	if err != nil {
		return err
	}

	issue, err := client.CreateIssue(owner, repo, gitea.CreateIssueOption{
		Title: issuesTitle,
		Body:  issuesBody,
	})
	if err != nil {
		return err
	}

	cli.Blank()
	cli.Success(fmt.Sprintf("Created issue #%d: %s", issue.Index, issue.Title))
	cli.Print("  %s %s\n", dimStyle.Render("URL:"), valueStyle.Render(issue.HTMLURL))
	cli.Blank()

	return nil
}

func printGiteaIssue(issue *gitea.Issue, owner, repo string) {
	num := numberStyle.Render(fmt.Sprintf("#%d", issue.Index))
	title := valueStyle.Render(cli.Truncate(issue.Title, 60))

	line := fmt.Sprintf("  %s %s", num, title)

	// Add labels
	if len(issue.Labels) > 0 {
		var labels []string
		for _, l := range issue.Labels {
			labels = append(labels, l.Name)
		}
		line += " " + warningStyle.Render("["+strings.Join(labels, ", ")+"]")
	}

	// Add assignees
	if len(issue.Assignees) > 0 {
		var assignees []string
		for _, a := range issue.Assignees {
			assignees = append(assignees, "@"+a.UserName)
		}
		line += " " + infoStyle.Render(strings.Join(assignees, ", "))
	}

	cli.Text(line)
}

// splitOwnerRepo splits "owner/repo" into its parts.
func splitOwnerRepo(s string) (string, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", cli.Err("expected format: owner/repo (got %q)", s)
	}
	return parts[0], parts[1], nil
}
