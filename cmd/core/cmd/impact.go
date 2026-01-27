package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/Snider/Core/pkg/repos"
	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
)

var (
	impactDirectStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ef4444")) // red-500

	impactIndirectStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f59e0b")) // amber-500

	impactSafeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22c55e")) // green-500
)

// AddImpactCommand adds the 'impact' command to the given parent command.
func AddImpactCommand(parent *clir.Cli) {
	var registryPath string

	impactCmd := parent.NewSubCommand("impact", "Show impact of changing a repo")
	impactCmd.LongDescription("Analyzes the dependency graph to show which repos\n" +
		"would be affected by changes to the specified repo.")

	impactCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)

	impactCmd.Action(func() error {
		args := os.Args[2:] // Skip "core" and "impact"
		// Filter out flags
		var repoName string
		for _, arg := range args {
			if arg[0] != '-' {
				repoName = arg
				break
			}
		}
		if repoName == "" {
			return fmt.Errorf("usage: core impact <repo-name>")
		}
		return runImpact(registryPath, repoName)
	})
}

func runImpact(registryPath string, repoName string) error {
	// Find or use provided registry
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
			return fmt.Errorf("impact analysis requires repos.yaml with dependency information")
		}
	}

	// Check repo exists
	repo, exists := reg.Get(repoName)
	if !exists {
		return fmt.Errorf("repo '%s' not found in registry", repoName)
	}

	// Build reverse dependency graph
	dependents := buildDependentsGraph(reg)

	// Find all affected repos (direct and transitive)
	direct := dependents[repoName]
	allAffected := findAllDependents(repoName, dependents)

	// Separate direct vs indirect
	directSet := make(map[string]bool)
	for _, d := range direct {
		directSet[d] = true
	}

	var indirect []string
	for _, a := range allAffected {
		if !directSet[a] {
			indirect = append(indirect, a)
		}
	}

	// Sort for consistent output
	sort.Strings(direct)
	sort.Strings(indirect)

	// Print results
	fmt.Println()
	fmt.Printf("%s %s\n", dimStyle.Render("Impact analysis for"), repoNameStyle.Render(repoName))
	if repo.Description != "" {
		fmt.Printf("%s\n", dimStyle.Render(repo.Description))
	}
	fmt.Println()

	if len(allAffected) == 0 {
		fmt.Printf("%s No repos depend on %s\n", impactSafeStyle.Render("✓"), repoName)
		return nil
	}

	// Direct dependents
	if len(direct) > 0 {
		fmt.Printf("%s %d direct dependent(s):\n",
			impactDirectStyle.Render("●"),
			len(direct),
		)
		for _, d := range direct {
			r, _ := reg.Get(d)
			desc := ""
			if r != nil && r.Description != "" {
				desc = dimStyle.Render(" - " + truncate(r.Description, 40))
			}
			fmt.Printf("    %s%s\n", d, desc)
		}
		fmt.Println()
	}

	// Indirect dependents
	if len(indirect) > 0 {
		fmt.Printf("%s %d transitive dependent(s):\n",
			impactIndirectStyle.Render("○"),
			len(indirect),
		)
		for _, d := range indirect {
			r, _ := reg.Get(d)
			desc := ""
			if r != nil && r.Description != "" {
				desc = dimStyle.Render(" - " + truncate(r.Description, 40))
			}
			fmt.Printf("    %s%s\n", d, desc)
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("%s Changes to %s affect %s\n",
		dimStyle.Render("Summary:"),
		repoNameStyle.Render(repoName),
		impactDirectStyle.Render(fmt.Sprintf("%d/%d repos", len(allAffected), len(reg.Repos)-1)),
	)

	return nil
}

// buildDependentsGraph creates a reverse dependency map
// key = repo, value = repos that depend on it
func buildDependentsGraph(reg *repos.Registry) map[string][]string {
	dependents := make(map[string][]string)

	for name, repo := range reg.Repos {
		for _, dep := range repo.DependsOn {
			dependents[dep] = append(dependents[dep], name)
		}
	}

	return dependents
}

// findAllDependents recursively finds all repos that depend on the given repo
func findAllDependents(repoName string, dependents map[string][]string) []string {
	visited := make(map[string]bool)
	var result []string

	var visit func(name string)
	visit = func(name string) {
		for _, dep := range dependents[name] {
			if !visited[dep] {
				visited[dep] = true
				result = append(result, dep)
				visit(dep) // Recurse for transitive deps
			}
		}
	}

	visit(repoName)
	return result
}
