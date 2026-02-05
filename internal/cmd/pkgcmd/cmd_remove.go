// cmd_remove.go implements the 'pkg remove' command with safety checks.
//
// Before removing a package, it verifies:
// 1. No uncommitted changes exist
// 2. No unpushed branches exist
// This prevents accidental data loss from agents or tools that might
// attempt to remove packages without cleaning up first.
package pkgcmd

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/i18n"
	coreio "github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

var removeForce bool

func addPkgRemoveCommand(parent *cobra.Command) {
	removeCmd := &cobra.Command{
		Use:   "remove <package>",
		Short: "Remove a package (with safety checks)",
		Long: `Removes a package directory after verifying it has no uncommitted
changes or unpushed branches. Use --force to skip safety checks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New(i18n.T("cmd.pkg.error.repo_required"))
			}
			return runPkgRemove(args[0], removeForce)
		},
	}

	removeCmd.Flags().BoolVar(&removeForce, "force", false, "Skip safety checks (dangerous)")

	parent.AddCommand(removeCmd)
}

func runPkgRemove(name string, force bool) error {
	// Find package path via registry
	regPath, err := repos.FindRegistry(coreio.Local)
	if err != nil {
		return errors.New(i18n.T("cmd.pkg.error.no_repos_yaml"))
	}

	reg, err := repos.LoadRegistry(coreio.Local, regPath)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.load", "registry"), err)
	}

	basePath := reg.BasePath
	if basePath == "" {
		basePath = "."
	}
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(filepath.Dir(regPath), basePath)
	}

	repoPath := filepath.Join(basePath, name)

	if !coreio.Local.IsDir(filepath.Join(repoPath, ".git")) {
		return fmt.Errorf("package %s is not installed at %s", name, repoPath)
	}

	if !force {
		blocked, reasons := checkRepoSafety(repoPath)
		if blocked {
			fmt.Printf("%s Cannot remove %s:\n", errorStyle.Render("Blocked:"), repoNameStyle.Render(name))
			for _, r := range reasons {
				fmt.Printf("  %s %s\n", errorStyle.Render("·"), r)
			}
			fmt.Printf("\nResolve the issues above or use --force to override.\n")
			return errors.New("package has unresolved changes")
		}
	}

	// Remove the directory
	fmt.Printf("%s %s... ", dimStyle.Render("Removing"), repoNameStyle.Render(name))

	if err := coreio.Local.DeleteAll(repoPath); err != nil {
		fmt.Printf("%s\n", errorStyle.Render("x "+err.Error()))
		return err
	}

	fmt.Printf("%s\n", successStyle.Render("ok"))
	return nil
}

// checkRepoSafety checks a git repo for uncommitted changes and unpushed branches.
func checkRepoSafety(repoPath string) (blocked bool, reasons []string) {
	// Check for uncommitted changes (staged, unstaged, untracked)
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		blocked = true
		reasons = append(reasons, fmt.Sprintf("has %d uncommitted changes", len(lines)))
	}

	// Check for unpushed commits on current branch
	cmd = exec.Command("git", "-C", repoPath, "log", "--oneline", "@{u}..HEAD")
	output, err = cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		blocked = true
		reasons = append(reasons, fmt.Sprintf("has %d unpushed commits on current branch", len(lines)))
	}

	// Check all local branches for unpushed work
	cmd = exec.Command("git", "-C", repoPath, "branch", "--no-merged", "origin/HEAD")
	output, _ = cmd.Output()
	if trimmed := strings.TrimSpace(string(output)); trimmed != "" {
		branches := strings.Split(trimmed, "\n")
		var unmerged []string
		for _, b := range branches {
			b = strings.TrimSpace(b)
			b = strings.TrimPrefix(b, "* ")
			if b != "" {
				unmerged = append(unmerged, b)
			}
		}
		if len(unmerged) > 0 {
			blocked = true
			reasons = append(reasons, fmt.Sprintf("has %d unmerged branches: %s",
				len(unmerged), strings.Join(unmerged, ", ")))
		}
	}

	// Check for stashed changes
	cmd = exec.Command("git", "-C", repoPath, "stash", "list")
	output, err = cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		blocked = true
		reasons = append(reasons, fmt.Sprintf("has %d stashed entries", len(lines)))
	}

	return blocked, reasons
}
