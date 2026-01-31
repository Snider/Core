package dev

import (
	"context"
	"os"
	"path/filepath"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
)

// Push command flags
var (
	pushRegistryPath string
	pushForce        bool
)

// addPushCommand adds the 'push' command to the given parent command.
func addPushCommand(parent *cli.Command) {
	pushCmd := &cli.Command{
		Use:   "push",
		Short: i18n.T("cmd.dev.push.short"),
		Long:  i18n.T("cmd.dev.push.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runPush(pushRegistryPath, pushForce)
		},
	}

	pushCmd.Flags().StringVar(&pushRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, i18n.T("cmd.dev.push.flag.force"))

	parent.AddCommand(pushCmd)
}

func runPush(registryPath string, force bool) error {
	ctx := context.Background()
	cwd, _ := os.Getwd()

	// Check if current directory is a git repo (single-repo mode)
	if registryPath == "" && isGitRepo(cwd) {
		return runPushSingleRepo(ctx, cwd, force)
	}

	// Multi-repo mode: find or use provided registry
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return cli.Wrap(err, "failed to load registry")
		}
		cli.Print("%s %s\n", dimStyle.Render(i18n.Label("registry")), registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return cli.Wrap(err, "failed to load registry")
			}
			cli.Print("%s %s\n", dimStyle.Render(i18n.Label("registry")), registryPath)
		} else {
			// Fallback: scan current directory for repos
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return cli.Wrap(err, "failed to scan directory")
			}
			cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.scanning_label")), cwd)
		}
	}

	// Build paths and names for git operations
	var paths []string
	names := make(map[string]string)

	for _, repo := range reg.List() {
		if repo.IsGitRepo() {
			paths = append(paths, repo.Path)
			names[repo.Path] = repo.Name
		}
	}

	if len(paths) == 0 {
		cli.Text(i18n.T("cmd.dev.no_git_repos"))
		return nil
	}

	// Get status for all repos
	statuses := git.Status(ctx, git.StatusOptions{
		Paths: paths,
		Names: names,
	})

	// Find repos with unpushed commits
	var aheadRepos []git.RepoStatus
	for _, s := range statuses {
		if s.Error == nil && s.HasUnpushed() {
			aheadRepos = append(aheadRepos, s)
		}
	}

	if len(aheadRepos) == 0 {
		cli.Text(i18n.T("cmd.dev.push.all_up_to_date"))
		return nil
	}

	// Show repos to push
	cli.Print("\n%s\n\n", i18n.T("common.count.repos_unpushed", map[string]interface{}{"Count": len(aheadRepos)}))
	totalCommits := 0
	for _, s := range aheadRepos {
		cli.Print("  %s: %s\n",
			repoNameStyle.Render(s.Name),
			aheadStyle.Render(i18n.T("common.count.commits", map[string]interface{}{"Count": s.Ahead})),
		)
		totalCommits += s.Ahead
	}

	// Confirm unless --force
	if !force {
		cli.Line("")
		if !cli.Confirm(i18n.T("cmd.dev.push.confirm_push", map[string]interface{}{"Commits": totalCommits, "Repos": len(aheadRepos)})) {
			cli.Text(i18n.T("cli.aborted"))
			return nil
		}
	}

	cli.Line("")

	// Push sequentially (SSH passphrase needs interaction)
	var pushPaths []string
	for _, s := range aheadRepos {
		pushPaths = append(pushPaths, s.Path)
	}

	results := git.PushMultiple(ctx, pushPaths, names)

	var succeeded, failed int
	var divergedRepos []git.PushResult

	for _, r := range results {
		if r.Success {
			cli.Print("  %s %s\n", successStyle.Render("v"), r.Name)
			succeeded++
		} else {
			// Check if this is a non-fast-forward error (diverged branch)
			if git.IsNonFastForward(r.Error) {
				cli.Print("  %s %s: %s\n", warningStyle.Render("!"), r.Name, i18n.T("cmd.dev.push.diverged"))
				divergedRepos = append(divergedRepos, r)
			} else {
				cli.Print("  %s %s: %s\n", errorStyle.Render("x"), r.Name, r.Error)
			}
			failed++
		}
	}

	// Handle diverged repos - offer to pull and retry
	if len(divergedRepos) > 0 {
		cli.Line("")
		cli.Print("%s\n", i18n.T("cmd.dev.push.diverged_help"))
		if cli.Confirm(i18n.T("cmd.dev.push.pull_and_retry")) {
			cli.Line("")
			for _, r := range divergedRepos {
				cli.Print("  %s %s...\n", dimStyle.Render("↓"), r.Name)
				if err := git.Pull(ctx, r.Path); err != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), r.Name, err)
					continue
				}
				cli.Print("  %s %s...\n", dimStyle.Render("↑"), r.Name)
				if err := git.Push(ctx, r.Path); err != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), r.Name, err)
					continue
				}
				cli.Print("  %s %s\n", successStyle.Render("v"), r.Name)
				succeeded++
				failed--
			}
		}
	}

	// Summary
	cli.Line("")
	cli.Print("%s", successStyle.Render(i18n.T("cmd.dev.push.done_pushed", map[string]interface{}{"Count": succeeded})))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(i18n.T("common.count.failed", map[string]interface{}{"Count": failed})))
	}
	cli.Line("")

	return nil
}

// runPushSingleRepo handles push for a single repo (current directory).
func runPushSingleRepo(ctx context.Context, repoPath string, force bool) error {
	repoName := filepath.Base(repoPath)

	// Get status
	statuses := git.Status(ctx, git.StatusOptions{
		Paths: []string{repoPath},
		Names: map[string]string{repoPath: repoName},
	})

	if len(statuses) == 0 {
		return cli.Err("failed to get repo status")
	}

	s := statuses[0]
	if s.Error != nil {
		return s.Error
	}

	if !s.HasUnpushed() {
		// Check if there are uncommitted changes
		if s.IsDirty() {
			cli.Print("%s: ", repoNameStyle.Render(s.Name))
			if s.Modified > 0 {
				cli.Print("%s ", dirtyStyle.Render(i18n.T("cmd.dev.modified", map[string]interface{}{"Count": s.Modified})))
			}
			if s.Untracked > 0 {
				cli.Print("%s ", dirtyStyle.Render(i18n.T("cmd.dev.untracked", map[string]interface{}{"Count": s.Untracked})))
			}
			if s.Staged > 0 {
				cli.Print("%s ", aheadStyle.Render(i18n.T("cmd.dev.staged", map[string]interface{}{"Count": s.Staged})))
			}
			cli.Line("")
			cli.Line("")
			if cli.Confirm(i18n.T("cmd.dev.push.uncommitted_changes_commit")) {
				cli.Line("")
				// Use edit-enabled commit if only untracked files (may need .gitignore fix)
				var err error
				if s.Modified == 0 && s.Staged == 0 && s.Untracked > 0 {
					err = claudeEditCommit(ctx, repoPath, repoName, "")
				} else {
					err = runCommitSingleRepo(ctx, repoPath, false)
				}
				if err != nil {
					return err
				}
				// Re-check - only push if Claude created commits
				newStatuses := git.Status(ctx, git.StatusOptions{
					Paths: []string{repoPath},
					Names: map[string]string{repoPath: repoName},
				})
				if len(newStatuses) > 0 && newStatuses[0].HasUnpushed() {
					return runPushSingleRepo(ctx, repoPath, force)
				}
			}
			return nil
		}
		cli.Text(i18n.T("cmd.dev.push.all_up_to_date"))
		return nil
	}

	// Show commits to push
	cli.Print("%s: %s\n", repoNameStyle.Render(s.Name),
		aheadStyle.Render(i18n.T("common.count.commits", map[string]interface{}{"Count": s.Ahead})))

	// Confirm unless --force
	if !force {
		cli.Line("")
		if !cli.Confirm(i18n.T("cmd.dev.push.confirm_push", map[string]interface{}{"Commits": s.Ahead, "Repos": 1})) {
			cli.Text(i18n.T("cli.aborted"))
			return nil
		}
	}

	cli.Line("")

	// Push
	err := git.Push(ctx, repoPath)
	if err != nil {
		if git.IsNonFastForward(err) {
			cli.Print("  %s %s: %s\n", warningStyle.Render("!"), repoName, i18n.T("cmd.dev.push.diverged"))
			cli.Line("")
			cli.Print("%s\n", i18n.T("cmd.dev.push.diverged_help"))
			if cli.Confirm(i18n.T("cmd.dev.push.pull_and_retry")) {
				cli.Line("")
				cli.Print("  %s %s...\n", dimStyle.Render("↓"), repoName)
				if pullErr := git.Pull(ctx, repoPath); pullErr != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), repoName, pullErr)
					return pullErr
				}
				cli.Print("  %s %s...\n", dimStyle.Render("↑"), repoName)
				if pushErr := git.Push(ctx, repoPath); pushErr != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), repoName, pushErr)
					return pushErr
				}
				cli.Print("  %s %s\n", successStyle.Render("v"), repoName)
				return nil
			}
		}
		cli.Print("  %s %s: %s\n", errorStyle.Render("x"), repoName, err)
		return err
	}

	cli.Print("  %s %s\n", successStyle.Render("v"), repoName)
	return nil
}
