// cmd_bootstrap.go implements bootstrap mode for new workspaces.
//
// Bootstrap mode is activated when no repos.yaml exists in the current
// directory or any parent. It clones core-devops first, then uses its
// repos.yaml to present the package wizard.

package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/internal/cmd/workspace"
	"github.com/host-uk/core/pkg/i18n"
	coreio "github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/repos"
)

// runSetupOrchestrator decides between registry mode and bootstrap mode.
func runSetupOrchestrator(registryPath, only string, dryRun, all bool, projectName string, runBuild bool) error {
	ctx := context.Background()

	// Try to find an existing registry
	var foundRegistry string
	var err error

	if registryPath != "" {
		foundRegistry = registryPath
	} else {
		foundRegistry, err = repos.FindRegistry()
	}

	// If registry exists, use registry mode
	if err == nil && foundRegistry != "" {
		return runRegistrySetup(ctx, foundRegistry, only, dryRun, all, runBuild)
	}

	// No registry found - enter bootstrap mode
	return runBootstrap(ctx, only, dryRun, all, projectName, runBuild)
}

// runBootstrap handles the case where no repos.yaml exists.
func runBootstrap(ctx context.Context, only string, dryRun, all bool, projectName string, runBuild bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	fmt.Printf("%s %s\n", dimStyle.Render(">>"), i18n.T("cmd.setup.bootstrap_mode"))

	var targetDir string

	// Check if current directory is empty
	empty, err := isDirEmpty(cwd)
	if err != nil {
		return fmt.Errorf("failed to check directory: %w", err)
	}

	if empty {
		// Clone into current directory
		targetDir = cwd
		fmt.Printf("%s %s\n", dimStyle.Render(">>"), i18n.T("cmd.setup.cloning_current_dir"))
	} else {
		// Directory has content - check if it's a git repo root
		isRepo := isGitRepoRoot(cwd)

		if isRepo && isTerminal() && !all {
			// Offer choice: setup working directory or create package
			choice, err := promptSetupChoice()
			if err != nil {
				return fmt.Errorf("failed to get choice: %w", err)
			}

			if choice == "setup" {
				// Setup this working directory with .core/ config
				return runRepoSetup(cwd, dryRun)
			}
			// Otherwise continue to "create package" flow
		}

		// Create package flow - need a project name
		if projectName == "" {
			if !isTerminal() || all {
				projectName = defaultOrg
			} else {
				projectName, err = promptProjectName(defaultOrg)
				if err != nil {
					return fmt.Errorf("failed to get project name: %w", err)
				}
			}
		}

		targetDir = filepath.Join(cwd, projectName)
		fmt.Printf("%s %s: %s\n", dimStyle.Render(">>"), i18n.T("cmd.setup.creating_project_dir"), projectName)

		if !dryRun {
			if err := coreio.Local.EnsureDir(targetDir); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		}
	}

	// Clone core-devops first
	devopsPath := filepath.Join(targetDir, devopsRepo)
	if !coreio.Local.Exists(filepath.Join(devopsPath, ".git")) {
		fmt.Printf("%s %s %s...\n", dimStyle.Render(">>"), i18n.T("common.status.cloning"), devopsRepo)

		if !dryRun {
			if err := gitClone(ctx, defaultOrg, devopsRepo, devopsPath); err != nil {
				return fmt.Errorf("failed to clone %s: %w", devopsRepo, err)
			}
			fmt.Printf("%s %s %s\n", successStyle.Render(">>"), devopsRepo, i18n.T("cmd.setup.cloned"))
		} else {
			fmt.Printf("  %s %s/%s to %s\n", i18n.T("cmd.setup.would_clone"), defaultOrg, devopsRepo, devopsPath)
		}
	} else {
		fmt.Printf("%s %s %s\n", dimStyle.Render(">>"), devopsRepo, i18n.T("cmd.setup.already_exists"))
	}

	// Load the repos.yaml from core-devops
	registryPath := filepath.Join(devopsPath, devopsReposYaml)

	if dryRun {
		fmt.Printf("\n%s %s %s\n", dimStyle.Render(">>"), i18n.T("cmd.setup.would_load_registry"), registryPath)
		return nil
	}

	reg, err := repos.LoadRegistry(registryPath)
	if err != nil {
		return fmt.Errorf("failed to load registry from %s: %w", devopsRepo, err)
	}

	// Override base path to target directory
	reg.BasePath = targetDir

	// Check workspace config for default_only if no filter specified
	if only == "" {
		if wsConfig, err := workspace.LoadConfig(devopsPath); err == nil && wsConfig != nil && len(wsConfig.DefaultOnly) > 0 {
			only = strings.Join(wsConfig.DefaultOnly, ",")
		}
	}

	// Now run the regular setup with the loaded registry
	return runRegistrySetupWithReg(ctx, reg, registryPath, only, dryRun, all, runBuild)
}

// isGitRepoRoot returns true if the directory is a git repository root.
// Handles both regular repos (.git is a directory) and worktrees (.git is a file).
func isGitRepoRoot(path string) bool {
	return coreio.Local.Exists(filepath.Join(path, ".git"))
}

// isDirEmpty returns true if the directory is empty or contains only hidden files.
func isDirEmpty(path string) (bool, error) {
	entries, err := coreio.Local.List(path)
	if err != nil {
		return false, err
	}

	for _, e := range entries {
		name := e.Name()
		// Ignore common hidden/metadata files
		if name == ".DS_Store" || name == ".git" || name == ".gitignore" {
			continue
		}
		// Any other non-hidden file means directory is not empty
		if len(name) > 0 && name[0] != '.' {
			return false, nil
		}
	}

	return true, nil
}
