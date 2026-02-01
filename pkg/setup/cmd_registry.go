// cmd_registry.go implements registry mode for cloning packages.
//
// Registry mode is activated when a repos.yaml exists. It reads the registry
// and clones all (or selected) packages into the configured packages directory.

package setup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/host-uk/core/pkg/workspace"
)

// runRegistrySetup loads a registry from path and runs setup.
func runRegistrySetup(ctx context.Context, registryPath, only string, dryRun, all, runBuild bool) error {
	reg, err := repos.LoadRegistry(registryPath)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	return runRegistrySetupWithReg(ctx, reg, registryPath, only, dryRun, all, runBuild)
}

// runRegistrySetupWithReg runs setup with an already-loaded registry.
func runRegistrySetupWithReg(ctx context.Context, reg *repos.Registry, registryPath, only string, dryRun, all, runBuild bool) error {
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.Label("registry")), registryPath)
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.setup.org_label")), reg.Org)

	registryDir := filepath.Dir(registryPath)

	// Determine base path for cloning
	basePath := reg.BasePath
	if basePath == "" {
		// Load workspace config to see if packages_dir is set
		wsConfig, err := workspace.LoadConfig(registryDir)
		if err != nil {
			return fmt.Errorf("failed to load workspace config: %w", err)
		}
		if wsConfig.PackagesDir != "" {
			basePath = wsConfig.PackagesDir
		} else {
			basePath = "./packages"
		}
	}

	// Expand ~
	if strings.HasPrefix(basePath, "~/") {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, basePath[2:])
	}

	// Resolve relative to registry location
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(registryDir, basePath)
	}

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.Label("target")), basePath)

	// Parse type filter
	var typeFilter []string
	if only != "" {
		for _, t := range strings.Split(only, ",") {
			typeFilter = append(typeFilter, strings.TrimSpace(t))
		}
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.Label("filter")), only)
	}

	// Ensure base path exists
	if !dryRun {
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return fmt.Errorf("failed to create packages directory: %w", err)
		}
	}

	// Get all available repos
	allRepos := reg.List()

	// Determine which repos to clone
	var toClone []*repos.Repo
	var skipped, exists int

	// Use wizard in interactive mode, unless --all specified
	useWizard := isTerminal() && !all && !dryRun

	if useWizard {
		selected, err := runPackageWizard(reg, typeFilter)
		if err != nil {
			return fmt.Errorf("wizard error: %w", err)
		}

		// Build set of selected repos
		selectedSet := make(map[string]bool)
		for _, name := range selected {
			selectedSet[name] = true
		}

		// Filter repos based on selection
		for _, repo := range allRepos {
			if !selectedSet[repo.Name] {
				skipped++
				continue
			}

			// Check if already exists
			repoPath := filepath.Join(basePath, repo.Name)
			if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
				exists++
				continue
			}

			toClone = append(toClone, repo)
		}
	} else {
		// Non-interactive: filter by type
		typeFilterSet := make(map[string]bool)
		for _, t := range typeFilter {
			typeFilterSet[t] = true
		}

		for _, repo := range allRepos {
			// Skip if type filter doesn't match (when filter is specified)
			if len(typeFilterSet) > 0 && !typeFilterSet[repo.Type] {
				skipped++
				continue
			}

			// Skip if clone: false
			if repo.Clone != nil && !*repo.Clone {
				skipped++
				continue
			}

			// Check if already exists
			repoPath := filepath.Join(basePath, repo.Name)
			if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
				exists++
				continue
			}

			toClone = append(toClone, repo)
		}
	}

	// Summary
	fmt.Println()
	fmt.Printf("%s, %s, %s\n",
		i18n.T("cmd.setup.to_clone", map[string]interface{}{"Count": len(toClone)}),
		i18n.T("cmd.setup.exist", map[string]interface{}{"Count": exists}),
		i18n.T("common.count.skipped", map[string]interface{}{"Count": skipped}))

	if len(toClone) == 0 {
		fmt.Printf("\n%s\n", i18n.T("cmd.setup.nothing_to_clone"))
		return nil
	}

	if dryRun {
		fmt.Printf("\n%s\n", i18n.T("cmd.setup.would_clone_list"))
		for _, repo := range toClone {
			fmt.Printf("  %s (%s)\n", repoNameStyle.Render(repo.Name), repo.Type)
		}
		return nil
	}

	// Confirm in interactive mode
	if useWizard {
		confirmed, err := confirmClone(len(toClone), basePath)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println(i18n.T("cmd.setup.cancelled"))
			return nil
		}
	}

	// Clone repos
	fmt.Println()
	var succeeded, failed int

	for _, repo := range toClone {
		fmt.Printf("  %s %s... ", dimStyle.Render(i18n.T("common.status.cloning")), repo.Name)

		repoPath := filepath.Join(basePath, repo.Name)

		err := gitClone(ctx, reg.Org, repo.Name, repoPath)
		if err != nil {
			fmt.Printf("%s\n", errorStyle.Render("x "+err.Error()))
			failed++
		} else {
			fmt.Printf("%s\n", successStyle.Render(i18n.T("cmd.setup.done")))
			succeeded++
		}
	}

	// Summary
	fmt.Println()
	fmt.Printf("%s %s", successStyle.Render(i18n.Label("done")), i18n.T("cmd.setup.cloned_count", map[string]interface{}{"Count": succeeded}))
	if failed > 0 {
		fmt.Printf(", %s", errorStyle.Render(i18n.T("i18n.count.failed", failed)))
	}
	if exists > 0 {
		fmt.Printf(", %s", i18n.T("cmd.setup.already_exist_count", map[string]interface{}{"Count": exists}))
	}
	fmt.Println()

	// Run build if requested
	if runBuild && succeeded > 0 {
		fmt.Println()
		fmt.Printf("%s %s\n", dimStyle.Render(">>"), i18n.ProgressSubject("run", "build"))
		buildCmd := exec.Command("core", "build")
		buildCmd.Dir = basePath
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", i18n.T("i18n.fail.run", "build"), err)
		}
	}

	return nil
}

// gitClone clones a repository using gh CLI or git.
func gitClone(ctx context.Context, org, repo, path string) error {
	// Try gh clone first with HTTPS (works without SSH keys)
	if cli.GhAuthenticated() {
		// Use HTTPS URL directly to bypass git_protocol config
		httpsURL := fmt.Sprintf("https://github.com/%s/%s.git", org, repo)
		cmd := exec.CommandContext(ctx, "gh", "repo", "clone", httpsURL, path)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		errStr := strings.TrimSpace(string(output))
		// Only fall through to SSH if it's an auth error
		if !strings.Contains(errStr, "Permission denied") &&
			!strings.Contains(errStr, "could not read") {
			return fmt.Errorf("%s", errStr)
		}
	}

	// Fallback to git clone via SSH
	url := fmt.Sprintf("git@github.com:%s/%s.git", org, repo)
	cmd := exec.CommandContext(ctx, "git", "clone", url, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}
