package pkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

var (
	installTargetDir    string
	installAddToReg     bool
)

// addPkgInstallCommand adds the 'pkg install' command.
func addPkgInstallCommand(parent *cobra.Command) {
	installCmd := &cobra.Command{
		Use:   "install <org/repo>",
		Short: "Clone a package from GitHub",
		Long: "Clones a repository from GitHub.\n\n" +
			"Examples:\n" +
			"  core pkg install host-uk/core-php\n" +
			"  core pkg install host-uk/core-tenant --dir ./packages\n" +
			"  core pkg install host-uk/core-admin --add",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("repository is required (e.g., core pkg install host-uk/core-php)")
			}
			return runPkgInstall(args[0], installTargetDir, installAddToReg)
		},
	}

	installCmd.Flags().StringVar(&installTargetDir, "dir", "", "Target directory (default: ./packages or current dir)")
	installCmd.Flags().BoolVar(&installAddToReg, "add", false, "Add to repos.yaml registry")

	parent.AddCommand(installCmd)
}

func runPkgInstall(repoArg, targetDir string, addToRegistry bool) error {
	ctx := context.Background()

	// Parse org/repo
	parts := strings.Split(repoArg, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: use org/repo (e.g., host-uk/core-php)")
	}
	org, repoName := parts[0], parts[1]

	// Determine target directory
	if targetDir == "" {
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
		if targetDir == "" {
			targetDir = "."
		}
	}

	if strings.HasPrefix(targetDir, "~/") {
		home, _ := os.UserHomeDir()
		targetDir = filepath.Join(home, targetDir[2:])
	}

	repoPath := filepath.Join(targetDir, repoName)

	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		fmt.Printf("%s %s already exists at %s\n", dimStyle.Render("Skip:"), repoName, repoPath)
		return nil
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Printf("%s %s/%s\n", dimStyle.Render("Installing:"), org, repoName)
	fmt.Printf("%s %s\n", dimStyle.Render("Target:"), repoPath)
	fmt.Println()

	fmt.Printf("  %s... ", dimStyle.Render("Cloning"))
	err := gitClone(ctx, org, repoName, repoPath)
	if err != nil {
		fmt.Printf("%s\n", errorStyle.Render("✗ "+err.Error()))
		return err
	}
	fmt.Printf("%s\n", successStyle.Render("✓"))

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

	if _, exists := reg.Get(repoName); exists {
		return nil
	}

	f, err := os.OpenFile(regPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	repoType := detectRepoType(repoName)
	entry := fmt.Sprintf("\n  %s:\n    type: %s\n    description: (installed via core pkg install)\n",
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
		return "package"
	}
	return "package"
}
