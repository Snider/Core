// Package setup provides workspace setup and bootstrap commands.
package setup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/repos"
	"github.com/leaanthony/clir"
)

// Style aliases
var (
	repoNameStyle = shared.RepoNameStyle
	successStyle  = shared.SuccessStyle
	errorStyle    = shared.ErrorStyle
	dimStyle      = shared.DimStyle
)

// Default organization and devops repo for bootstrap
const (
	defaultOrg        = "host-uk"
	devopsRepo        = "core-devops"
	devopsReposYaml   = "repos.yaml"
)

// AddSetupCommand adds the 'setup' command to the given parent command.
func AddSetupCommand(parent *clir.Cli) {
	var registryPath string
	var only string
	var dryRun bool
	var all bool
	var name string
	var build bool

	setupCmd := parent.NewSubCommand("setup", "Bootstrap workspace or clone packages from registry")
	setupCmd.LongDescription("Sets up a development workspace.\n\n" +
		"REGISTRY MODE (repos.yaml exists):\n" +
		"  Clones all repositories defined in repos.yaml into packages/.\n" +
		"  Skips repos that already exist. Use --only to filter by type.\n\n" +
		"BOOTSTRAP MODE (no repos.yaml):\n" +
		"  1. Clones core-devops to set up the workspace\n" +
		"  2. Presents an interactive wizard to select packages\n" +
		"  3. Clones selected packages\n\n" +
		"Use --all to skip the wizard and clone everything.")

	setupCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)
	setupCmd.StringFlag("only", "Only clone repos of these types (comma-separated: foundation,module,product)", &only)
	setupCmd.BoolFlag("dry-run", "Show what would be cloned without cloning", &dryRun)
	setupCmd.BoolFlag("all", "Skip wizard, clone all packages (non-interactive)", &all)
	setupCmd.StringFlag("name", "Project directory name for bootstrap mode", &name)
	setupCmd.BoolFlag("build", "Run build after cloning", &build)

	setupCmd.Action(func() error {
		return runSetupOrchestrator(registryPath, only, dryRun, all, name, build)
	})
}

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

	fmt.Printf("%s Bootstrap mode (no repos.yaml found)\n", dimStyle.Render(">>"))

	var targetDir string

	// Check if current directory is empty
	empty, err := isDirEmpty(cwd)
	if err != nil {
		return fmt.Errorf("failed to check directory: %w", err)
	}

	if empty {
		// Clone into current directory
		targetDir = cwd
		fmt.Printf("%s Cloning into current directory\n", dimStyle.Render(">>"))
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
		fmt.Printf("%s Creating project directory: %s\n", dimStyle.Render(">>"), projectName)

		if !dryRun {
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		}
	}

	// Clone core-devops first
	devopsPath := filepath.Join(targetDir, devopsRepo)
	if _, err := os.Stat(filepath.Join(devopsPath, ".git")); os.IsNotExist(err) {
		fmt.Printf("%s Cloning %s...\n", dimStyle.Render(">>"), devopsRepo)

		if !dryRun {
			if err := gitClone(ctx, defaultOrg, devopsRepo, devopsPath); err != nil {
				return fmt.Errorf("failed to clone %s: %w", devopsRepo, err)
			}
			fmt.Printf("%s %s cloned\n", successStyle.Render(">>"), devopsRepo)
		} else {
			fmt.Printf("  Would clone %s/%s to %s\n", defaultOrg, devopsRepo, devopsPath)
		}
	} else {
		fmt.Printf("%s %s already exists\n", dimStyle.Render(">>"), devopsRepo)
	}

	// Load the repos.yaml from core-devops
	registryPath := filepath.Join(devopsPath, devopsReposYaml)

	if dryRun {
		fmt.Printf("\n%s Would load registry from %s and present package wizard\n", dimStyle.Render(">>"), registryPath)
		return nil
	}

	reg, err := repos.LoadRegistry(registryPath)
	if err != nil {
		return fmt.Errorf("failed to load registry from %s: %w", devopsRepo, err)
	}

	// Override base path to target directory
	reg.BasePath = targetDir

	// Now run the regular setup with the loaded registry
	return runRegistrySetupWithReg(ctx, reg, registryPath, only, dryRun, all, runBuild)
}

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
	var typeFilter []string
	if only != "" {
		for _, t := range strings.Split(only, ",") {
			typeFilter = append(typeFilter, strings.TrimSpace(t))
		}
		fmt.Printf("%s %s\n", dimStyle.Render("Filter:"), only)
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

	// Confirm in interactive mode
	if useWizard {
		confirmed, err := confirmClone(len(toClone), basePath)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Clone repos
	fmt.Println()
	var succeeded, failed int

	for _, repo := range toClone {
		fmt.Printf("  %s %s... ", dimStyle.Render("Cloning"), repo.Name)

		repoPath := filepath.Join(basePath, repo.Name)

		err := gitClone(ctx, reg.Org, repo.Name, repoPath)
		if err != nil {
			fmt.Printf("%s\n", errorStyle.Render("x "+err.Error()))
			failed++
		} else {
			fmt.Printf("%s\n", successStyle.Render("done"))
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

	// Run build if requested
	if runBuild && succeeded > 0 {
		fmt.Println()
		fmt.Printf("%s Running build...\n", dimStyle.Render(">>"))
		buildCmd := exec.Command("core", "build")
		buildCmd.Dir = basePath
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
	}

	return nil
}

// isGitRepoRoot returns true if the directory is a git repository root.
func isGitRepoRoot(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

// runRepoSetup sets up the current repository with .core/ configuration.
func runRepoSetup(repoPath string, dryRun bool) error {
	fmt.Printf("%s Setting up repository: %s\n", dimStyle.Render(">>"), repoPath)

	// Detect project type
	projectType := detectProjectType(repoPath)
	fmt.Printf("%s Detected project type: %s\n", dimStyle.Render(">>"), projectType)

	// Create .core directory
	coreDir := filepath.Join(repoPath, ".core")
	if !dryRun {
		if err := os.MkdirAll(coreDir, 0755); err != nil {
			return fmt.Errorf("failed to create .core directory: %w", err)
		}
	}

	// Generate configs based on project type
	name := filepath.Base(repoPath)
	configs := map[string]string{
		"build.yaml":   generateBuildConfig(repoPath, projectType),
		"release.yaml": generateReleaseConfig(name, projectType),
		"test.yaml":    generateTestConfig(projectType),
	}

	if dryRun {
		fmt.Printf("\n%s Would create:\n", dimStyle.Render(">>"))
		for filename, content := range configs {
			fmt.Printf("\n  %s:\n", filepath.Join(coreDir, filename))
			// Indent content for display
			for _, line := range strings.Split(content, "\n") {
				fmt.Printf("    %s\n", line)
			}
		}
		return nil
	}

	for filename, content := range configs {
		configPath := filepath.Join(coreDir, filename)
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
		fmt.Printf("%s Created %s\n", successStyle.Render(">>"), configPath)
	}

	return nil
}

// detectProjectType identifies the project type from files present.
func detectProjectType(path string) string {
	// Check in priority order
	if _, err := os.Stat(filepath.Join(path, "wails.json")); err == nil {
		return "wails"
	}
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(path, "composer.json")); err == nil {
		return "php"
	}
	if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
		return "node"
	}
	return "unknown"
}

// generateBuildConfig creates a build.yaml configuration based on project type.
func generateBuildConfig(path, projectType string) string {
	name := filepath.Base(path)

	switch projectType {
	case "go", "wails":
		return fmt.Sprintf(`version: 1
project:
  name: %s
  description: Go application
  main: ./cmd/%s
  binary: %s
build:
  cgo: false
  flags:
    - -trimpath
  ldflags:
    - -s
    - -w
targets:
  - os: linux
    arch: amd64
  - os: linux
    arch: arm64
  - os: darwin
    arch: amd64
  - os: darwin
    arch: arm64
  - os: windows
    arch: amd64
`, name, name, name)

	case "php":
		return fmt.Sprintf(`version: 1
project:
  name: %s
  description: PHP application
  type: php
build:
  dockerfile: Dockerfile
  image: %s
`, name, name)

	case "node":
		return fmt.Sprintf(`version: 1
project:
  name: %s
  description: Node.js application
  type: node
build:
  script: npm run build
  output: dist
`, name)

	default:
		return fmt.Sprintf(`version: 1
project:
  name: %s
  description: Application
`, name)
	}
}

// generateReleaseConfig creates a release.yaml configuration.
func generateReleaseConfig(name, projectType string) string {
	// Try to detect GitHub repo from git remote
	repo := detectGitHubRepo()
	if repo == "" {
		repo = "owner/" + name
	}

	base := fmt.Sprintf(`version: 1
project:
  name: %s
  repository: %s
`, name, repo)

	switch projectType {
	case "go", "wails":
		return base + `
changelog:
  include:
    - feat
    - fix
    - perf
    - refactor
  exclude:
    - chore
    - docs
    - style
    - test

publishers:
  - type: github
    draft: false
    prerelease: false
`
	case "php":
		return base + `
changelog:
  include:
    - feat
    - fix
    - perf

publishers:
  - type: github
    draft: false
`
	default:
		return base + `
changelog:
  include:
    - feat
    - fix

publishers:
  - type: github
`
	}
}

// generateTestConfig creates a test.yaml configuration.
func generateTestConfig(projectType string) string {
	switch projectType {
	case "go", "wails":
		return `version: 1

commands:
  - name: unit
    run: go test ./...
  - name: coverage
    run: go test -coverprofile=coverage.out ./...
  - name: race
    run: go test -race ./...

env:
  CGO_ENABLED: "0"
`
	case "php":
		return `version: 1

commands:
  - name: unit
    run: vendor/bin/pest --parallel
  - name: types
    run: vendor/bin/phpstan analyse
  - name: lint
    run: vendor/bin/pint --test

env:
  APP_ENV: testing
  DB_CONNECTION: sqlite
`
	case "node":
		return `version: 1

commands:
  - name: unit
    run: npm test
  - name: lint
    run: npm run lint
  - name: typecheck
    run: npm run typecheck

env:
  NODE_ENV: test
`
	default:
		return `version: 1

commands:
  - name: test
    run: echo "No tests configured"
`
	}
}

// detectGitHubRepo tries to extract owner/repo from git remote.
func detectGitHubRepo() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	url := strings.TrimSpace(string(output))

	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		repo := strings.TrimPrefix(url, "git@github.com:")
		repo = strings.TrimSuffix(repo, ".git")
		return repo
	}

	// Handle HTTPS format: https://github.com/owner/repo.git
	if strings.Contains(url, "github.com/") {
		parts := strings.Split(url, "github.com/")
		if len(parts) == 2 {
			repo := strings.TrimSuffix(parts[1], ".git")
			return repo
		}
	}

	return ""
}

// isDirEmpty returns true if the directory is empty or contains only hidden files.
func isDirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
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
		if !strings.HasPrefix(name, ".") {
			return false, nil
		}
	}

	return true, nil
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
