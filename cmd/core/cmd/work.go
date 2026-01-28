package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/repos"
	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
)

var (
	// Table styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3b82f6")). // blue-500
			Padding(0, 1)

	cellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	dirtyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")). // red-500
			Padding(0, 1)

	aheadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22c55e")). // green-500
			Padding(0, 1)

	cleanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")). // gray-500
			Padding(0, 1)

	repoNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#e2e8f0")). // gray-200
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22c55e")). // green-500
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")). // red-500
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")) // gray-500
)

// AddWorkCommand adds the 'work' command to the given parent command.
func AddWorkCommand(parent *clir.Cli) {
	var statusOnly bool
	var autoCommit bool
	var registryPath string

	workCmd := parent.NewSubCommand("work", "Multi-repo git operations")
	workCmd.LongDescription("Manage git status, commits, and pushes across multiple repositories.\n\n" +
		"Reads repos.yaml to discover repositories and their relationships.\n" +
		"Shows status, optionally commits with Claude, and pushes changes.")

	workCmd.BoolFlag("status", "Show status only, don't push", &statusOnly)
	workCmd.BoolFlag("commit", "Use Claude to commit dirty repos before pushing", &autoCommit)
	workCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)

	workCmd.Action(func() error {
		return runWork(registryPath, statusOnly, autoCommit)
	})
}

func runWork(registryPath string, statusOnly, autoCommit bool) error {
	ctx := context.Background()

	// Find or use provided registry, fall back to directory scan
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
		fmt.Printf("%s %s\n\n", dimStyle.Render("Registry:"), registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return fmt.Errorf("failed to load registry: %w", err)
			}
			fmt.Printf("%s %s\n\n", dimStyle.Render("Registry:"), registryPath)
		} else {
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}
			fmt.Printf("%s %s\n\n", dimStyle.Render("Scanning:"), cwd)
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
		fmt.Println("No git repositories found.")
		return nil
	}

	// Get status for all repos
	statuses := git.Status(ctx, git.StatusOptions{
		Paths: paths,
		Names: names,
	})

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
		fmt.Println()
		fmt.Printf("%s\n", headerStyle.Render("Committing dirty repos with Claude..."))
		fmt.Println()

		for _, s := range dirtyRepos {
			if err := claudeCommit(ctx, s.Path, s.Name, registryPath); err != nil {
				fmt.Printf("  %s %s: %s\n", errorStyle.Render("✗"), s.Name, err)
			} else {
				fmt.Printf("  %s %s\n", successStyle.Render("✓"), s.Name)
			}
		}

		// Re-check status after commits
		statuses = git.Status(ctx, git.StatusOptions{
			Paths: paths,
			Names: names,
		})

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
			fmt.Println()
			fmt.Printf("%s\n", dimStyle.Render("Use --commit to have Claude create commits"))
		}
		return nil
	}

	// Push repos with unpushed commits
	if len(aheadRepos) == 0 {
		fmt.Println()
		fmt.Println("All repos up to date.")
		return nil
	}

	fmt.Println()
	fmt.Printf("%d repo(s) with unpushed commits:\n", len(aheadRepos))
	for _, s := range aheadRepos {
		fmt.Printf("  %s: %d commit(s)\n", s.Name, s.Ahead)
	}

	fmt.Println()
	if !confirm("Push all?") {
		fmt.Println("Aborted.")
		return nil
	}

	fmt.Println()

	// Push sequentially (SSH passphrase needs interaction)
	var pushPaths []string
	for _, s := range aheadRepos {
		pushPaths = append(pushPaths, s.Path)
	}

	results := git.PushMultiple(ctx, pushPaths, names)

	for _, r := range results {
		if r.Success {
			fmt.Printf("  %s %s\n", successStyle.Render("✓"), r.Name)
		} else {
			fmt.Printf("  %s %s: %s\n", errorStyle.Render("✗"), r.Name, r.Error)
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
	hdrStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3b82f6"))
	fmt.Printf("%-*s  %8s  %9s  %6s  %5s\n",
		nameWidth,
		hdrStyle.Render("Repo"),
		hdrStyle.Render("Modified"),
		hdrStyle.Render("Untracked"),
		hdrStyle.Render("Staged"),
		hdrStyle.Render("Ahead"),
	)

	// Print separator
	fmt.Println(strings.Repeat("─", nameWidth+2+10+11+8+7))

	// Print rows
	for _, s := range statuses {
		if s.Error != nil {
			paddedName := fmt.Sprintf("%-*s", nameWidth, s.Name)
			fmt.Printf("%s  %s\n",
				repoNameStyle.Render(paddedName),
				errorStyle.Render("error: "+s.Error.Error()),
			)
			continue
		}

		// Style numbers based on values
		modStr := fmt.Sprintf("%d", s.Modified)
		if s.Modified > 0 {
			modStr = dirtyStyle.Render(modStr)
		} else {
			modStr = cleanStyle.Render(modStr)
		}

		untrackedStr := fmt.Sprintf("%d", s.Untracked)
		if s.Untracked > 0 {
			untrackedStr = dirtyStyle.Render(untrackedStr)
		} else {
			untrackedStr = cleanStyle.Render(untrackedStr)
		}

		stagedStr := fmt.Sprintf("%d", s.Staged)
		if s.Staged > 0 {
			stagedStr = aheadStyle.Render(stagedStr)
		} else {
			stagedStr = cleanStyle.Render(stagedStr)
		}

		aheadStr := fmt.Sprintf("%d", s.Ahead)
		if s.Ahead > 0 {
			aheadStr = aheadStyle.Render(aheadStr)
		} else {
			aheadStr = cleanStyle.Render(aheadStr)
		}

		// Pad name before styling to avoid ANSI code length issues
		paddedName := fmt.Sprintf("%-*s", nameWidth, s.Name)
		fmt.Printf("%s  %8s  %9s  %6s  %5s\n",
			repoNameStyle.Render(paddedName),
			modStr,
			untrackedStr,
			stagedStr,
			aheadStr,
		)
	}
}

func claudeCommit(ctx context.Context, repoPath, repoName, registryPath string) error {
	// Load AGENTS.md context if available
	agentsPath := filepath.Join(filepath.Dir(registryPath), "AGENTS.md")
	var context string
	if data, err := os.ReadFile(agentsPath); err == nil {
		context = string(data) + "\n\n"
	}

	prompt := context + "Review the uncommitted changes and create an appropriate commit. " +
		"Use Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>. Be concise."

	cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--allowedTools", "Bash,Read,Glob,Grep")
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
