// cmd_apply.go implements safe command/script execution across repos for AI agents.
//
// Usage:
//   core dev apply --command="sed -i 's/old/new/g' README.md"
//   core dev apply --script="./scripts/update-version.sh"
//   core dev apply --command="..." --commit --message="chore: update"

package dev

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/errors"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
)

// Apply command flags
var (
	applyCommand  string
	applyScript   string
	applyRepos    string
	applyCommit   bool
	applyMessage  string
	applyCoAuthor string
	applyDryRun   bool
	applyPush     bool
	applyContinue bool // Continue on error
	applyYes      bool // Skip confirmation prompt
)

// AddApplyCommand adds the 'apply' command to dev.
func AddApplyCommand(parent *cli.Command) {
	applyCmd := &cli.Command{
		Use:   "apply",
		Short: i18n.T("cmd.dev.apply.short"),
		Long:  i18n.T("cmd.dev.apply.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runApply()
		},
	}

	applyCmd.Flags().StringVar(&applyCommand, "command", "", i18n.T("cmd.dev.apply.flag.command"))
	applyCmd.Flags().StringVar(&applyScript, "script", "", i18n.T("cmd.dev.apply.flag.script"))
	applyCmd.Flags().StringVar(&applyRepos, "repos", "", i18n.T("cmd.dev.apply.flag.repos"))
	applyCmd.Flags().BoolVar(&applyCommit, "commit", false, i18n.T("cmd.dev.apply.flag.commit"))
	applyCmd.Flags().StringVarP(&applyMessage, "message", "m", "", i18n.T("cmd.dev.apply.flag.message"))
	applyCmd.Flags().StringVar(&applyCoAuthor, "co-author", "", i18n.T("cmd.dev.apply.flag.co_author"))
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, i18n.T("cmd.dev.apply.flag.dry_run"))
	applyCmd.Flags().BoolVar(&applyPush, "push", false, i18n.T("cmd.dev.apply.flag.push"))
	applyCmd.Flags().BoolVar(&applyContinue, "continue", false, i18n.T("cmd.dev.apply.flag.continue"))
	applyCmd.Flags().BoolVarP(&applyYes, "yes", "y", false, i18n.T("cmd.dev.apply.flag.yes"))

	parent.AddCommand(applyCmd)
}

func runApply() error {
	ctx := context.Background()

	// Validate inputs
	if applyCommand == "" && applyScript == "" {
		return errors.E("dev.apply", i18n.T("cmd.dev.apply.error.no_command"), nil)
	}
	if applyCommand != "" && applyScript != "" {
		return errors.E("dev.apply", i18n.T("cmd.dev.apply.error.both_command_script"), nil)
	}
	if applyCommit && applyMessage == "" {
		return errors.E("dev.apply", i18n.T("cmd.dev.apply.error.commit_needs_message"), nil)
	}

	// Validate script exists
	if applyScript != "" {
		if _, err := os.Stat(applyScript); err != nil {
			return errors.E("dev.apply", "script not found: "+applyScript, err)
		}
	}

	// Get target repos
	targetRepos, err := getApplyTargetRepos()
	if err != nil {
		return err
	}

	if len(targetRepos) == 0 {
		return errors.E("dev.apply", i18n.T("cmd.dev.apply.error.no_repos"), nil)
	}

	// Show plan
	action := applyCommand
	if applyScript != "" {
		action = applyScript
	}
	cli.Print("%s: %s\n", dimStyle.Render(i18n.T("cmd.dev.apply.action")), action)
	cli.Print("%s: %d repos\n", dimStyle.Render(i18n.T("cmd.dev.apply.targets")), len(targetRepos))
	if applyDryRun {
		cli.Print("%s\n", warningStyle.Render(i18n.T("cmd.dev.apply.dry_run_mode")))
	}
	cli.Blank()

	// Require confirmation unless --yes or --dry-run
	if !applyYes && !applyDryRun {
		cli.Print("%s\n", warningStyle.Render(i18n.T("cmd.dev.apply.warning")))
		cli.Blank()

		if !cli.Confirm(i18n.T("cmd.dev.apply.confirm"), cli.Required()) {
			cli.Print("%s\n", dimStyle.Render(i18n.T("cmd.dev.apply.cancelled")))
			return nil
		}
		cli.Blank()
	}

	var succeeded, skipped, failed int

	for _, repo := range targetRepos {
		repoName := filepath.Base(repo.Path)

		if applyDryRun {
			cli.Print("  %s %s\n", dimStyle.Render("[dry-run]"), repoName)
			succeeded++
			continue
		}

		// Step 1: Run command or script
		var cmdErr error
		if applyCommand != "" {
			cmdErr = runCommandInRepo(ctx, repo.Path, applyCommand)
		} else {
			cmdErr = runScriptInRepo(ctx, repo.Path, applyScript)
		}

		if cmdErr != nil {
			cli.Print("  %s %s: %s\n", errorStyle.Render("x"), repoName, cmdErr)
			failed++
			if !applyContinue {
				return cli.Err("%s", i18n.T("cmd.dev.apply.error.command_failed"))
			}
			continue
		}

		// Step 2: Check if anything changed
		statuses := git.Status(ctx, git.StatusOptions{
			Paths: []string{repo.Path},
			Names: map[string]string{repo.Path: repoName},
		})
		if len(statuses) == 0 || !statuses[0].IsDirty() {
			cli.Print("  %s %s: %s\n", dimStyle.Render("-"), repoName, i18n.T("cmd.dev.apply.no_changes"))
			skipped++
			continue
		}

		// Step 3: Commit if requested
		if applyCommit {
			commitMsg := applyMessage
			if applyCoAuthor != "" {
				commitMsg += "\n\nCo-Authored-By: " + applyCoAuthor
			}

			// Stage all changes
			if _, err := gitCommandQuiet(ctx, repo.Path, "add", "-A"); err != nil {
				cli.Print("  %s %s: stage failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				if !applyContinue {
					return err
				}
				continue
			}

			// Commit
			if _, err := gitCommandQuiet(ctx, repo.Path, "commit", "-m", commitMsg); err != nil {
				cli.Print("  %s %s: commit failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				if !applyContinue {
					return err
				}
				continue
			}

			// Step 4: Push if requested
			if applyPush {
				if err := safePush(ctx, repo.Path); err != nil {
					cli.Print("  %s %s: push failed: %s\n", errorStyle.Render("x"), repoName, err)
					failed++
					if !applyContinue {
						return err
					}
					continue
				}
			}
		}

		cli.Print("  %s %s\n", successStyle.Render("v"), repoName)
		succeeded++
	}

	// Summary
	cli.Blank()
	cli.Print("%s: ", i18n.T("cmd.dev.apply.summary"))
	if succeeded > 0 {
		cli.Print("%s", successStyle.Render(i18n.T("common.count.succeeded", map[string]interface{}{"Count": succeeded})))
	}
	if skipped > 0 {
		if succeeded > 0 {
			cli.Print(", ")
		}
		cli.Print("%s", dimStyle.Render(i18n.T("common.count.skipped", map[string]interface{}{"Count": skipped})))
	}
	if failed > 0 {
		if succeeded > 0 || skipped > 0 {
			cli.Print(", ")
		}
		cli.Print("%s", errorStyle.Render(i18n.T("common.count.failed", map[string]interface{}{"Count": failed})))
	}
	cli.Blank()

	return nil
}

// getApplyTargetRepos gets repos to apply command to
func getApplyTargetRepos() ([]*repos.Repo, error) {
	// Load registry
	registryPath, err := repos.FindRegistry()
	if err != nil {
		return nil, errors.E("dev.apply", "failed to find registry", err)
	}

	registry, err := repos.LoadRegistry(registryPath)
	if err != nil {
		return nil, errors.E("dev.apply", "failed to load registry", err)
	}

	// If --repos specified, filter to those
	if applyRepos != "" {
		repoNames := strings.Split(applyRepos, ",")
		nameSet := make(map[string]bool)
		for _, name := range repoNames {
			nameSet[strings.TrimSpace(name)] = true
		}

		var matched []*repos.Repo
		for _, repo := range registry.Repos {
			if nameSet[repo.Name] {
				matched = append(matched, repo)
			}
		}
		return matched, nil
	}

	// Return all repos as slice
	var all []*repos.Repo
	for _, repo := range registry.Repos {
		all = append(all, repo)
	}
	return all, nil
}

// runCommandInRepo runs a shell command in a repo directory
func runCommandInRepo(ctx context.Context, repoPath, command string) error {
	// Use shell to execute command
	var cmd *exec.Cmd
	if isWindows() {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// runScriptInRepo runs a script in a repo directory
func runScriptInRepo(ctx context.Context, repoPath, scriptPath string) error {
	// Get absolute path to script
	absScript, err := filepath.Abs(scriptPath)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	if isWindows() {
		cmd = exec.CommandContext(ctx, "cmd", "/C", absScript)
	} else {
		// Execute script directly to honor shebang
		cmd = exec.CommandContext(ctx, absScript)
	}
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// isWindows returns true if running on Windows
func isWindows() bool {
	return os.PathSeparator == '\\'
}
