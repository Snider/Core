package dev

import (
	"fmt"
	"sort"

	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

// Impact-specific styles (aliases to shared)
var (
	impactDirectStyle   = shared.ErrorStyle
	impactIndirectStyle = shared.StatusWarningStyle
	impactSafeStyle     = shared.StatusSuccessStyle
)

// Impact command flags
var impactRegistryPath string

// addImpactCommand adds the 'impact' command to the given parent command.
func addImpactCommand(parent *cobra.Command) {
	impactCmd := &cobra.Command{
		Use:   "impact <repo-name>",
		Short: i18n.T("cmd.dev.impact.short"),
		Long:  i18n.T("cmd.dev.impact.long"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImpact(impactRegistryPath, args[0])
		},
	}

	impactCmd.Flags().StringVar(&impactRegistryPath, "registry", "", i18n.T("cmd.dev.impact.flag.registry"))

	parent.AddCommand(impactCmd)
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
			return fmt.Errorf(i18n.T("cmd.dev.impact.requires_registry"))
		}
	}

	// Check repo exists
	repo, exists := reg.Get(repoName)
	if !exists {
		return fmt.Errorf(i18n.T("error.repo_not_found", map[string]interface{}{"Name": repoName}))
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
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.impact.analysis_for")), repoNameStyle.Render(repoName))
	if repo.Description != "" {
		fmt.Printf("%s\n", dimStyle.Render(repo.Description))
	}
	fmt.Println()

	if len(allAffected) == 0 {
		fmt.Printf("%s %s\n", impactSafeStyle.Render("v"), i18n.T("cmd.dev.impact.no_dependents", map[string]interface{}{"Name": repoName}))
		return nil
	}

	// Direct dependents
	if len(direct) > 0 {
		fmt.Printf("%s %s\n",
			impactDirectStyle.Render("*"),
			i18n.T("cmd.dev.impact.direct_dependents", map[string]interface{}{"Count": len(direct)}),
		)
		for _, d := range direct {
			r, _ := reg.Get(d)
			desc := ""
			if r != nil && r.Description != "" {
				desc = dimStyle.Render(" - " + shared.Truncate(r.Description, 40))
			}
			fmt.Printf("    %s%s\n", d, desc)
		}
		fmt.Println()
	}

	// Indirect dependents
	if len(indirect) > 0 {
		fmt.Printf("%s %s\n",
			impactIndirectStyle.Render("o"),
			i18n.T("cmd.dev.impact.transitive_dependents", map[string]interface{}{"Count": len(indirect)}),
		)
		for _, d := range indirect {
			r, _ := reg.Get(d)
			desc := ""
			if r != nil && r.Description != "" {
				desc = dimStyle.Render(" - " + shared.Truncate(r.Description, 40))
			}
			fmt.Printf("    %s%s\n", d, desc)
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("%s %s\n",
		dimStyle.Render(i18n.T("cmd.dev.impact.summary")),
		i18n.T("cmd.dev.impact.changes_affect", map[string]interface{}{
			"Repo":     repoNameStyle.Render(repoName),
			"Affected": len(allAffected),
			"Total":    len(reg.Repos) - 1,
		}),
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
