package dev

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/repos"
	"github.com/leaanthony/clir"
)

// addHealthCommand adds the 'health' command to the given parent command.
func addHealthCommand(parent *clir.Command) {
	var registryPath string
	var verbose bool

	healthCmd := parent.NewSubCommand("health", "Quick health check across all repos")
	healthCmd.LongDescription("Shows a summary of repository health:\n" +
		"total repos, dirty repos, unpushed commits, etc.")

	healthCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)
	healthCmd.BoolFlag("verbose", "Show detailed breakdown", &verbose)

	healthCmd.Action(func() error {
		return runHealth(registryPath, verbose)
	})
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
	// Total repos
	fmt.Print(valueStyle.Render(fmt.Sprintf("%d", total)))
	fmt.Print(dimStyle.Render(" repos"))

	// Separator
	fmt.Print(dimStyle.Render(" | "))

	// Dirty
	if len(dirty) > 0 {
		fmt.Print(warningStyle.Render(fmt.Sprintf("%d", len(dirty))))
		fmt.Print(dimStyle.Render(" dirty"))
	} else {
		fmt.Print(successStyle.Render("clean"))
	}

	// Separator
	fmt.Print(dimStyle.Render(" | "))

	// Ahead
	if len(ahead) > 0 {
		fmt.Print(valueStyle.Render(fmt.Sprintf("%d", len(ahead))))
		fmt.Print(dimStyle.Render(" to push"))
	} else {
		fmt.Print(successStyle.Render("synced"))
	}

	// Separator
	fmt.Print(dimStyle.Render(" | "))

	// Behind
	if len(behind) > 0 {
		fmt.Print(warningStyle.Render(fmt.Sprintf("%d", len(behind))))
		fmt.Print(dimStyle.Render(" to pull"))
	} else {
		fmt.Print(successStyle.Render("up to date"))
	}

	// Errors (only if any)
	if len(errors) > 0 {
		fmt.Print(dimStyle.Render(" | "))
		fmt.Print(errorStyle.Render(fmt.Sprintf("%d", len(errors))))
		fmt.Print(dimStyle.Render(" errors"))
	}

	fmt.Println()
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
