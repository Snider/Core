package dev

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

// Health command flags
var (
	healthRegistryPath string
	healthVerbose      bool
)

// addHealthCommand adds the 'health' command to the given parent command.
func addHealthCommand(parent *cobra.Command) {
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Quick health check across all repos",
		Long: `Shows a summary of repository health:
total repos, dirty repos, unpushed commits, etc.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHealth(healthRegistryPath, healthVerbose)
		},
	}

	healthCmd.Flags().StringVar(&healthRegistryPath, "registry", "", "Path to repos.yaml (auto-detected if not specified)")
	healthCmd.Flags().BoolVarP(&healthVerbose, "verbose", "v", false, "Show detailed breakdown")

	parent.AddCommand(healthCmd)
}

func runHealth(registryPath string, verbose bool) error {
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

	// Sort for consistent verbose output
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	// Aggregate stats
	var (
		totalRepos  = len(statuses)
		dirtyRepos  []string
		aheadRepos  []string
		behindRepos []string
		errorRepos  []string
	)

	for _, s := range statuses {
		if s.Error != nil {
			errorRepos = append(errorRepos, s.Name)
			continue
		}
		if s.IsDirty() {
			dirtyRepos = append(dirtyRepos, s.Name)
		}
		if s.HasUnpushed() {
			aheadRepos = append(aheadRepos, s.Name)
		}
		if s.HasUnpulled() {
			behindRepos = append(behindRepos, s.Name)
		}
	}

	// Print summary line
	fmt.Println()
	printHealthSummary(totalRepos, dirtyRepos, aheadRepos, behindRepos, errorRepos)
	fmt.Println()

	// Verbose output
	if verbose {
		if len(dirtyRepos) > 0 {
			fmt.Printf("%s %s\n", warningStyle.Render("Dirty:"), formatRepoList(dirtyRepos))
		}
		if len(aheadRepos) > 0 {
			fmt.Printf("%s %s\n", successStyle.Render("Ahead:"), formatRepoList(aheadRepos))
		}
		if len(behindRepos) > 0 {
			fmt.Printf("%s %s\n", warningStyle.Render("Behind:"), formatRepoList(behindRepos))
		}
		if len(errorRepos) > 0 {
			fmt.Printf("%s %s\n", errorStyle.Render("Errors:"), formatRepoList(errorRepos))
		}
		fmt.Println()
	}

	return nil
}

func printHealthSummary(total int, dirty, ahead, behind, errors []string) {
	parts := []string{
		shared.StatusPart(total, "repos", shared.ValueStyle),
	}

	// Dirty status
	if len(dirty) > 0 {
		parts = append(parts, shared.StatusPart(len(dirty), "dirty", shared.WarningStyle))
	} else {
		parts = append(parts, shared.StatusText("clean", shared.SuccessStyle))
	}

	// Push status
	if len(ahead) > 0 {
		parts = append(parts, shared.StatusPart(len(ahead), "to push", shared.ValueStyle))
	} else {
		parts = append(parts, shared.StatusText("synced", shared.SuccessStyle))
	}

	// Pull status
	if len(behind) > 0 {
		parts = append(parts, shared.StatusPart(len(behind), "to pull", shared.WarningStyle))
	} else {
		parts = append(parts, shared.StatusText("up to date", shared.SuccessStyle))
	}

	// Errors (only if any)
	if len(errors) > 0 {
		parts = append(parts, shared.StatusPart(len(errors), "errors", shared.ErrorStyle))
	}

	fmt.Println(shared.StatusLine(parts...))
}

func formatRepoList(reposList []string) string {
	if len(reposList) <= 5 {
		return joinRepos(reposList)
	}
	return joinRepos(reposList[:5]) + fmt.Sprintf(" +%d more", len(reposList)-5)
}

func joinRepos(reposList []string) string {
	result := ""
	for i, r := range reposList {
		if i > 0 {
			result += ", "
		}
		result += r
	}
	return result
}
