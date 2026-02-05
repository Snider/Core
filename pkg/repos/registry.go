// Package repos provides functionality for managing multi-repo workspaces.
// It reads a repos.yaml registry file that defines repositories, their types,
// dependencies, and metadata.
package repos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/io"
	"gopkg.in/yaml.v3"
)

// Registry represents a collection of repositories defined in repos.yaml.
type Registry struct {
	Version  int              `yaml:"version"`
	Org      string           `yaml:"org"`
	BasePath string           `yaml:"base_path"`
	Repos    map[string]*Repo `yaml:"repos"`
	Defaults RegistryDefaults `yaml:"defaults"`
}

// RegistryDefaults contains default values applied to all repos.
type RegistryDefaults struct {
	CI      string `yaml:"ci"`
	License string `yaml:"license"`
	Branch  string `yaml:"branch"`
}

// RepoType indicates the role of a repository in the ecosystem.
type RepoType string

// Repository type constants for ecosystem classification.
const (
	// RepoTypeFoundation indicates core foundation packages.
	RepoTypeFoundation RepoType = "foundation"
	// RepoTypeModule indicates reusable module packages.
	RepoTypeModule RepoType = "module"
	// RepoTypeProduct indicates end-user product applications.
	RepoTypeProduct RepoType = "product"
	// RepoTypeTemplate indicates starter templates.
	RepoTypeTemplate RepoType = "template"
)

// Repo represents a single repository in the registry.
type Repo struct {
	Name        string   `yaml:"-"` // Set from map key
	Type        string   `yaml:"type"`
	DependsOn   []string `yaml:"depends_on"`
	Description string   `yaml:"description"`
	Docs        bool     `yaml:"docs"`
	CI          string   `yaml:"ci"`
	Domain      string   `yaml:"domain,omitempty"`
	Clone       *bool    `yaml:"clone,omitempty"` // nil = true, false = skip cloning

	// Computed fields
	Path string `yaml:"-"` // Full path to repo directory
}

// LoadRegistry reads and parses a repos.yaml file.
func LoadRegistry(path string) (*Registry, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	content, err := io.Local.Read(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry file: %w", err)
	}
	data := []byte(content)

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("failed to parse registry file: %w", err)
	}

	// Expand base path
	reg.BasePath = expandPath(reg.BasePath)

	// Set computed fields on each repo
	for name, repo := range reg.Repos {
		repo.Name = name
		repo.Path = filepath.Join(reg.BasePath, name)

		// Apply defaults if not set
		if repo.CI == "" {
			repo.CI = reg.Defaults.CI
		}
	}

	return &reg, nil
}

// FindRegistry searches for repos.yaml in common locations.
// It checks: current directory, parent directories, and home directory.
func FindRegistry() (string, error) {
	// Check current directory and parents
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, "repos.yaml")
		if io.Local.Exists(candidate) {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Check home directory common locations
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	commonPaths := []string{
		filepath.Join(home, "Code", "host-uk", "repos.yaml"),
		filepath.Join(home, ".config", "core", "repos.yaml"),
	}

	for _, p := range commonPaths {
		if io.Local.Exists(p) {
			return p, nil
		}
	}

	return "", fmt.Errorf("repos.yaml not found")
}

// ScanDirectory creates a Registry by scanning a directory for git repos.
// This is used as a fallback when no repos.yaml is found.
func ScanDirectory(dir string) (*Registry, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory path: %w", err)
	}

	entries, err := io.Local.List(absDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	reg := &Registry{
		Version:  1,
		BasePath: absDir,
		Repos:    make(map[string]*Repo),
	}

	// Try to detect org from git remote
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(absDir, entry.Name())
		gitPath := filepath.Join(repoPath, ".git")

		if !io.Local.IsDir(gitPath) {
			continue // Not a git repo
		}

		repo := &Repo{
			Name: entry.Name(),
			Path: repoPath,
			Type: "module", // Default type
		}

		reg.Repos[entry.Name()] = repo

		// Try to detect org from first repo's remote
		if reg.Org == "" {
			reg.Org = detectOrg(repoPath)
		}
	}

	return reg, nil
}

// detectOrg tries to extract the GitHub org from a repo's origin remote.
func detectOrg(repoPath string) string {
	// Try to read git remote
	configPath := filepath.Join(repoPath, ".git", "config")
	content, err := io.Local.Read(configPath)
	if err != nil {
		return ""
	}
	// Look for patterns like github.com:org/repo or github.com/org/repo
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "url = ") {
			continue
		}
		url := strings.TrimPrefix(line, "url = ")

		// git@github.com:org/repo.git
		if strings.Contains(url, "github.com:") {
			parts := strings.Split(url, ":")
			if len(parts) >= 2 {
				orgRepo := strings.TrimSuffix(parts[1], ".git")
				orgParts := strings.Split(orgRepo, "/")
				if len(orgParts) >= 1 {
					return orgParts[0]
				}
			}
		}

		// https://github.com/org/repo.git
		if strings.Contains(url, "github.com/") {
			parts := strings.Split(url, "github.com/")
			if len(parts) >= 2 {
				orgRepo := strings.TrimSuffix(parts[1], ".git")
				orgParts := strings.Split(orgRepo, "/")
				if len(orgParts) >= 1 {
					return orgParts[0]
				}
			}
		}
	}

	return ""
}

// List returns all repos in the registry.
func (r *Registry) List() []*Repo {
	repos := make([]*Repo, 0, len(r.Repos))
	for _, repo := range r.Repos {

		repos = append(repos, repo)
	}
	return repos
}

// Get returns a repo by name.
func (r *Registry) Get(name string) (*Repo, bool) {
	repo, ok := r.Repos[name]
	return repo, ok
}

// ByType returns repos filtered by type.
func (r *Registry) ByType(t string) []*Repo {
	var repos []*Repo
	for _, repo := range r.Repos {
		if repo.Type == t {
			repos = append(repos, repo)
		}
	}
	return repos
}

// TopologicalOrder returns repos sorted by dependency order.
// Foundation repos come first, then modules, then products.
func (r *Registry) TopologicalOrder() ([]*Repo, error) {
	// Build dependency graph
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var result []*Repo

	var visit func(name string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		if visiting[name] {
			return fmt.Errorf("circular dependency detected: %s", name)
		}

		repo, ok := r.Repos[name]
		if !ok {
			return fmt.Errorf("unknown repo: %s", name)
		}

		visiting[name] = true
		for _, dep := range repo.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		result = append(result, repo)
		return nil
	}

	for name := range r.Repos {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Exists checks if the repo directory exists on disk.
func (repo *Repo) Exists() bool {
	return io.Local.IsDir(repo.Path)
}

// IsGitRepo checks if the repo directory contains a .git folder.
func (repo *Repo) IsGitRepo() bool {
	gitPath := filepath.Join(repo.Path, ".git")
	return io.Local.IsDir(gitPath)
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
