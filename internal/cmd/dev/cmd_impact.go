package dev

import (
	"errors"
	"sort"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
)

// Impact-specific styles (aliases to shared)
var (
	impactDirectStyle   = cli.ErrorStyle
	impactIndirectStyle = cli.WarningStyle
	impactSafeStyle     = cli.SuccessStyle
)

// Impact command flags
var impactRegistryPath string

// addImpactCommand adds the 'impact' command to the given parent command.
func addImpactCommand(parent *cli.Command) {
	impactCmd := &cli.Command{
		Use:   "impact <repo-name>",
		Short: i18n.T("cmd.dev.impact.short"),
		Long:  i18n.T("cmd.dev.impact.long"),
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			return runImpact(impactRegistryPath, args[0])
		},
	}

	impactCmd.Flags().StringVar(&impactRegistryPath, "registry", "", i18n.T("common.flag.registry"))

	parent.AddCommand(impactCmd)
}

func runImpact(registryPath string, repoName string) error {
	// Find or use provided registry
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return cli.Wrap(err, "failed to load registry")
		}
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return cli.Wrap(err, "failed to load registry")
			}
		} else {
			return errors.New(i18n.T("cmd.dev.impact.requires_registry"))
		}
	}

	// Check repo exists
	repo, exists := reg.Get(repoName)
	if !exists {
		return errors.New(i18n.T("error.repo_not_found", map[string]interface{}{"Name": repoName}))
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
	cli.Blank()
	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.impact.analysis_for")), repoNameStyle.Render(repoName))
	if repo.Description != "" {
		cli.Print("%s\n", dimStyle.Render(repo.Description))
	}
	cli.Blank()

	if len(allAffected) == 0 {
		cli.Print("%s %s\n", impactSafeStyle.Render("v"), i18n.T("cmd.dev.impact.no_dependents", map[string]interface{}{"Name": repoName}))
		return nil
	}

	// Direct dependents
	if len(direct) > 0 {
		cli.Print("%s %s\n",
			impactDirectStyle.Render("*"),
			i18n.T("cmd.dev.impact.direct_dependents", map[string]interface{}{"Count": len(direct)}),
		)
		for _, d := range direct {
			r, _ := reg.Get(d)
			desc := ""
			if r != nil && r.Description != "" {
				desc = dimStyle.Render(" - " + cli.Truncate(r.Description, 40))
			}
			cli.Print("    %s%s\n", d, desc)
		}
		cli.Blank()
	}

	// Indirect dependents
	if len(indirect) > 0 {
		cli.Print("%s %s\n",
			impactIndirectStyle.Render("o"),
			i18n.T("cmd.dev.impact.transitive_dependents", map[string]interface{}{"Count": len(indirect)}),
		)
		for _, d := range indirect {
			r, _ := reg.Get(d)
			desc := ""
			if r != nil && r.Description != "" {
				desc = dimStyle.Render(" - " + cli.Truncate(r.Description, 40))
			}
			cli.Print("    %s%s\n", d, desc)
		}
		cli.Blank()
	}

	// Summary
	cli.Print("%s %s\n",
		dimStyle.Render(i18n.Label("summary")),
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
