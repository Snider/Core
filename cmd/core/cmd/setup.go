package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/repos"
	"github.com/leaanthony/clir"
)

// AddSetupCommand adds the 'setup' command to the given parent command.
func AddSetupCommand(parent *clir.Cli) {
	var registryPath string
	var only string
	var dryRun bool

	setupCmd := parent.NewSubCommand("setup", "Clone all repos from registry")
	setupCmd.LongDescription("Clones all repositories defined in repos.yaml into packages/.\n" +
		"Skips repos that already exist. Use --only to filter by type.")

	setupCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)
	setupCmd.StringFlag("only", "Only clone repos of these types (comma-separated: foundation,module,product)", &only)
	setupCmd.BoolFlag("dry-run", "Show what would be cloned without cloning", &dryRun)

	setupCmd.Action(func() error {
		return runSetup(registryPath, only, dryRun)
	})
}

func runSetup(registryPath, only string, dryRun bool) error {
	ctx := context.Background()

	// Find registry
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
	} else {
		registryPath, err = repos.FindRegistry()
		if err != nil {
			return fmt.Errorf("no repos.yaml found - run this from a workspace directory")
		}
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Registry:"), registryPath)
	fmt.Printf("%s %s\n", dimStyle.Render("Org:"), reg.Org)

	// Determine base path for cloning
	basePath := reg.BasePath
	if basePath == "" {
		basePath = "./packages"
	}
	// Resolve relative to registry location
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(filepath.Dir(registryPath), basePath)
	}
	// Expand ~
	if strings.HasPrefix(basePath, "~/") {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, basePath[2:])
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Target:"), basePath)

	// Parse type filter
	var typeFilter map[string]bool
	if only != "" {
		typeFilter = make(map[string]bool)
		for _, t := range strings.Split(only, ",") {
			typeFilter[strings.TrimSpace(t)] = true
		}
		fmt.Printf("%s %s\n", dimStyle.Render("Filter:"), only)
	}

	// Ensure base path exists
	if !dryRun {
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return fmt.Errorf("failed to create packages directory: %w", err)
		}
	}

	// Get repos to clone
	allRepos := reg.List()
	var toClone []*repos.Repo
	var skipped, exists int

	for _, repo := range allRepos {
		// Skip if type filter doesn't match
		if typeFilter != nil && !typeFilter[repo.Type] {
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

	// Summary
	fmt.Println()
	fmt.Printf("%d to clone, %d exist, %d skipped\n", len(toClone), exists, skipped)

	if len(toClone) == 0 {
		fmt.Println("\nNothing to clone.")
		return nil
	}

	if dryRun {
		fmt.Println("\nWould clone:")
		for _, repo := range toClone {
			fmt.Printf("  %s (%s)\n", repoNameStyle.Render(repo.Name), repo.Type)
		}
		return nil
	}

	// Clone repos
	fmt.Println()
	var succeeded, failed int

	for _, repo := range toClone {
		fmt.Printf("  %s %s... ", dimStyle.Render("Cloning"), repo.Name)

		repoPath := filepath.Join(basePath, repo.Name)

		err := gitClone(ctx, reg.Org, repo.Name, repoPath)
		if err != nil {
			fmt.Printf("%s\n", errorStyle.Render("✗ "+err.Error()))
			failed++
		} else {
			fmt.Printf("%s\n", successStyle.Render("✓"))
			succeeded++
		}
	}

	// Summary
	fmt.Println()
	fmt.Printf("%s %d cloned", successStyle.Render("Done:"), succeeded)
	if failed > 0 {
		fmt.Printf(", %s", errorStyle.Render(fmt.Sprintf("%d failed", failed)))
	}
	if exists > 0 {
		fmt.Printf(", %d already exist", exists)
	}
	fmt.Println()

	return nil
}

func gitClone(ctx context.Context, org, repo, path string) error {
	// Try gh clone first with HTTPS (works without SSH keys)
	if ghAuthenticated() {
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

func ghAuthenticated() bool {
	cmd := exec.Command("gh", "auth", "status")
	output, _ := cmd.CombinedOutput()
	return strings.Contains(string(output), "Logged in")
}
