package dev

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

// Commit command flags
var (
	commitRegistryPath string
	commitAll          bool
)

// addCommitCommand adds the 'commit' command to the given parent command.
func addCommitCommand(parent *cobra.Command) {
	commitCmd := &cobra.Command{
		Use:   "commit",
		Short: i18n.T("cmd.dev.commit.short"),
		Long:  i18n.T("cmd.dev.commit.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommit(commitRegistryPath, commitAll)
		},
	}

	commitCmd.Flags().StringVar(&commitRegistryPath, "registry", "", i18n.T("cmd.dev.commit.flag.registry"))
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
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.registry_label")), registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return fmt.Errorf("failed to load registry: %w", err)
			}
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.registry_label")), registryPath)
		} else {
			// Fallback: scan current directory for repos
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.scanning_label")), cwd)
			registryPath = cwd
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
		fmt.Println(i18n.T("cmd.dev.no_git_repos"))
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
		fmt.Println(i18n.T("cmd.dev.no_changes"))
		return nil
	}

	// Show dirty repos
	fmt.Printf("\n%s\n\n", i18n.T("cmd.dev.repos_with_changes", map[string]interface{}{"Count": len(dirtyRepos)}))
	for _, s := range dirtyRepos {
		fmt.Printf("  %s: ", repoNameStyle.Render(s.Name))
		if s.Modified > 0 {
			fmt.Printf("%s ", dirtyStyle.Render(i18n.T("cmd.dev.modified", map[string]interface{}{"Count": s.Modified})))
		}
		if s.Untracked > 0 {
			fmt.Printf("%s ", dirtyStyle.Render(i18n.T("cmd.dev.untracked", map[string]interface{}{"Count": s.Untracked})))
		}
		if s.Staged > 0 {
			fmt.Printf("%s ", aheadStyle.Render(i18n.T("cmd.dev.staged", map[string]interface{}{"Count": s.Staged})))
		}
		fmt.Println()
	}

	// Confirm unless --all
	if !all {
		fmt.Println()
		if !shared.Confirm(i18n.T("cmd.dev.confirm_claude_commit")) {
			fmt.Println(i18n.T("cli.aborted"))
			return nil
		}
	}

	fmt.Println()

	// Commit each dirty repo
	var succeeded, failed int
	for _, s := range dirtyRepos {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.committing")), s.Name)

		if err := claudeCommit(ctx, s.Path, s.Name, registryPath); err != nil {
			fmt.Printf("  %s %s\n", errorStyle.Render("x"), err)
			failed++
		} else {
			fmt.Printf("  %s %s\n", successStyle.Render("v"), i18n.T("cmd.dev.committed"))
			succeeded++
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("%s", successStyle.Render(i18n.T("cmd.dev.done_succeeded", map[string]interface{}{"Count": succeeded})))
	if failed > 0 {
		fmt.Printf(", %s", errorStyle.Render(i18n.T("cmd.dev.count_failed", map[string]interface{}{"Count": failed})))
	}
	fmt.Println()

	return nil
}

// isGitRepo checks if a directory is a git repository.
func isGitRepo(path string) bool {
	gitDir := path + "/.git"
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
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
		return fmt.Errorf("failed to get repo status")
	}

	s := statuses[0]
	if !s.IsDirty() {
		fmt.Println(i18n.T("cmd.dev.no_changes"))
		return nil
	}

	// Show status
	fmt.Printf("%s: ", repoNameStyle.Render(s.Name))
	if s.Modified > 0 {
		fmt.Printf("%s ", dirtyStyle.Render(i18n.T("cmd.dev.modified", map[string]interface{}{"Count": s.Modified})))
	}
	if s.Untracked > 0 {
		fmt.Printf("%s ", dirtyStyle.Render(i18n.T("cmd.dev.untracked", map[string]interface{}{"Count": s.Untracked})))
	}
	if s.Staged > 0 {
		fmt.Printf("%s ", aheadStyle.Render(i18n.T("cmd.dev.staged", map[string]interface{}{"Count": s.Staged})))
	}
	fmt.Println()

	// Confirm unless --all
	if !all {
		fmt.Println()
		if !shared.Confirm(i18n.T("cmd.dev.confirm_claude_commit")) {
			fmt.Println(i18n.T("cli.aborted"))
			return nil
		}
	}

	fmt.Println()

	// Commit
	if err := claudeCommit(ctx, repoPath, repoName, ""); err != nil {
		fmt.Printf("  %s %s\n", errorStyle.Render("x"), err)
		return err
	}
	fmt.Printf("  %s %s\n", successStyle.Render("v"), i18n.T("cmd.dev.committed"))
	return nil
}
