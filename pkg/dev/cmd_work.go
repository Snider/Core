package dev

import (
	"context"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/host-uk/core/pkg/agentic"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
)

// Work command flags
var (
	workStatusOnly   bool
	workAutoCommit   bool
	workRegistryPath string
)

// addWorkCommand adds the 'work' command to the given parent command.
func addWorkCommand(parent *cli.Command) {
	workCmd := &cli.Command{
		Use:   "work",
		Short: i18n.T("cmd.dev.work.short"),
		Long:  i18n.T("cmd.dev.work.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runWork(workRegistryPath, workStatusOnly, workAutoCommit)
		},
	}

	workCmd.Flags().BoolVar(&workStatusOnly, "status", false, i18n.T("cmd.dev.work.flag.status"))
	workCmd.Flags().BoolVar(&workAutoCommit, "commit", false, i18n.T("cmd.dev.work.flag.commit"))
	workCmd.Flags().StringVar(&workRegistryPath, "registry", "", i18n.T("common.flag.registry"))

	parent.AddCommand(workCmd)
}

func runWork(registryPath string, statusOnly, autoCommit bool) error {
	ctx := context.Background()

	// Build worker bundle with required services
	bundle, err := NewWorkBundle(WorkBundleOptions{
		RegistryPath: registryPath,
	})
	if err != nil {
		return err
	}

	// Start services (registers handlers)
	if err := bundle.Start(ctx); err != nil {
		return err
	}
	defer bundle.Stop(ctx)

	// Load registry and get paths
	paths, names, err := loadRegistry(registryPath)
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		cli.Text(i18n.T("cmd.dev.no_git_repos"))
		return nil
	}

	// QUERY git status
	result, handled, err := bundle.Core.QUERY(git.QueryStatus{
		Paths: paths,
		Names: names,
	})
	if !handled {
		return cli.Err("git service not available")
	}
	if err != nil {
		return err
	}
	statuses := result.([]git.RepoStatus)

	// Sort by repo name for consistent output
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	// Display status table
	printStatusTable(statuses)

	// Collect dirty and ahead repos
	var dirtyRepos []git.RepoStatus
	var aheadRepos []git.RepoStatus

	for _, s := range statuses {
		if s.Error != nil {
			continue
		}
		if s.IsDirty() {
			dirtyRepos = append(dirtyRepos, s)
		}
		if s.HasUnpushed() {
			aheadRepos = append(aheadRepos, s)
		}
	}

	// Auto-commit dirty repos if requested
	if autoCommit && len(dirtyRepos) > 0 {
		cli.Line("")
		cli.Print("%s\n", cli.TitleStyle.Render(i18n.T("cmd.dev.commit.committing")))
		cli.Line("")

		for _, s := range dirtyRepos {
			// PERFORM commit via agentic service
			_, handled, err := bundle.Core.PERFORM(agentic.TaskCommit{
				Path: s.Path,
				Name: s.Name,
			})
			if !handled {
				cli.Print("  %s %s: %s\n", warningStyle.Render("!"), s.Name, "agentic service not available")
				continue
			}
			if err != nil {
				cli.Print("  %s %s: %s\n", errorStyle.Render("x"), s.Name, err)
			} else {
				cli.Print("  %s %s\n", successStyle.Render("v"), s.Name)
			}
		}

		// Re-QUERY status after commits
		result, _, _ = bundle.Core.QUERY(git.QueryStatus{
			Paths: paths,
			Names: names,
		})
		statuses = result.([]git.RepoStatus)

		// Rebuild ahead repos list
		aheadRepos = nil
		for _, s := range statuses {
			if s.Error == nil && s.HasUnpushed() {
				aheadRepos = append(aheadRepos, s)
			}
		}
	}

	// If status only, we're done
	if statusOnly {
		if len(dirtyRepos) > 0 && !autoCommit {
			cli.Line("")
			cli.Print("%s\n", dimStyle.Render(i18n.T("cmd.dev.work.use_commit_flag")))
		}
		return nil
	}

	// Push repos with unpushed commits
	if len(aheadRepos) == 0 {
		cli.Line("")
		cli.Text(i18n.T("cmd.dev.work.all_up_to_date"))
		return nil
	}

	cli.Line("")
	cli.Print("%s\n", i18n.T("common.count.repos_unpushed", map[string]interface{}{"Count": len(aheadRepos)}))
	for _, s := range aheadRepos {
		cli.Print("  %s: %s\n", s.Name, i18n.T("common.count.commits", map[string]interface{}{"Count": s.Ahead}))
	}

	cli.Line("")
	if !cli.Confirm(i18n.T("cmd.dev.push.confirm")) {
		cli.Text(i18n.T("cli.aborted"))
		return nil
	}

	cli.Line("")

	// PERFORM push for each repo
	var divergedRepos []git.RepoStatus

	for _, s := range aheadRepos {
		_, handled, err := bundle.Core.PERFORM(git.TaskPush{
			Path: s.Path,
			Name: s.Name,
		})
		if !handled {
			cli.Print("  %s %s: %s\n", errorStyle.Render("x"), s.Name, "git service not available")
			continue
		}
		if err != nil {
			if git.IsNonFastForward(err) {
				cli.Print("  %s %s: %s\n", warningStyle.Render("!"), s.Name, i18n.T("cmd.dev.push.diverged"))
				divergedRepos = append(divergedRepos, s)
			} else {
				cli.Print("  %s %s: %s\n", errorStyle.Render("x"), s.Name, err)
			}
		} else {
			cli.Print("  %s %s\n", successStyle.Render("v"), s.Name)
		}
	}

	// Handle diverged repos - offer to pull and retry
	if len(divergedRepos) > 0 {
		cli.Line("")
		cli.Print("%s\n", i18n.T("cmd.dev.push.diverged_help"))
		if cli.Confirm(i18n.T("cmd.dev.push.pull_and_retry")) {
			cli.Line("")
			for _, s := range divergedRepos {
				cli.Print("  %s %s...\n", dimStyle.Render("↓"), s.Name)

				// PERFORM pull
				_, _, err := bundle.Core.PERFORM(git.TaskPull{Path: s.Path, Name: s.Name})
				if err != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), s.Name, err)
					continue
				}

				cli.Print("  %s %s...\n", dimStyle.Render("↑"), s.Name)

				// PERFORM push
				_, _, err = bundle.Core.PERFORM(git.TaskPush{Path: s.Path, Name: s.Name})
				if err != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), s.Name, err)
					continue
				}

				cli.Print("  %s %s\n", successStyle.Render("v"), s.Name)
			}
		}
	}

	return nil
}

func printStatusTable(statuses []git.RepoStatus) {
	// Calculate column widths
	nameWidth := 4 // "Repo"
	for _, s := range statuses {
		if len(s.Name) > nameWidth {
			nameWidth = len(s.Name)
		}
	}

	// Print header with fixed-width formatting
	cli.Print("%-*s  %8s  %9s  %6s  %5s\n",
		nameWidth,
		cli.TitleStyle.Render(i18n.Label("repo")),
		cli.TitleStyle.Render(i18n.T("cmd.dev.work.table_modified")),
		cli.TitleStyle.Render(i18n.T("cmd.dev.work.table_untracked")),
		cli.TitleStyle.Render(i18n.T("cmd.dev.work.table_staged")),
		cli.TitleStyle.Render(i18n.T("cmd.dev.work.table_ahead")),
	)

	// Print separator
	cli.Text(strings.Repeat("-", nameWidth+2+10+11+8+7))

	// Print rows
	for _, s := range statuses {
		if s.Error != nil {
			paddedName := cli.Sprintf("%-*s", nameWidth, s.Name)
			cli.Print("%s  %s\n",
				repoNameStyle.Render(paddedName),
				errorStyle.Render(i18n.T("cmd.dev.work.error_prefix")+" "+s.Error.Error()),
			)
			continue
		}

		// Style numbers based on values
		modStr := cli.Sprintf("%d", s.Modified)
		if s.Modified > 0 {
			modStr = dirtyStyle.Render(modStr)
		} else {
			modStr = cleanStyle.Render(modStr)
		}

		untrackedStr := cli.Sprintf("%d", s.Untracked)
		if s.Untracked > 0 {
			untrackedStr = dirtyStyle.Render(untrackedStr)
		} else {
			untrackedStr = cleanStyle.Render(untrackedStr)
		}

		stagedStr := cli.Sprintf("%d", s.Staged)
		if s.Staged > 0 {
			stagedStr = aheadStyle.Render(stagedStr)
		} else {
			stagedStr = cleanStyle.Render(stagedStr)
		}

		aheadStr := cli.Sprintf("%d", s.Ahead)
		if s.Ahead > 0 {
			aheadStr = aheadStyle.Render(aheadStr)
		} else {
			aheadStr = cleanStyle.Render(aheadStr)
		}

		// Pad name before styling to avoid ANSI code length issues
		paddedName := cli.Sprintf("%-*s", nameWidth, s.Name)
		cli.Print("%s  %8s  %9s  %6s  %5s\n",
			repoNameStyle.Render(paddedName),
			modStr,
			untrackedStr,
			stagedStr,
			aheadStr,
		)
	}
}

// claudeCommit shells out to claude for committing (legacy helper for other commands)
func claudeCommit(ctx context.Context, repoPath, repoName, registryPath string) error {
	prompt := agentic.Prompt("commit")

	cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--allowedTools", "Bash,Read,Glob,Grep")
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// claudeEditCommit shells out to claude with edit permissions (legacy helper)
func claudeEditCommit(ctx context.Context, repoPath, repoName, registryPath string) error {
	prompt := agentic.Prompt("commit")

	cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--allowedTools", "Bash,Read,Write,Edit,Glob,Grep")
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func loadRegistry(registryPath string) ([]string, map[string]string, error) {
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return nil, nil, cli.Wrap(err, "failed to load registry")
		}
		cli.Print("%s %s\n\n", dimStyle.Render(i18n.Label("registry")), registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return nil, nil, cli.Wrap(err, "failed to load registry")
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.Label("registry")), registryPath)
		} else {
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return nil, nil, cli.Wrap(err, "failed to scan directory")
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.dev.scanning_label")), cwd)
		}
	}

	var paths []string
	names := make(map[string]string)

	for _, repo := range reg.List() {
		if repo.IsGitRepo() {
			paths = append(paths, repo.Path)
			names[repo.Path] = repo.Name
		}
	}

	return paths, names, nil
}
