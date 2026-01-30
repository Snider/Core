package dev

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

// Pull command flags
var (
	pullRegistryPath string
	pullAll          bool
)

// addPullCommand adds the 'pull' command to the given parent command.
func addPullCommand(parent *cobra.Command) {
	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: i18n.T("cmd.dev.pull.short"),
		Long:  i18n.T("cmd.dev.pull.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPull(pullRegistryPath, pullAll)
		},
	}

	pullCmd.Flags().StringVar(&pullRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	pullCmd.Flags().BoolVar(&pullAll, "all", false, i18n.T("cmd.dev.pull.flag.all"))

	parent.AddCommand(pullCmd)
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
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.Label("registry")), registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return fmt.Errorf("failed to load registry: %w", err)
			}
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.Label("registry")), registryPath)
		} else {
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.scanning_label")), cwd)
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
		fmt.Println(i18n.T("cmd.dev.no_git_repos"))
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
		fmt.Println(i18n.T("cmd.dev.pull.all_up_to_date"))
		return nil
	}

	// Show what we're pulling
	if all {
		fmt.Printf("\n%s\n\n", i18n.T("cmd.dev.pull.pulling_repos", map[string]interface{}{"Count": len(toPull)}))
	} else {
		fmt.Printf("\n%s\n\n", i18n.T("cmd.dev.pull.repos_behind", map[string]interface{}{"Count": len(toPull)}))
		for _, s := range toPull {
			fmt.Printf("  %s: %s\n",
				repoNameStyle.Render(s.Name),
				dimStyle.Render(i18n.T("cmd.dev.pull.commits_behind", map[string]interface{}{"Count": s.Behind})),
			)
		}
		fmt.Println()
	}

	// Pull each repo
	var succeeded, failed int
	for _, s := range toPull {
		fmt.Printf("  %s %s... ", dimStyle.Render(i18n.T("cmd.dev.pull.pulling")), s.Name)

		err := gitPull(ctx, s.Path)
		if err != nil {
			fmt.Printf("%s\n", errorStyle.Render("x "+err.Error()))
			failed++
		} else {
			fmt.Printf("%s\n", successStyle.Render("v"))
			succeeded++
		}
	}

	// Summary
	fmt.Println()
	fmt.Printf("%s", successStyle.Render(i18n.T("cmd.dev.pull.done_pulled", map[string]interface{}{"Count": succeeded})))
	if failed > 0 {
		fmt.Printf(", %s", errorStyle.Render(i18n.T("common.count.failed", map[string]interface{}{"Count": failed})))
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
