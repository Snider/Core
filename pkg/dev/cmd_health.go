package dev

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
)

// Health command flags
var (
	healthRegistryPath string
	healthVerbose      bool
)

// addHealthCommand adds the 'health' command to the given parent command.
func addHealthCommand(parent *cli.Command) {
	healthCmd := &cli.Command{
		Use:   "health",
		Short: i18n.T("cmd.dev.health.short"),
		Long:  i18n.T("cmd.dev.health.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runHealth(healthRegistryPath, healthVerbose)
		},
	}

	healthCmd.Flags().StringVar(&healthRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	healthCmd.Flags().BoolVarP(&healthVerbose, "verbose", "v", false, i18n.T("cmd.dev.health.flag.verbose"))

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
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return cli.Wrap(err, "failed to scan directory")
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
		cli.Text(i18n.T("cmd.dev.no_git_repos"))
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
	cli.Blank()
	printHealthSummary(totalRepos, dirtyRepos, aheadRepos, behindRepos, errorRepos)
	cli.Blank()

	// Verbose output
	if verbose {
		if len(dirtyRepos) > 0 {
			cli.Print("%s %s\n", warningStyle.Render(i18n.T("cmd.dev.health.dirty_label")), formatRepoList(dirtyRepos))
		}
		if len(aheadRepos) > 0 {
			cli.Print("%s %s\n", successStyle.Render(i18n.T("cmd.dev.health.ahead_label")), formatRepoList(aheadRepos))
		}
		if len(behindRepos) > 0 {
			cli.Print("%s %s\n", warningStyle.Render(i18n.T("cmd.dev.health.behind_label")), formatRepoList(behindRepos))
		}
		if len(errorRepos) > 0 {
			cli.Print("%s %s\n", errorStyle.Render(i18n.T("cmd.dev.health.errors_label")), formatRepoList(errorRepos))
		}
		cli.Blank()
	}

	return nil
}

func printHealthSummary(total int, dirty, ahead, behind, errors []string) {
	parts := []string{
		statusPart(total, i18n.T("cmd.dev.health.repos"), cli.ValueStyle),
	}

	// Dirty status
	if len(dirty) > 0 {
		parts = append(parts, statusPart(len(dirty), i18n.T("common.status.dirty"), cli.WarningStyle))
	} else {
		parts = append(parts, statusText(i18n.T("cmd.dev.status.clean"), cli.SuccessStyle))
	}

	// Push status
	if len(ahead) > 0 {
		parts = append(parts, statusPart(len(ahead), i18n.T("cmd.dev.health.to_push"), cli.ValueStyle))
	} else {
		parts = append(parts, statusText(i18n.T("common.status.synced"), cli.SuccessStyle))
	}

	// Pull status
	if len(behind) > 0 {
		parts = append(parts, statusPart(len(behind), i18n.T("cmd.dev.health.to_pull"), cli.WarningStyle))
	} else {
		parts = append(parts, statusText(i18n.T("common.status.up_to_date"), cli.SuccessStyle))
	}

	// Errors (only if any)
	if len(errors) > 0 {
		parts = append(parts, statusPart(len(errors), i18n.T("cmd.dev.health.errors"), cli.ErrorStyle))
	}

	cli.Text(statusLine(parts...))
}

func formatRepoList(reposList []string) string {
	if len(reposList) <= 5 {
		return joinRepos(reposList)
	}
	return joinRepos(reposList[:5]) + " " + i18n.T("cmd.dev.health.more", map[string]interface{}{"Count": len(reposList) - 5})
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

func statusPart(count int, label string, style *cli.AnsiStyle) string {
	return style.Render(fmt.Sprintf("%d %s", count, label))
}

func statusText(text string, style *cli.AnsiStyle) string {
	return style.Render(text)
}

func statusLine(parts ...string) string {
	return strings.Join(parts, " | ")
}
