package dev

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/agentic"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

// Work command flags
var (
	workStatusOnly   bool
	workAutoCommit   bool
	workRegistryPath string
)

// addWorkCommand adds the 'work' command to the given parent command.
func addWorkCommand(parent *cobra.Command) {
	workCmd := &cobra.Command{
		Use:   "work",
		Short: i18n.T("cmd.dev.work.short"),
		Long:  i18n.T("cmd.dev.work.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWork(workRegistryPath, workStatusOnly, workAutoCommit)
		},
	}

	workCmd.Flags().BoolVar(&workStatusOnly, "status", false, i18n.T("cmd.dev.work.flag.status"))
	workCmd.Flags().BoolVar(&workAutoCommit, "commit", false, i18n.T("cmd.dev.work.flag.commit"))
	workCmd.Flags().StringVar(&workRegistryPath, "registry", "", i18n.T("cmd.dev.work.flag.registry"))

	parent.AddCommand(workCmd)
}

func runWork(registryPath string, statusOnly, autoCommit bool) error {
	ctx := context.Background()

	// Find or use provided registry, fall back to directory scan
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
		fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.dev.registry_label")), registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return fmt.Errorf("failed to load registry: %w", err)
			}
			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.dev.registry_label")), registryPath)
		} else {
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.dev.scanning_label")), cwd)
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

	// Sort by repo name for consistent output
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	// Display status table
	printStatusTable(statuses)

	// Collect dirty and ahead repos
	var dirtyRepos []git.RepoStatus
	var aheadRepos []git.RepoStatus

	for _, s := range statuses {
		if s.Error != nil {
			continue
		}
		if s.IsDirty() {
			dirtyRepos = append(dirtyRepos, s)
		}
		if s.HasUnpushed() {
			aheadRepos = append(aheadRepos, s)
		}
	}

	// Auto-commit dirty repos if requested
	if autoCommit && len(dirtyRepos) > 0 {
		fmt.Println()
		fmt.Printf("%s\n", shared.TitleStyle.Render(i18n.T("cmd.dev.commit.committing")))
		fmt.Println()

		for _, s := range dirtyRepos {
			if err := claudeCommit(ctx, s.Path, s.Name, registryPath); err != nil {
				fmt.Printf("  %s %s: %s\n", errorStyle.Render("x"), s.Name, err)
			} else {
				fmt.Printf("  %s %s\n", successStyle.Render("v"), s.Name)
			}
		}

		// Re-check status after commits
		statuses = git.Status(ctx, git.StatusOptions{
			Paths: paths,
			Names: names,
		})

		// Rebuild ahead repos list
		aheadRepos = nil
		for _, s := range statuses {
			if s.Error == nil && s.HasUnpushed() {
				aheadRepos = append(aheadRepos, s)
			}
		}
	}

	// If status only, we're done
	if statusOnly {
		if len(dirtyRepos) > 0 && !autoCommit {
			fmt.Println()
			fmt.Printf("%s\n", dimStyle.Render(i18n.T("cmd.dev.work.use_commit_flag")))
		}
		return nil
	}

	// Push repos with unpushed commits
	if len(aheadRepos) == 0 {
		fmt.Println()
		fmt.Println(i18n.T("cmd.dev.work.all_up_to_date"))
		return nil
	}

	fmt.Println()
	fmt.Printf("%s\n", i18n.T("cmd.dev.work.repos_with_unpushed", map[string]interface{}{"Count": len(aheadRepos)}))
	for _, s := range aheadRepos {
		fmt.Printf("  %s: %s\n", s.Name, i18n.T("cmd.dev.work.commits_count", map[string]interface{}{"Count": s.Ahead}))
	}

	fmt.Println()
	if !shared.Confirm(i18n.T("cmd.dev.push.confirm")) {
		fmt.Println(i18n.T("cli.aborted"))
		return nil
	}

	fmt.Println()

	// Push sequentially (SSH passphrase needs interaction)
	var pushPaths []string
	for _, s := range aheadRepos {
		pushPaths = append(pushPaths, s.Path)
	}

	results := git.PushMultiple(ctx, pushPaths, names)

	var divergedRepos []git.PushResult

	for _, r := range results {
		if r.Success {
			fmt.Printf("  %s %s\n", successStyle.Render("v"), r.Name)
		} else {
			// Check if this is a non-fast-forward error (diverged branch)
			if git.IsNonFastForward(r.Error) {
				fmt.Printf("  %s %s: %s\n", warningStyle.Render("!"), r.Name, i18n.T("cmd.dev.push.diverged"))
				divergedRepos = append(divergedRepos, r)
			} else {
				fmt.Printf("  %s %s: %s\n", errorStyle.Render("x"), r.Name, r.Error)
			}
		}
	}

	// Handle diverged repos - offer to pull and retry
	if len(divergedRepos) > 0 {
		fmt.Println()
		fmt.Printf("%s\n", i18n.T("cmd.dev.push.diverged_help"))
		if shared.Confirm(i18n.T("cmd.dev.push.pull_and_retry")) {
			fmt.Println()
			for _, r := range divergedRepos {
				fmt.Printf("  %s %s...\n", dimStyle.Render("↓"), r.Name)
				if err := git.Pull(ctx, r.Path); err != nil {
					fmt.Printf("  %s %s: %s\n", errorStyle.Render("x"), r.Name, err)
					continue
				}
				fmt.Printf("  %s %s...\n", dimStyle.Render("↑"), r.Name)
				if err := git.Push(ctx, r.Path); err != nil {
					fmt.Printf("  %s %s: %s\n", errorStyle.Render("x"), r.Name, err)
					continue
				}
				fmt.Printf("  %s %s\n", successStyle.Render("v"), r.Name)
			}
		}
	}

	return nil
}

func printStatusTable(statuses []git.RepoStatus) {
	// Calculate column widths
	nameWidth := 4 // "Repo"
	for _, s := range statuses {
		if len(s.Name) > nameWidth {
			nameWidth = len(s.Name)
		}
	}

	// Print header with fixed-width formatting
	fmt.Printf("%-*s  %8s  %9s  %6s  %5s\n",
		nameWidth,
		shared.TitleStyle.Render(i18n.T("cmd.dev.work.table_repo")),
		shared.TitleStyle.Render(i18n.T("cmd.dev.work.table_modified")),
		shared.TitleStyle.Render(i18n.T("cmd.dev.work.table_untracked")),
		shared.TitleStyle.Render(i18n.T("cmd.dev.work.table_staged")),
		shared.TitleStyle.Render(i18n.T("cmd.dev.work.table_ahead")),
	)

	// Print separator
	fmt.Println(strings.Repeat("-", nameWidth+2+10+11+8+7))

	// Print rows
	for _, s := range statuses {
		if s.Error != nil {
			paddedName := fmt.Sprintf("%-*s", nameWidth, s.Name)
			fmt.Printf("%s  %s\n",
				repoNameStyle.Render(paddedName),
				errorStyle.Render(i18n.T("cmd.dev.work.error_prefix")+" "+s.Error.Error()),
			)
			continue
		}

		// Style numbers based on values
		modStr := fmt.Sprintf("%d", s.Modified)
		if s.Modified > 0 {
			modStr = dirtyStyle.Render(modStr)
		} else {
			modStr = cleanStyle.Render(modStr)
		}

		untrackedStr := fmt.Sprintf("%d", s.Untracked)
		if s.Untracked > 0 {
			untrackedStr = dirtyStyle.Render(untrackedStr)
		} else {
			untrackedStr = cleanStyle.Render(untrackedStr)
		}

		stagedStr := fmt.Sprintf("%d", s.Staged)
		if s.Staged > 0 {
			stagedStr = aheadStyle.Render(stagedStr)
		} else {
			stagedStr = cleanStyle.Render(stagedStr)
		}

		aheadStr := fmt.Sprintf("%d", s.Ahead)
		if s.Ahead > 0 {
			aheadStr = aheadStyle.Render(aheadStr)
		} else {
			aheadStr = cleanStyle.Render(aheadStr)
		}

		// Pad name before styling to avoid ANSI code length issues
		paddedName := fmt.Sprintf("%-*s", nameWidth, s.Name)
		fmt.Printf("%s  %8s  %9s  %6s  %5s\n",
			repoNameStyle.Render(paddedName),
			modStr,
			untrackedStr,
			stagedStr,
			aheadStr,
		)
	}
}

func claudeCommit(ctx context.Context, repoPath, repoName, registryPath string) error {
	prompt := agentic.Prompt("commit")

	cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--allowedTools", "Bash,Read,Glob,Grep")
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func claudeEditCommit(ctx context.Context, repoPath, repoName, registryPath string) error {
	prompt := agentic.Prompt("commit")

	cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--allowedTools", "Bash,Read,Write,Edit,Glob,Grep")
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
