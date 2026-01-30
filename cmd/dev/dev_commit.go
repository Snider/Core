package dev

import (
	"context"
	"fmt"
	"os"

	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/git"
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
		Short: "Claude-assisted commits across repos",
		Long: `Uses Claude to create commits for dirty repos.
Shows uncommitted changes and invokes Claude to generate commit messages.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommit(commitRegistryPath, commitAll)
		},
	}

	commitCmd.Flags().StringVar(&commitRegistryPath, "registry", "", "Path to repos.yaml (auto-detected if not specified)")
	commitCmd.Flags().BoolVar(&commitAll, "all", false, "Commit all dirty repos without prompting")

	parent.AddCommand(commitCmd)
}

func runCommit(registryPath string, all bool) error {
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
		fmt.Println("No git repositories found.")
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
		fmt.Println("No uncommitted changes found.")
		return nil
	}

	// Show dirty repos
	fmt.Printf("\n%d repo(s) with uncommitted changes:\n\n", len(dirtyRepos))
	for _, s := range dirtyRepos {
		fmt.Printf("  %s: ", repoNameStyle.Render(s.Name))
		if s.Modified > 0 {
			fmt.Printf("%s ", dirtyStyle.Render(fmt.Sprintf("%d modified", s.Modified)))
		}
		if s.Untracked > 0 {
			fmt.Printf("%s ", dirtyStyle.Render(fmt.Sprintf("%d untracked", s.Untracked)))
		}
		if s.Staged > 0 {
			fmt.Printf("%s ", aheadStyle.Render(fmt.Sprintf("%d staged", s.Staged)))
		}
		fmt.Println()
	}

	// Confirm unless --all
	if !all {
		fmt.Println()
		if !shared.Confirm("Have Claude commit these repos?") {
			fmt.Println("Aborted.")
			return nil
		}
	}

	fmt.Println()

	// Commit each dirty repo
	var succeeded, failed int
	for _, s := range dirtyRepos {
		fmt.Printf("%s %s\n", dimStyle.Render("Committing"), s.Name)

		if err := claudeCommit(ctx, s.Path, s.Name, registryPath); err != nil {
			fmt.Printf("  %s %s\n", errorStyle.Render("x"), err)
			failed++
		} else {
			fmt.Printf("  %s committed\n", successStyle.Render("v"))
			succeeded++
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("%s %d succeeded", successStyle.Render("Done:"), succeeded)
	if failed > 0 {
		fmt.Printf(", %s", errorStyle.Render(fmt.Sprintf("%d failed", failed)))
	}
	fmt.Println()

	return nil
}
