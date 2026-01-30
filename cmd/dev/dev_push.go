package dev

import (
	"context"
	"fmt"
	"os"

	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

// Push command flags
var (
	pushRegistryPath string
	pushForce        bool
)

// addPushCommand adds the 'push' command to the given parent command.
func addPushCommand(parent *cobra.Command) {
	pushCmd := &cobra.Command{
		Use:   "push",
		Short: i18n.T("cmd.dev.push.short"),
		Long:  i18n.T("cmd.dev.push.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(pushRegistryPath, pushForce)
		},
	}

	pushCmd.Flags().StringVar(&pushRegistryPath, "registry", "", i18n.T("cmd.dev.push.flag.registry"))
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, i18n.T("cmd.dev.push.flag.force"))

	parent.AddCommand(pushCmd)
}

func runPush(registryPath string, force bool) error {
	ctx := context.Background()

	// Find or use provided registry, fall back to directory scan
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
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.scanning_label")), cwd)
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

	// Find repos with unpushed commits
	var aheadRepos []git.RepoStatus
	for _, s := range statuses {
		if s.Error == nil && s.HasUnpushed() {
			aheadRepos = append(aheadRepos, s)
		}
	}

	if len(aheadRepos) == 0 {
		fmt.Println(i18n.T("cmd.dev.push.all_up_to_date"))
		return nil
	}

	// Show repos to push
	fmt.Printf("\n%s\n\n", i18n.T("cmd.dev.push.repos_with_unpushed", map[string]interface{}{"Count": len(aheadRepos)}))
	totalCommits := 0
	for _, s := range aheadRepos {
		fmt.Printf("  %s: %s\n",
			repoNameStyle.Render(s.Name),
			aheadStyle.Render(i18n.T("cmd.dev.push.commits_count", map[string]interface{}{"Count": s.Ahead})),
		)
		totalCommits += s.Ahead
	}

	// Confirm unless --force
	if !force {
		fmt.Println()
		if !shared.Confirm(i18n.T("cmd.dev.push.confirm_push", map[string]interface{}{"Commits": totalCommits, "Repos": len(aheadRepos)})) {
			fmt.Println(i18n.T("cli.aborted"))
			return nil
		}
	}

	fmt.Println()

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
			fmt.Printf("  %s %s\n", successStyle.Render("v"), r.Name)
			succeeded++
		} else {
			// Check if this is a non-fast-forward error (diverged branch)
			isNFF := git.IsNonFastForward(r.Error)
			fmt.Fprintf(os.Stderr, "[DEBUG] IsNonFastForward=%v, err=%q\n", isNFF, r.Error)
			if isNFF {
				fmt.Printf("  %s %s: %s\n", warningStyle.Render("!"), r.Name, i18n.T("cmd.dev.push.diverged"))
				divergedRepos = append(divergedRepos, r)
			} else {
				fmt.Printf("  %s %s: %s\n", errorStyle.Render("x"), r.Name, r.Error)
			}
			failed++
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
				succeeded++
				failed--
			}
		}
	}

	// Summary
	fmt.Println()
	fmt.Printf("%s", successStyle.Render(i18n.T("cmd.dev.push.done_pushed", map[string]interface{}{"Count": succeeded})))
	if failed > 0 {
		fmt.Printf(", %s", errorStyle.Render(i18n.T("cmd.dev.count_failed", map[string]interface{}{"Count": failed})))
	}
	fmt.Println()

	return nil
}
