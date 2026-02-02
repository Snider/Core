package dev

import (
	"context"
	"os"
	"path/filepath"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/i18n"
	coreio "github.com/host-uk/core/pkg/io"
)

// Commit command flags
var (
	commitRegistryPath string
	commitAll          bool
)

// AddCommitCommand adds the 'commit' command to the given parent command.
func AddCommitCommand(parent *cli.Command) {
	commitCmd := &cli.Command{
		Use:   "commit",
		Short: i18n.T("cmd.dev.commit.short"),
		Long:  i18n.T("cmd.dev.commit.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runCommit(commitRegistryPath, commitAll)
		},
	}

	commitCmd.Flags().StringVar(&commitRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	commitCmd.Flags().BoolVar(&commitAll, "all", false, i18n.T("cmd.dev.commit.flag.all"))

	parent.AddCommand(commitCmd)
}

func runCommit(registryPath string, all bool) error {
	ctx := context.Background()
	cwd, _ := os.Getwd()

	// Check if current directory is a git repo (single-repo mode)
	if registryPath == "" && isGitRepo(cwd) {
		return runCommitSingleRepo(ctx, cwd, all)
	}

	// Multi-repo mode: find or use provided registry
	reg, regDir, err := loadRegistryWithConfig(registryPath)
	if err != nil {
		return err
	}
	registryPath = regDir // Use resolved registry directory for relative paths

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

	// Find dirty repos
	var dirtyRepos []git.RepoStatus
	for _, s := range statuses {
		if s.Error == nil && s.IsDirty() {
			dirtyRepos = append(dirtyRepos, s)
		}
	}

	if len(dirtyRepos) == 0 {
		cli.Text(i18n.T("cmd.dev.no_changes"))
		return nil
	}

	// Show dirty repos
	cli.Print("\n%s\n\n", i18n.T("cmd.dev.repos_with_changes", map[string]interface{}{"Count": len(dirtyRepos)}))
	for _, s := range dirtyRepos {
		cli.Print("  %s: ", repoNameStyle.Render(s.Name))
		if s.Modified > 0 {
			cli.Print("%s ", dirtyStyle.Render(i18n.T("cmd.dev.modified", map[string]interface{}{"Count": s.Modified})))
		}
		if s.Untracked > 0 {
			cli.Print("%s ", dirtyStyle.Render(i18n.T("cmd.dev.untracked", map[string]interface{}{"Count": s.Untracked})))
		}
		if s.Staged > 0 {
			cli.Print("%s ", aheadStyle.Render(i18n.T("cmd.dev.staged", map[string]interface{}{"Count": s.Staged})))
		}
		cli.Blank()
	}

	// Confirm unless --all
	if !all {
		cli.Blank()
		if !cli.Confirm(i18n.T("cmd.dev.confirm_claude_commit")) {
			cli.Text(i18n.T("cli.aborted"))
			return nil
		}
	}

	cli.Blank()

	// Commit each dirty repo
	var succeeded, failed int
	for _, s := range dirtyRepos {
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.committing")), s.Name)

		if err := claudeCommit(ctx, s.Path, s.Name, registryPath); err != nil {
			cli.Print("  %s %s\n", errorStyle.Render("x"), err)
			failed++
		} else {
			cli.Print("  %s %s\n", successStyle.Render("v"), i18n.T("cmd.dev.committed"))
			succeeded++
		}
		cli.Blank()
	}

	// Summary
	cli.Print("%s", successStyle.Render(i18n.T("cmd.dev.done_succeeded", map[string]interface{}{"Count": succeeded})))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(i18n.T("common.count.failed", map[string]interface{}{"Count": failed})))
	}
	cli.Blank()

	return nil
}

// isGitRepo checks if a directory is a git repository.
func isGitRepo(path string) bool {
	gitDir := path + "/.git"
	_, err := coreio.Local.List(gitDir)
	return err == nil
}

// runCommitSingleRepo handles commit for a single repo (current directory).
func runCommitSingleRepo(ctx context.Context, repoPath string, all bool) error {
	repoName := filepath.Base(repoPath)

	// Get status
	statuses := git.Status(ctx, git.StatusOptions{
		Paths: []string{repoPath},
		Names: map[string]string{repoPath: repoName},
	})

	if len(statuses) == 0 || statuses[0].Error != nil {
		if len(statuses) > 0 && statuses[0].Error != nil {
			return statuses[0].Error
		}
		return cli.Err("failed to get repo status")
	}

	s := statuses[0]
	if !s.IsDirty() {
		cli.Text(i18n.T("cmd.dev.no_changes"))
		return nil
	}

	// Show status
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
	cli.Blank()

	// Confirm unless --all
	if !all {
		cli.Blank()
		if !cli.Confirm(i18n.T("cmd.dev.confirm_claude_commit")) {
			cli.Text(i18n.T("cli.aborted"))
			return nil
		}
	}

	cli.Blank()

	// Commit
	if err := claudeCommit(ctx, repoPath, repoName, ""); err != nil {
		cli.Print("  %s %s\n", errorStyle.Render("x"), err)
		return err
	}
	cli.Print("  %s %s\n", successStyle.Render("v"), i18n.T("cmd.dev.committed"))
	return nil
}
