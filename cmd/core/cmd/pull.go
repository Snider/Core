package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/Snider/Core/pkg/git"
	"github.com/Snider/Core/pkg/repos"
	"github.com/leaanthony/clir"
)

// AddPullCommand adds the 'pull' command to the given parent command.
func AddPullCommand(parent *clir.Cli) {
	var registryPath string
	var all bool

	pullCmd := parent.NewSubCommand("pull", "Pull updates across all repos")
	pullCmd.LongDescription("Pulls updates for all repos.\n" +
		"By default only pulls repos that are behind. Use --all to pull all repos.")

	pullCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)
	pullCmd.BoolFlag("all", "Pull all repos, not just those behind", &all)

	pullCmd.Action(func() error {
		return runPull(registryPath, all)
	})
}

func runPull(registryPath string, all bool) error {
	ctx := context.Background()

	// Find or use provided registry, fall back to directory scan
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
		fmt.Printf("%s %s\n", dimStyle.Render("Registry:"), registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return fmt.Errorf("failed to load registry: %w", err)
			}
			fmt.Printf("%s %s\n", dimStyle.Render("Registry:"), registryPath)
		} else {
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
			fmt.Printf("%s %s\n", dimStyle.Render("Scanning:"), cwd)
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

	// Find repos to pull
	var toPull []git.RepoStatus
	for _, s := range statuses {
		if s.Error != nil {
			continue
		}
		if all || s.HasUnpulled() {
			toPull = append(toPull, s)
		}
	}

	if len(toPull) == 0 {
		fmt.Println("All repos up to date. Nothing to pull.")
		return nil
	}

	// Show what we're pulling
	if all {
		fmt.Printf("\nPulling %d repo(s):\n\n", len(toPull))
	} else {
		fmt.Printf("\n%d repo(s) behind upstream:\n\n", len(toPull))
		for _, s := range toPull {
			fmt.Printf("  %s: %s\n",
				repoNameStyle.Render(s.Name),
				dimStyle.Render(fmt.Sprintf("%d commit(s) behind", s.Behind)),
			)
		}
		fmt.Println()
	}

	// Pull each repo
	var succeeded, failed int
	for _, s := range toPull {
		fmt.Printf("  %s %s... ", dimStyle.Render("Pulling"), s.Name)

		err := gitPull(ctx, s.Path)
		if err != nil {
			fmt.Printf("%s\n", errorStyle.Render("✗ "+err.Error()))
			failed++
		} else {
			fmt.Printf("%s\n", successStyle.Render("✓"))
			succeeded++
		}
	}

	// Summary
	fmt.Println()
	fmt.Printf("%s %d pulled", successStyle.Render("Done:"), succeeded)
	if failed > 0 {
		fmt.Printf(", %s", errorStyle.Render(fmt.Sprintf("%d failed", failed)))
	}
	fmt.Println()

	return nil
}

func gitPull(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "git", "pull", "--ff-only")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}
