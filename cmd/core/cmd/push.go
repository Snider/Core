package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Snider/Core/pkg/git"
	"github.com/Snider/Core/pkg/repos"
	"github.com/leaanthony/clir"
)

// AddPushCommand adds the 'push' command to the given parent command.
func AddPushCommand(parent *clir.Cli) {
	var registryPath string
	var force bool

	pushCmd := parent.NewSubCommand("push", "Push commits across all repos")
	pushCmd.LongDescription("Pushes unpushed commits for all repos.\n" +
		"Shows repos with commits to push and confirms before pushing.")

	pushCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)
	pushCmd.BoolFlag("force", "Skip confirmation prompt", &force)

	pushCmd.Action(func() error {
		return runPush(registryPath, force)
	})
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
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return fmt.Errorf("failed to load registry: %w", err)
			}
		} else {
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
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
		fmt.Println("No git repositories found.")
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
		fmt.Println("All repos up to date. Nothing to push.")
		return nil
	}

	// Show repos to push
	fmt.Printf("\n%d repo(s) with unpushed commits:\n\n", len(aheadRepos))
	totalCommits := 0
	for _, s := range aheadRepos {
		fmt.Printf("  %s: %s\n",
			repoNameStyle.Render(s.Name),
			aheadStyle.Render(fmt.Sprintf("%d commit(s)", s.Ahead)),
		)
		totalCommits += s.Ahead
	}

	// Confirm unless --force
	if !force {
		fmt.Println()
		if !confirm(fmt.Sprintf("Push %d commit(s) to %d repo(s)?", totalCommits, len(aheadRepos))) {
			fmt.Println("Aborted.")
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
	for _, r := range results {
		if r.Success {
			fmt.Printf("  %s %s\n", successStyle.Render("✓"), r.Name)
			succeeded++
		} else {
			fmt.Printf("  %s %s: %s\n", errorStyle.Render("✗"), r.Name, r.Error)
			failed++
		}
	}

	// Summary
	fmt.Println()
	fmt.Printf("%s %d pushed", successStyle.Render("Done:"), succeeded)
	if failed > 0 {
		fmt.Printf(", %s", errorStyle.Render(fmt.Sprintf("%d failed", failed)))
	}
	fmt.Println()

	return nil
}
