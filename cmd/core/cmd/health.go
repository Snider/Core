package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/Snider/Core/pkg/git"
	"github.com/Snider/Core/pkg/repos"
	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
)

var (
	healthLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6b7280")) // gray-500

	healthValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#e2e8f0")) // gray-200

	healthGoodStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22c55e")) // green-500

	healthWarnStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f59e0b")) // amber-500

	healthBadStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ef4444")) // red-500
)

// AddHealthCommand adds the 'health' command to the given parent command.
func AddHealthCommand(parent *clir.Cli) {
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
		totalRepos   = len(statuses)
		dirtyRepos   []string
		aheadRepos   []string
		behindRepos  []string
		errorRepos   []string
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
			fmt.Printf("%s %s\n", healthWarnStyle.Render("Dirty:"), formatRepoList(dirtyRepos))
		}
		if len(aheadRepos) > 0 {
			fmt.Printf("%s %s\n", healthGoodStyle.Render("Ahead:"), formatRepoList(aheadRepos))
		}
		if len(behindRepos) > 0 {
			fmt.Printf("%s %s\n", healthWarnStyle.Render("Behind:"), formatRepoList(behindRepos))
		}
		if len(errorRepos) > 0 {
			fmt.Printf("%s %s\n", healthBadStyle.Render("Errors:"), formatRepoList(errorRepos))
		}
		fmt.Println()
	}

	return nil
}

func printHealthSummary(total int, dirty, ahead, behind, errors []string) {
	// Total repos
	fmt.Print(healthValueStyle.Render(fmt.Sprintf("%d", total)))
	fmt.Print(healthLabelStyle.Render(" repos"))

	// Separator
	fmt.Print(healthLabelStyle.Render(" │ "))

	// Dirty
	if len(dirty) > 0 {
		fmt.Print(healthWarnStyle.Render(fmt.Sprintf("%d", len(dirty))))
		fmt.Print(healthLabelStyle.Render(" dirty"))
	} else {
		fmt.Print(healthGoodStyle.Render("clean"))
	}

	// Separator
	fmt.Print(healthLabelStyle.Render(" │ "))

	// Ahead
	if len(ahead) > 0 {
		fmt.Print(healthValueStyle.Render(fmt.Sprintf("%d", len(ahead))))
		fmt.Print(healthLabelStyle.Render(" to push"))
	} else {
		fmt.Print(healthGoodStyle.Render("synced"))
	}

	// Separator
	fmt.Print(healthLabelStyle.Render(" │ "))

	// Behind
	if len(behind) > 0 {
		fmt.Print(healthWarnStyle.Render(fmt.Sprintf("%d", len(behind))))
		fmt.Print(healthLabelStyle.Render(" to pull"))
	} else {
		fmt.Print(healthGoodStyle.Render("up to date"))
	}

	// Errors (only if any)
	if len(errors) > 0 {
		fmt.Print(healthLabelStyle.Render(" │ "))
		fmt.Print(healthBadStyle.Render(fmt.Sprintf("%d", len(errors))))
		fmt.Print(healthLabelStyle.Render(" errors"))
	}

	fmt.Println()
}

func formatRepoList(repos []string) string {
	if len(repos) <= 5 {
		return joinRepos(repos)
	}
	return joinRepos(repos[:5]) + fmt.Sprintf(" +%d more", len(repos)-5)
}

func joinRepos(repos []string) string {
	result := ""
	for i, r := range repos {
		if i > 0 {
			result += ", "
		}
		result += r
	}
	return result
}
