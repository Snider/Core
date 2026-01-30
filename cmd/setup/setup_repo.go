// setup_repo.go implements repository setup with .core/ configuration.
//
// When running setup in an existing git repository, this generates
// build.yaml, release.yaml, and test.yaml configurations based on
// detected project type.

package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
