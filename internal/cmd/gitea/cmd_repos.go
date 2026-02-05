package gitea

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	gt "github.com/host-uk/core/pkg/gitea"
)

// Repos command flags.
var (
	reposOrg     string
	reposMirrors bool
)

// addReposCommand adds the 'repos' subcommand for listing repositories.
func addReposCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "repos",
		Short: "List repositories",
		Long:  "List repositories from your Gitea instance, optionally filtered by organisation or mirror status.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runRepos()
		},
	}

	cmd.Flags().StringVar(&reposOrg, "org", "", "Filter by organisation")
	cmd.Flags().BoolVar(&reposMirrors, "mirrors", false, "Show only mirror repositories")

	parent.AddCommand(cmd)
}

func runRepos() error {
	client, err := gt.NewFromConfig("", "")
	if err != nil {
		return err
	}

	var repos []*giteaRepo
	if reposOrg != "" {
		raw, err := client.ListOrgRepos(reposOrg)
		if err != nil {
			return err
		}
		for _, r := range raw {
			repos = append(repos, &giteaRepo{
				Name:     r.Name,
				FullName: r.FullName,
				Mirror:   r.Mirror,
				Private:  r.Private,
				Stars:    r.Stars,
				CloneURL: r.CloneURL,
			})
		}
	} else {
		raw, err := client.ListUserRepos()
		if err != nil {
			return err
		}
		for _, r := range raw {
			repos = append(repos, &giteaRepo{
				Name:     r.Name,
				FullName: r.FullName,
				Mirror:   r.Mirror,
				Private:  r.Private,
				Stars:    r.Stars,
				CloneURL: r.CloneURL,
			})
		}
	}

	// Filter mirrors if requested
	if reposMirrors {
		var filtered []*giteaRepo
		for _, r := range repos {
			if r.Mirror {
				filtered = append(filtered, r)
			}
		}
		repos = filtered
	}

	if len(repos) == 0 {
		cli.Text("No repositories found.")
		return nil
	}

	// Build table
	table := cli.NewTable("Name", "Type", "Visibility", "Stars")

	for _, r := range repos {
		repoType := "source"
		if r.Mirror {
			repoType = "mirror"
		}

		visibility := successStyle.Render("public")
		if r.Private {
			visibility = warningStyle.Render("private")
		}

		table.AddRow(
			repoStyle.Render(r.FullName),
			dimStyle.Render(repoType),
			visibility,
			fmt.Sprintf("%d", r.Stars),
		)
	}

	cli.Blank()
	cli.Print("  %s\n\n", fmt.Sprintf("%d repositories", len(repos)))
	table.Render()

	return nil
}

// giteaRepo is a simplified repo for display purposes.
type giteaRepo struct {
	Name     string
	FullName string
	Mirror   bool
	Private  bool
	Stars    int
	CloneURL string
}
