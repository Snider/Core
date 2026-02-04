package gitea

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	gt "github.com/host-uk/core/pkg/gitea"
)

// Mirror command flags.
var (
	mirrorOrg     string
	mirrorGHToken string
)

// addMirrorCommand adds the 'mirror' subcommand for creating GitHub-to-Gitea mirrors.
func addMirrorCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "mirror <github-owner/repo>",
		Short: "Mirror a GitHub repo to Gitea",
		Long: `Create a pull mirror of a GitHub repository on your Gitea instance.

The mirror will be created under the specified Gitea organisation (or your user account).
Gitea will periodically sync changes from GitHub.

For private repos, a GitHub token is needed. By default it uses 'gh auth token'.`,
		Args: cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			owner, repo, err := splitOwnerRepo(args[0])
			if err != nil {
				return err
			}
			return runMirror(owner, repo)
		},
	}

	cmd.Flags().StringVar(&mirrorOrg, "org", "", "Gitea organisation to mirror into (default: your user account)")
	cmd.Flags().StringVar(&mirrorGHToken, "github-token", "", "GitHub token for private repos (default: from gh auth token)")

	parent.AddCommand(cmd)
}

func runMirror(githubOwner, githubRepo string) error {
	client, err := gt.NewFromConfig("", "")
	if err != nil {
		return err
	}

	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", githubOwner, githubRepo)

	// Determine target owner on Gitea
	targetOwner := mirrorOrg
	if targetOwner == "" {
		user, _, err := client.API().GetMyUserInfo()
		if err != nil {
			return cli.WrapVerb(err, "get", "current user")
		}
		targetOwner = user.UserName
	}

	// Resolve GitHub token for source auth
	ghToken := mirrorGHToken
	if ghToken == "" {
		ghToken = resolveGHToken()
	}

	cli.Print("  Mirroring %s/%s -> %s/%s on Gitea...\n", githubOwner, githubRepo, targetOwner, githubRepo)

	repo, err := client.CreateMirror(targetOwner, githubRepo, cloneURL, ghToken)
	if err != nil {
		return err
	}

	cli.Blank()
	cli.Success(fmt.Sprintf("Mirror created: %s", repo.FullName))
	cli.Print("  %s %s\n", dimStyle.Render("URL:"), valueStyle.Render(repo.HTMLURL))
	cli.Print("  %s %s\n", dimStyle.Render("Clone:"), valueStyle.Render(repo.CloneURL))
	cli.Blank()

	return nil
}

// resolveGHToken tries to get a GitHub token from the gh CLI.
func resolveGHToken() string {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
