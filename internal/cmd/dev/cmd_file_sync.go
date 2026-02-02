// cmd_file_sync.go implements safe file synchronization across repos for AI agents.
//
// Usage:
//   core dev sync workflow.yml --to="packages/core-*"
//   core dev sync .github/workflows/ --to="packages/core-*" --message="feat: add CI"
//   core dev sync config.yaml --to="packages/core-*" --dry-run

package dev

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/i18n"
	coreio "github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/log"
	"github.com/host-uk/core/pkg/repos"
)

// File sync command flags
var (
	fileSyncTo       string
	fileSyncMessage  string
	fileSyncCoAuthor string
	fileSyncDryRun   bool
	fileSyncPush     bool
)

// AddFileSyncCommand adds the 'sync' command to dev for file syncing.
func AddFileSyncCommand(parent *cli.Command) {
	syncCmd := &cli.Command{
		Use:   "sync <file-or-dir>",
		Short: i18n.T("cmd.dev.file_sync.short"),
		Long:  i18n.T("cmd.dev.file_sync.long"),
		Args:  cli.MinimumNArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			return runFileSync(args[0])
		},
	}

	syncCmd.Flags().StringVar(&fileSyncTo, "to", "", i18n.T("cmd.dev.file_sync.flag.to"))
	syncCmd.Flags().StringVarP(&fileSyncMessage, "message", "m", "", i18n.T("cmd.dev.file_sync.flag.message"))
	syncCmd.Flags().StringVar(&fileSyncCoAuthor, "co-author", "", i18n.T("cmd.dev.file_sync.flag.co_author"))
	syncCmd.Flags().BoolVar(&fileSyncDryRun, "dry-run", false, i18n.T("cmd.dev.file_sync.flag.dry_run"))
	syncCmd.Flags().BoolVar(&fileSyncPush, "push", false, i18n.T("cmd.dev.file_sync.flag.push"))

	_ = syncCmd.MarkFlagRequired("to")

	parent.AddCommand(syncCmd)
}

func runFileSync(source string) error {
	ctx := context.Background()

	// Security: Reject path traversal attempts
	if strings.Contains(source, "..") {
		return log.E("dev.sync", "path traversal not allowed", nil)
	}

	// Validate source exists
	sourceInfo, err := os.Stat(source) // Keep os.Stat for local source check or use coreio? coreio.Local.IsFile is bool.
	// If source is local file on disk (not in medium), we can use os.Stat.
	// But concept is everything is via Medium?
	// User is running CLI on host. `source` is relative to CWD.
	// coreio.Local uses absolute path or relative to root (which is "/" by default).
	// So coreio.Local works.
	if !coreio.Local.IsFile(source) {
		// Might be directory
		// IsFile returns false for directory.
	}
	// Let's rely on os.Stat for initial source check to distinguish dir vs file easily if coreio doesn't expose Stat.
	// coreio doesn't expose Stat.

	// Check using standard os for source determination as we are outside strict sandbox for input args potentially?
	// But we should use coreio where possible.
	// coreio.Local.List worked for dirs.
	// Let's stick to os.Stat for source properties finding as typically allowed for CLI args.

	if err != nil {
		return log.E("dev.sync", i18n.T("cmd.dev.file_sync.error.source_not_found", map[string]interface{}{"Path": source}), err)
	}

	// Find target repos
	targetRepos, err := resolveTargetRepos(fileSyncTo)
	if err != nil {
		return err
	}

	if len(targetRepos) == 0 {
		return cli.Err("%s", i18n.T("cmd.dev.file_sync.error.no_targets"))
	}

	// Show plan
	cli.Print("%s: %s\n", dimStyle.Render(i18n.T("cmd.dev.file_sync.source")), source)
	cli.Print("%s: %d repos\n", dimStyle.Render(i18n.T("cmd.dev.file_sync.targets")), len(targetRepos))
	if fileSyncDryRun {
		cli.Print("%s\n", warningStyle.Render(i18n.T("cmd.dev.file_sync.dry_run_mode")))
	}
	cli.Blank()

	var succeeded, skipped, failed int

	for _, repo := range targetRepos {
		repoName := filepath.Base(repo.Path)

		if fileSyncDryRun {
			cli.Print("  %s %s\n", dimStyle.Render("[dry-run]"), repoName)
			succeeded++
			continue
		}

		// Step 1: Pull latest (safe sync)
		if err := safePull(ctx, repo.Path); err != nil {
			cli.Print("  %s %s: pull failed: %s\n", errorStyle.Render("x"), repoName, err)
			failed++
			continue
		}

		// Step 2: Copy file(s)
		destPath := filepath.Join(repo.Path, source)
		if sourceInfo.IsDir() {
			if err := copyDir(source, destPath); err != nil {
				cli.Print("  %s %s: copy failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				continue
			}
		} else {
			// Ensure dir exists
			coreio.Local.EnsureDir(filepath.Dir(destPath))
			if err := coreio.Copy(coreio.Local, source, coreio.Local, destPath); err != nil {
				cli.Print("  %s %s: copy failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				continue
			}
		}

		// Step 3: Check if anything changed
		statuses := git.Status(ctx, git.StatusOptions{
			Paths: []string{repo.Path},
			Names: map[string]string{repo.Path: repoName},
		})
		if len(statuses) == 0 || !statuses[0].IsDirty() {
			cli.Print("  %s %s: %s\n", dimStyle.Render("-"), repoName, i18n.T("cmd.dev.file_sync.no_changes"))
			skipped++
			continue
		}

		// Step 4: Commit if message provided
		if fileSyncMessage != "" {
			commitMsg := fileSyncMessage
			if fileSyncCoAuthor != "" {
				commitMsg += "\n\nCo-Authored-By: " + fileSyncCoAuthor
			}

			if err := gitAddCommit(ctx, repo.Path, source, commitMsg); err != nil {
				cli.Print("  %s %s: commit failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				continue
			}

			// Step 5: Push if requested
			if fileSyncPush {
				if err := safePush(ctx, repo.Path); err != nil {
					cli.Print("  %s %s: push failed: %s\n", errorStyle.Render("x"), repoName, err)
					failed++
					continue
				}
			}
		}

		cli.Print("  %s %s\n", successStyle.Render("v"), repoName)
		succeeded++
	}

	// Summary
	cli.Blank()
	cli.Print("%s: ", i18n.T("cmd.dev.file_sync.summary"))
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

// resolveTargetRepos resolves the --to pattern to actual repos
func resolveTargetRepos(pattern string) ([]*repos.Repo, error) {
	// Load registry
	registryPath, err := repos.FindRegistry()
	if err != nil {
		return nil, log.E("dev.sync", "failed to find registry", err)
	}

	registry, err := repos.LoadRegistry(registryPath)
	if err != nil {
		return nil, log.E("dev.sync", "failed to load registry", err)
	}

	// Match pattern against repo names
	var matched []*repos.Repo
	for _, repo := range registry.Repos {
		if matchGlob(repo.Name, pattern) || matchGlob(repo.Path, pattern) {
			matched = append(matched, repo)
		}
	}

	return matched, nil
}

// matchGlob performs simple glob matching with * wildcards
func matchGlob(s, pattern string) bool {
	// Handle exact match
	if s == pattern {
		return true
	}

	// Handle * at end
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(s, prefix)
	}

	// Handle * at start
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(s, suffix)
	}

	// Handle * in middle
	if strings.Contains(pattern, "*") {
		parts := strings.SplitN(pattern, "*", 2)
		return strings.HasPrefix(s, parts[0]) && strings.HasSuffix(s, parts[1])
	}

	return false
}

// safePull pulls with rebase, handling errors gracefully
func safePull(ctx context.Context, path string) error {
	// Check if we have upstream
	_, err := gitCommandQuiet(ctx, path, "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil {
		// No upstream set, skip pull
		return nil
	}

	return git.Pull(ctx, path)
}

// safePush pushes with automatic pull-rebase on rejection
func safePush(ctx context.Context, path string) error {
	err := git.Push(ctx, path)
	if err == nil {
		return nil
	}

	// If non-fast-forward, try pull and push again
	if git.IsNonFastForward(err) {
		if pullErr := git.Pull(ctx, path); pullErr != nil {
			return pullErr
		}
		return git.Push(ctx, path)
	}

	return err
}

// gitAddCommit stages and commits a file/directory
func gitAddCommit(ctx context.Context, repoPath, filePath, message string) error {
	// Stage the file(s)
	if _, err := gitCommandQuiet(ctx, repoPath, "add", filePath); err != nil {
		return err
	}

	// Commit
	_, err := gitCommandQuiet(ctx, repoPath, "commit", "-m", message)
	return err
}

// gitCommandQuiet runs a git command without output
func gitCommandQuiet(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", cli.Err("%s", strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	entries, err := coreio.Local.List(src)
	if err != nil {
		return err
	}

	if err := coreio.Local.EnsureDir(dst); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := coreio.Copy(coreio.Local, srcPath, coreio.Local, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
