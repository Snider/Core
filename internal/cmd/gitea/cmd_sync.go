package gitea

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.gitea.io/sdk/gitea"

	"github.com/host-uk/core/pkg/cli"
	gt "github.com/host-uk/core/pkg/gitea"
)

// Sync command flags.
var (
	syncOrg      string
	syncBasePath string
	syncSetup    bool
)

// addSyncCommand adds the 'sync' subcommand for syncing GitHub repos to Gitea upstream branches.
func addSyncCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "sync <owner/repo> [owner/repo...]",
		Short: "Sync GitHub repos to Gitea upstream branches",
		Long: `Push local GitHub content to Gitea as 'upstream' branches.

Each repo gets:
  - An 'upstream' branch tracking the GitHub default branch
  - A 'main' branch (default) for private tasks, processes, and AI workflows

Use --setup on first run to create the Gitea repos and configure remotes.
Without --setup, updates existing upstream branches from local clones.`,
		Args: cli.MinimumNArgs(0),
		RunE: func(cmd *cli.Command, args []string) error {
			return runSync(args)
		},
	}

	cmd.Flags().StringVar(&syncOrg, "org", "Host-UK", "Gitea organisation")
	cmd.Flags().StringVar(&syncBasePath, "base-path", "~/Code/host-uk", "Base path for local repo clones")
	cmd.Flags().BoolVar(&syncSetup, "setup", false, "Initial setup: create repos, configure remotes, push upstream branches")

	parent.AddCommand(cmd)
}

// repoEntry holds info for a repo to sync.
type repoEntry struct {
	name          string
	localPath     string
	defaultBranch string // the GitHub default branch (main, dev, etc.)
}

func runSync(args []string) error {
	client, err := gt.NewFromConfig("", "")
	if err != nil {
		return err
	}

	// Expand base path
	basePath := syncBasePath
	if strings.HasPrefix(basePath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to resolve home directory: %w", err)
		}
		basePath = filepath.Join(home, basePath[2:])
	}

	// Build repo list: either from args or from the Gitea org
	repos, err := buildRepoList(client, args, basePath)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		cli.Text("No repos to sync.")
		return nil
	}

	giteaURL := client.URL()

	if syncSetup {
		return runSyncSetup(client, repos, giteaURL)
	}

	return runSyncUpdate(repos, giteaURL)
}

func buildRepoList(client *gt.Client, args []string, basePath string) ([]repoEntry, error) {
	var repos []repoEntry

	if len(args) > 0 {
		// Specific repos from args
		for _, arg := range args {
			name := arg
			// Strip owner/ prefix if given
			if parts := strings.SplitN(arg, "/", 2); len(parts) == 2 {
				name = parts[1]
			}
			localPath := filepath.Join(basePath, name)
			branch := detectDefaultBranch(localPath)
			repos = append(repos, repoEntry{
				name:          name,
				localPath:     localPath,
				defaultBranch: branch,
			})
		}
	} else {
		// All repos from the Gitea org
		orgRepos, err := client.ListOrgRepos(syncOrg)
		if err != nil {
			return nil, err
		}
		for _, r := range orgRepos {
			localPath := filepath.Join(basePath, r.Name)
			branch := detectDefaultBranch(localPath)
			repos = append(repos, repoEntry{
				name:          r.Name,
				localPath:     localPath,
				defaultBranch: branch,
			})
		}
	}

	return repos, nil
}

// runSyncSetup handles first-time setup: delete mirrors, create repos, push upstream branches.
func runSyncSetup(client *gt.Client, repos []repoEntry, giteaURL string) error {
	cli.Blank()
	cli.Print("  Setting up %d repos in %s with upstream branches...\n\n", len(repos), syncOrg)

	var succeeded, failed int

	for _, repo := range repos {
		cli.Print("  %s %s\n", dimStyle.Render(">>"), repoStyle.Render(repo.name))

		// Step 1: Delete existing repo (mirror) if it exists
		cli.Print("     Deleting existing mirror... ")
		err := client.DeleteRepo(syncOrg, repo.name)
		if err != nil {
			cli.Print("%s (may not exist)\n", dimStyle.Render("skipped"))
		} else {
			cli.Print("%s\n", successStyle.Render("done"))
		}

		// Step 2: Create empty repo
		cli.Print("     Creating repo... ")
		_, err = client.CreateOrgRepo(syncOrg, gitea.CreateRepoOption{
			Name:          repo.name,
			AutoInit:      false,
			DefaultBranch: "main",
		})
		if err != nil {
			cli.Print("%s\n", errorStyle.Render(err.Error()))
			failed++
			continue
		}
		cli.Print("%s\n", successStyle.Render("done"))

		// Step 3: Add gitea remote to local clone
		cli.Print("     Configuring remote... ")
		remoteURL := fmt.Sprintf("%s/%s/%s.git", giteaURL, syncOrg, repo.name)
		err = configureGiteaRemote(repo.localPath, remoteURL)
		if err != nil {
			cli.Print("%s\n", errorStyle.Render(err.Error()))
			failed++
			continue
		}
		cli.Print("%s\n", successStyle.Render("done"))

		// Step 4: Push default branch as 'upstream' to Gitea
		cli.Print("     Pushing %s -> upstream... ", repo.defaultBranch)
		err = pushUpstream(repo.localPath, repo.defaultBranch)
		if err != nil {
			cli.Print("%s\n", errorStyle.Render(err.Error()))
			failed++
			continue
		}
		cli.Print("%s\n", successStyle.Render("done"))

		// Step 5: Create 'main' branch from 'upstream' on Gitea
		cli.Print("     Creating main branch... ")
		err = createMainFromUpstream(client, syncOrg, repo.name)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "409") {
				cli.Print("%s\n", dimStyle.Render("exists"))
			} else {
				cli.Print("%s\n", errorStyle.Render(err.Error()))
				failed++
				continue
			}
		} else {
			cli.Print("%s\n", successStyle.Render("done"))
		}

		// Step 6: Set default branch to 'main'
		cli.Print("     Setting default branch... ")
		_, _, err = client.API().EditRepo(syncOrg, repo.name, gitea.EditRepoOption{
			DefaultBranch: strPtr("main"),
		})
		if err != nil {
			cli.Print("%s\n", warningStyle.Render(err.Error()))
		} else {
			cli.Print("%s\n", successStyle.Render("main"))
		}

		succeeded++
		cli.Blank()
	}

	cli.Print("  %s", successStyle.Render(fmt.Sprintf("%d repos set up", succeeded)))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(fmt.Sprintf("%d failed", failed)))
	}
	cli.Blank()

	return nil
}

// runSyncUpdate pushes latest from local clones to Gitea upstream branches.
func runSyncUpdate(repos []repoEntry, giteaURL string) error {
	cli.Blank()
	cli.Print("  Syncing %d repos to %s upstream branches...\n\n", len(repos), syncOrg)

	var succeeded, failed int

	for _, repo := range repos {
		cli.Print("  %s -> upstream  ", repoStyle.Render(repo.name))

		// Ensure remote exists
		remoteURL := fmt.Sprintf("%s/%s/%s.git", giteaURL, syncOrg, repo.name)
		_ = configureGiteaRemote(repo.localPath, remoteURL)

		// Fetch latest from GitHub (origin)
		err := gitFetch(repo.localPath, "origin")
		if err != nil {
			cli.Print("%s\n", errorStyle.Render("fetch failed: "+err.Error()))
			failed++
			continue
		}

		// Push to Gitea upstream branch
		err = pushUpstream(repo.localPath, repo.defaultBranch)
		if err != nil {
			cli.Print("%s\n", errorStyle.Render(err.Error()))
			failed++
			continue
		}

		cli.Print("%s\n", successStyle.Render("ok"))
		succeeded++
	}

	cli.Blank()
	cli.Print("  %s", successStyle.Render(fmt.Sprintf("%d synced", succeeded)))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(fmt.Sprintf("%d failed", failed)))
	}
	cli.Blank()

	return nil
}

// detectDefaultBranch returns the default branch for a local git repo.
func detectDefaultBranch(path string) string {
	// Check what origin/HEAD points to
	out, err := exec.Command("git", "-C", path, "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		// refs/remotes/origin/main -> main
		if parts := strings.Split(ref, "/"); len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	// Fallback: check current branch
	out, err = exec.Command("git", "-C", path, "branch", "--show-current").Output()
	if err == nil {
		branch := strings.TrimSpace(string(out))
		if branch != "" {
			return branch
		}
	}

	return "main"
}

// configureGiteaRemote adds or updates the 'gitea' remote on a local repo.
func configureGiteaRemote(localPath, remoteURL string) error {
	// Check if remote exists
	out, err := exec.Command("git", "-C", localPath, "remote", "get-url", "gitea").Output()
	if err == nil {
		// Remote exists — update if URL changed
		existing := strings.TrimSpace(string(out))
		if existing != remoteURL {
			cmd := exec.Command("git", "-C", localPath, "remote", "set-url", "gitea", remoteURL)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to update remote: %w", err)
			}
		}
		return nil
	}

	// Add new remote
	cmd := exec.Command("git", "-C", localPath, "remote", "add", "gitea", remoteURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	return nil
}

// pushUpstream pushes the local default branch to Gitea as 'upstream'.
func pushUpstream(localPath, defaultBranch string) error {
	// Push origin's default branch as 'upstream' to gitea
	refspec := fmt.Sprintf("refs/remotes/origin/%s:refs/heads/upstream", defaultBranch)
	cmd := exec.Command("git", "-C", localPath, "push", "--force", "gitea", refspec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}

	return nil
}

// gitFetch fetches latest from a remote.
func gitFetch(localPath, remote string) error {
	cmd := exec.Command("git", "-C", localPath, "fetch", remote)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// createMainFromUpstream creates a 'main' branch from 'upstream' on Gitea via the API.
func createMainFromUpstream(client *gt.Client, org, repo string) error {
	_, _, err := client.API().CreateBranch(org, repo, gitea.CreateBranchOption{
		BranchName:    "main",
		OldBranchName: "upstream",
	})
	if err != nil {
		return fmt.Errorf("create branch: %w", err)
	}

	return nil
}

func strPtr(s string) *string { return &s }
