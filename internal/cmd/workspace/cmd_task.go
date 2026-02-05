// cmd_task.go implements task workspace isolation using git worktrees.
//
// Each task gets an isolated workspace at .core/workspace/p{epic}/i{issue}/
// containing git worktrees of required repos. This prevents agents from
// writing to the implementor's working tree.
//
// Safety checks enforce that workspaces cannot be removed if they contain
// uncommitted changes or unpushed branches.
package workspace

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	coreio "github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

var (
	taskEpic    int
	taskIssue   int
	taskRepos   []string
	taskForce   bool
	taskBranch  string
)

func addTaskCommands(parent *cobra.Command) {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage isolated task workspaces for agents",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an isolated task workspace with git worktrees",
		Long: `Creates a workspace at .core/workspace/p{epic}/i{issue}/ with git
worktrees for each specified repo. Each worktree gets a fresh branch
(issue/{id} by default) so agents work in isolation.`,
		RunE: runTaskCreate,
	}
	createCmd.Flags().IntVar(&taskEpic, "epic", 0, "Epic/project number")
	createCmd.Flags().IntVar(&taskIssue, "issue", 0, "Issue number")
	createCmd.Flags().StringSliceVar(&taskRepos, "repo", nil, "Repos to include (default: all from registry)")
	createCmd.Flags().StringVar(&taskBranch, "branch", "", "Branch name (default: issue/{issue})")
	_ = createCmd.MarkFlagRequired("epic")
	_ = createCmd.MarkFlagRequired("issue")

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a task workspace (with safety checks)",
		Long: `Removes a task workspace after checking for uncommitted changes and
unpushed branches. Use --force to skip safety checks.`,
		RunE: runTaskRemove,
	}
	removeCmd.Flags().IntVar(&taskEpic, "epic", 0, "Epic/project number")
	removeCmd.Flags().IntVar(&taskIssue, "issue", 0, "Issue number")
	removeCmd.Flags().BoolVar(&taskForce, "force", false, "Skip safety checks")
	_ = removeCmd.MarkFlagRequired("epic")
	_ = removeCmd.MarkFlagRequired("issue")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all task workspaces",
		RunE:  runTaskList,
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of a task workspace",
		RunE:  runTaskStatus,
	}
	statusCmd.Flags().IntVar(&taskEpic, "epic", 0, "Epic/project number")
	statusCmd.Flags().IntVar(&taskIssue, "issue", 0, "Issue number")
	_ = statusCmd.MarkFlagRequired("epic")
	_ = statusCmd.MarkFlagRequired("issue")

	addAgentCommands(taskCmd)

	taskCmd.AddCommand(createCmd, removeCmd, listCmd, statusCmd)
	parent.AddCommand(taskCmd)
}

// taskWorkspacePath returns the path for a task workspace.
func taskWorkspacePath(root string, epic, issue int) string {
	return filepath.Join(root, ".core", "workspace", fmt.Sprintf("p%d", epic), fmt.Sprintf("i%d", issue))
}

func runTaskCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	root, err := FindWorkspaceRoot()
	if err != nil {
		return cli.Err("not in a workspace — run from workspace root or a package directory")
	}

	wsPath := taskWorkspacePath(root, taskEpic, taskIssue)

	if coreio.Local.IsDir(wsPath) {
		return cli.Err("task workspace already exists: %s", wsPath)
	}

	branch := taskBranch
	if branch == "" {
		branch = fmt.Sprintf("issue/%d", taskIssue)
	}

	// Determine repos to include
	repoNames := taskRepos
	if len(repoNames) == 0 {
		repoNames, err = registryRepoNames(root)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
	}

	if len(repoNames) == 0 {
		return cli.Err("no repos specified and no registry found")
	}

	// Resolve package paths
	config, _ := LoadConfig(root)
	pkgDir := "./packages"
	if config != nil && config.PackagesDir != "" {
		pkgDir = config.PackagesDir
	}
	if !filepath.IsAbs(pkgDir) {
		pkgDir = filepath.Join(root, pkgDir)
	}

	if err := coreio.Local.EnsureDir(wsPath); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	cli.Print("Creating task workspace: %s\n", cli.ValueStyle.Render(fmt.Sprintf("p%d/i%d", taskEpic, taskIssue)))
	cli.Print("Branch: %s\n", cli.ValueStyle.Render(branch))
	cli.Print("Path:   %s\n\n", cli.DimStyle.Render(wsPath))

	var created, skipped int
	for _, repoName := range repoNames {
		repoPath := filepath.Join(pkgDir, repoName)
		if !coreio.Local.IsDir(filepath.Join(repoPath, ".git")) {
			cli.Print("  %s %s (not cloned, skipping)\n", cli.DimStyle.Render("·"), repoName)
			skipped++
			continue
		}

		worktreePath := filepath.Join(wsPath, repoName)
		cli.Print("  %s %s... ", cli.DimStyle.Render("·"), repoName)

		if err := createWorktree(ctx, repoPath, worktreePath, branch); err != nil {
			cli.Print("%s\n", cli.ErrorStyle.Render("x "+err.Error()))
			skipped++
			continue
		}

		cli.Print("%s\n", cli.SuccessStyle.Render("ok"))
		created++
	}

	cli.Print("\n%s %d worktrees created", cli.SuccessStyle.Render("Done:"), created)
	if skipped > 0 {
		cli.Print(", %d skipped", skipped)
	}
	cli.Print("\n")

	return nil
}

func runTaskRemove(cmd *cobra.Command, args []string) error {
	root, err := FindWorkspaceRoot()
	if err != nil {
		return cli.Err("not in a workspace")
	}

	wsPath := taskWorkspacePath(root, taskEpic, taskIssue)
	if !coreio.Local.IsDir(wsPath) {
		return cli.Err("task workspace does not exist: p%d/i%d", taskEpic, taskIssue)
	}

	if !taskForce {
		dirty, reasons := checkWorkspaceSafety(wsPath)
		if dirty {
			cli.Print("%s Cannot remove workspace p%d/i%d:\n", cli.ErrorStyle.Render("Blocked:"), taskEpic, taskIssue)
			for _, r := range reasons {
				cli.Print("  %s %s\n", cli.ErrorStyle.Render("·"), r)
			}
			cli.Print("\nUse --force to override or resolve the issues first.\n")
			return errors.New("workspace has unresolved changes")
		}
	}

	// Remove worktrees first (so git knows they're gone)
	entries, err := coreio.Local.List(wsPath)
	if err != nil {
		return fmt.Errorf("failed to list workspace: %w", err)
	}

	config, _ := LoadConfig(root)
	pkgDir := "./packages"
	if config != nil && config.PackagesDir != "" {
		pkgDir = config.PackagesDir
	}
	if !filepath.IsAbs(pkgDir) {
		pkgDir = filepath.Join(root, pkgDir)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		worktreePath := filepath.Join(wsPath, entry.Name())
		repoPath := filepath.Join(pkgDir, entry.Name())

		// Remove worktree from git
		if coreio.Local.IsDir(filepath.Join(repoPath, ".git")) {
			removeWorktree(repoPath, worktreePath)
		}
	}

	// Remove the workspace directory
	if err := coreio.Local.DeleteAll(wsPath); err != nil {
		return fmt.Errorf("failed to remove workspace directory: %w", err)
	}

	// Clean up empty parent (p{epic}/) if it's now empty
	epicDir := filepath.Dir(wsPath)
	if entries, err := coreio.Local.List(epicDir); err == nil && len(entries) == 0 {
		coreio.Local.DeleteAll(epicDir)
	}

	cli.Print("%s Removed workspace p%d/i%d\n", cli.SuccessStyle.Render("Done:"), taskEpic, taskIssue)
	return nil
}

func runTaskList(cmd *cobra.Command, args []string) error {
	root, err := FindWorkspaceRoot()
	if err != nil {
		return cli.Err("not in a workspace")
	}

	wsRoot := filepath.Join(root, ".core", "workspace")
	if !coreio.Local.IsDir(wsRoot) {
		cli.Println("No task workspaces found.")
		return nil
	}

	epics, err := coreio.Local.List(wsRoot)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	found := false
	for _, epicEntry := range epics {
		if !epicEntry.IsDir() || !strings.HasPrefix(epicEntry.Name(), "p") {
			continue
		}
		epicDir := filepath.Join(wsRoot, epicEntry.Name())
		issues, err := coreio.Local.List(epicDir)
		if err != nil {
			continue
		}
		for _, issueEntry := range issues {
			if !issueEntry.IsDir() || !strings.HasPrefix(issueEntry.Name(), "i") {
				continue
			}
			found = true
			wsPath := filepath.Join(epicDir, issueEntry.Name())

			// Count worktrees
			entries, _ := coreio.Local.List(wsPath)
			dirCount := 0
			for _, e := range entries {
				if e.IsDir() {
					dirCount++
				}
			}

			// Check safety
			dirty, _ := checkWorkspaceSafety(wsPath)
			status := cli.SuccessStyle.Render("clean")
			if dirty {
				status = cli.ErrorStyle.Render("dirty")
			}

			cli.Print("  %s/%s  %d repos  %s\n",
				epicEntry.Name(), issueEntry.Name(),
				dirCount, status)
		}
	}

	if !found {
		cli.Println("No task workspaces found.")
	}

	return nil
}

func runTaskStatus(cmd *cobra.Command, args []string) error {
	root, err := FindWorkspaceRoot()
	if err != nil {
		return cli.Err("not in a workspace")
	}

	wsPath := taskWorkspacePath(root, taskEpic, taskIssue)
	if !coreio.Local.IsDir(wsPath) {
		return cli.Err("task workspace does not exist: p%d/i%d", taskEpic, taskIssue)
	}

	cli.Print("Workspace: %s\n", cli.ValueStyle.Render(fmt.Sprintf("p%d/i%d", taskEpic, taskIssue)))
	cli.Print("Path:      %s\n\n", cli.DimStyle.Render(wsPath))

	entries, err := coreio.Local.List(wsPath)
	if err != nil {
		return fmt.Errorf("failed to list workspace: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		worktreePath := filepath.Join(wsPath, entry.Name())

		// Get branch
		branch := gitOutput(worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
		branch = strings.TrimSpace(branch)

		// Get status
		status := gitOutput(worktreePath, "status", "--porcelain")
		statusLabel := cli.SuccessStyle.Render("clean")
		if strings.TrimSpace(status) != "" {
			lines := len(strings.Split(strings.TrimSpace(status), "\n"))
			statusLabel = cli.ErrorStyle.Render(fmt.Sprintf("%d changes", lines))
		}

		// Get unpushed
		unpushed := gitOutput(worktreePath, "log", "--oneline", "@{u}..HEAD")
		unpushedLabel := ""
		if trimmed := strings.TrimSpace(unpushed); trimmed != "" {
			count := len(strings.Split(trimmed, "\n"))
			unpushedLabel = cli.WarningStyle.Render(fmt.Sprintf("  %d unpushed", count))
		}

		cli.Print("  %s  %s  %s%s\n",
			cli.RepoStyle.Render(entry.Name()),
			cli.DimStyle.Render(branch),
			statusLabel,
			unpushedLabel)
	}

	return nil
}

// createWorktree adds a git worktree at worktreePath for the given branch.
func createWorktree(ctx context.Context, repoPath, worktreePath, branch string) error {
	// Check if branch exists on remote first
	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", branch, worktreePath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		errStr := strings.TrimSpace(string(output))
		// If branch already exists, try without -b
		if strings.Contains(errStr, "already exists") {
			cmd = exec.CommandContext(ctx, "git", "worktree", "add", worktreePath, branch)
			cmd.Dir = repoPath
			output, err = cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("%s", strings.TrimSpace(string(output)))
			}
			return nil
		}
		return fmt.Errorf("%s", errStr)
	}
	return nil
}

// removeWorktree removes a git worktree.
func removeWorktree(repoPath, worktreePath string) {
	cmd := exec.Command("git", "worktree", "remove", worktreePath)
	cmd.Dir = repoPath
	_ = cmd.Run()

	// Prune stale worktrees
	cmd = exec.Command("git", "worktree", "prune")
	cmd.Dir = repoPath
	_ = cmd.Run()
}

// checkWorkspaceSafety checks all worktrees in a workspace for uncommitted/unpushed changes.
func checkWorkspaceSafety(wsPath string) (dirty bool, reasons []string) {
	entries, err := coreio.Local.List(wsPath)
	if err != nil {
		return false, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		worktreePath := filepath.Join(wsPath, entry.Name())

		// Check for uncommitted changes
		status := gitOutput(worktreePath, "status", "--porcelain")
		if strings.TrimSpace(status) != "" {
			dirty = true
			reasons = append(reasons, fmt.Sprintf("%s: has uncommitted changes", entry.Name()))
		}

		// Check for unpushed commits
		unpushed := gitOutput(worktreePath, "log", "--oneline", "@{u}..HEAD")
		if strings.TrimSpace(unpushed) != "" {
			dirty = true
			count := len(strings.Split(strings.TrimSpace(unpushed), "\n"))
			reasons = append(reasons, fmt.Sprintf("%s: %d unpushed commits", entry.Name(), count))
		}
	}

	return dirty, reasons
}

// gitOutput runs a git command and returns stdout.
func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, _ := cmd.Output()
	return string(out)
}

// registryRepoNames returns repo names from the workspace registry.
func registryRepoNames(root string) ([]string, error) {
	// Try to find repos.yaml
	regPath, err := repos.FindRegistry(coreio.Local)
	if err != nil {
		return nil, err
	}

	reg, err := repos.LoadRegistry(coreio.Local, regPath)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, repo := range reg.List() {
		// Only include cloneable repos
		if repo.Clone != nil && !*repo.Clone {
			continue
		}
		// Skip meta repos
		if repo.Type == "meta" {
			continue
		}
		names = append(names, repo.Name)
	}

	return names, nil
}

// epicBranchName returns the branch name for an EPIC.
func epicBranchName(epicID int) string {
	return "epic/" + strconv.Itoa(epicID)
}
