package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Snider/Core/pkg/repos"
	"github.com/leaanthony/clir"
)

// AddInstallCommand adds the 'install' command to the given parent command.
func AddInstallCommand(parent *clir.Cli) {
	var targetDir string
	var addToRegistry bool
	var repo string

	installCmd := parent.NewSubCommand("install", "Install a repo from GitHub")
	installCmd.LongDescription("Clones a repository from GitHub.\n\n" +
		"Examples:\n" +
		"  core install --repo host-uk/core-php\n" +
		"  core install --repo letheanvpn/lthn-mod-wallet\n" +
		"  core install --repo host-uk/core-tenant --dir ./packages")

	installCmd.StringFlag("repo", "Repository to install (org/repo format)", &repo)
	installCmd.StringFlag("dir", "Target directory (default: ./packages or current dir)", &targetDir)
	installCmd.BoolFlag("add", "Add to repos.yaml registry", &addToRegistry)

	installCmd.Action(func() error {
		if repo == "" {
			return fmt.Errorf("--repo is required (e.g., --repo host-uk/core-php)")
		}
		return runInstall(repo, targetDir, addToRegistry)
	})
}

func runInstall(repoArg, targetDir string, addToRegistry bool) error {
	ctx := context.Background()

	// Parse org/repo
	parts := strings.Split(repoArg, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: use org/repo (e.g., host-uk/core-php)")
	}
	org, repoName := parts[0], parts[1]

	// Determine target directory
	if targetDir == "" {
		// Try to find registry and use its base path
		if regPath, err := repos.FindRegistry(); err == nil {
			if reg, err := repos.LoadRegistry(regPath); err == nil {
				targetDir = reg.BasePath
				if targetDir == "" {
					targetDir = "./packages"
				}
				if !filepath.IsAbs(targetDir) {
					targetDir = filepath.Join(filepath.Dir(regPath), targetDir)
				}
			}
		}

		// Fallback to current directory
		if targetDir == "" {
			targetDir = "."
		}
	}

	// Expand ~ in path
	if strings.HasPrefix(targetDir, "~/") {
		home, _ := os.UserHomeDir()
		targetDir = filepath.Join(home, targetDir[2:])
	}

	repoPath := filepath.Join(targetDir, repoName)

	// Check if already exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		fmt.Printf("%s %s already exists at %s\n", dimStyle.Render("Skip:"), repoName, repoPath)
		return nil
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Printf("%s %s/%s\n", dimStyle.Render("Installing:"), org, repoName)
	fmt.Printf("%s %s\n", dimStyle.Render("Target:"), repoPath)
	fmt.Println()

	// Clone
	fmt.Printf("  %s... ", dimStyle.Render("Cloning"))
	err := gitClone(ctx, org, repoName, repoPath)
	if err != nil {
		fmt.Printf("%s\n", errorStyle.Render("✗ "+err.Error()))
		return err
	}
	fmt.Printf("%s\n", successStyle.Render("✓"))

	// Add to registry if requested
	if addToRegistry {
		if err := addToRegistryFile(org, repoName); err != nil {
			fmt.Printf("  %s add to registry: %s\n", errorStyle.Render("✗"), err)
		} else {
			fmt.Printf("  %s added to repos.yaml\n", successStyle.Render("✓"))
		}
	}

	fmt.Println()
	fmt.Printf("%s Installed %s\n", successStyle.Render("Done:"), repoName)

	return nil
}

func addToRegistryFile(org, repoName string) error {
	regPath, err := repos.FindRegistry()
	if err != nil {
		return fmt.Errorf("no repos.yaml found")
	}

	reg, err := repos.LoadRegistry(regPath)
	if err != nil {
		return err
	}

	// Check if already in registry
	if _, exists := reg.Get(repoName); exists {
		return nil // Already exists
	}

	// Append to file (simple approach - just add YAML at end)
	f, err := os.OpenFile(regPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Detect type from name
	repoType := detectRepoType(repoName)

	entry := fmt.Sprintf("\n  %s:\n    type: %s\n    description: (installed via core install)\n",
		repoName, repoType)

	_, err = f.WriteString(entry)
	return err
}

func detectRepoType(name string) string {
	lower := strings.ToLower(name)

	if strings.Contains(lower, "-mod-") || strings.HasSuffix(lower, "-mod") {
		return "module"
	}
	if strings.Contains(lower, "-plug-") || strings.HasSuffix(lower, "-plug") {
		return "plugin"
	}
	if strings.Contains(lower, "-services-") || strings.HasSuffix(lower, "-services") {
		return "service"
	}
	if strings.Contains(lower, "-website-") || strings.HasSuffix(lower, "-website") {
		return "website"
	}
	if strings.HasPrefix(lower, "core-") {
		return "module"
	}

	return "package"
}
