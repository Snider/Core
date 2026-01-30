package dev

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/host-uk/core/pkg/agentic"
	"github.com/host-uk/core/pkg/framework"
	"github.com/host-uk/core/pkg/git"
	"github.com/host-uk/core/pkg/repos"
)

// Tasks for dev service

// TaskWork runs the full dev workflow: status, commit, push.
type TaskWork struct {
	RegistryPath string
	StatusOnly   bool
	AutoCommit   bool
}

// TaskStatus displays git status for all repos.
type TaskStatus struct {
	RegistryPath string
}

// ServiceOptions for configuring the dev service.
type ServiceOptions struct {
	RegistryPath string
}

// Service provides dev workflow orchestration as a Core service.
type Service struct {
	*framework.ServiceRuntime[ServiceOptions]
}

// NewService creates a dev service factory.
func NewService(opts ServiceOptions) func(*framework.Core) (any, error) {
	return func(c *framework.Core) (any, error) {
		return &Service{
			ServiceRuntime: framework.NewServiceRuntime(c, opts),
		}, nil
	}
}

// OnStartup registers task handlers.
func (s *Service) OnStartup(ctx context.Context) error {
	s.Core().RegisterTask(s.handleTask)
	return nil
}

func (s *Service) handleTask(c *framework.Core, t framework.Task) (any, bool, error) {
	switch m := t.(type) {
	case TaskWork:
		err := s.runWork(m)
		return nil, true, err

	case TaskStatus:
		err := s.runStatus(m)
		return nil, true, err
	}
	return nil, false, nil
}

func (s *Service) runWork(task TaskWork) error {
	// Load registry
	paths, names, err := s.loadRegistry(task.RegistryPath)
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		fmt.Println("No git repositories found")
		return nil
	}

	// QUERY git status
	result, handled, err := s.Core().QUERY(git.QueryStatus{
		Paths: paths,
		Names: names,
	})
	if !handled {
		return fmt.Errorf("git service not available")
	}
	if err != nil {
		return err
	}
	statuses := result.([]git.RepoStatus)

	// Sort by name
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	// Display status table
	s.printStatusTable(statuses)

	// Collect dirty and ahead repos
	var dirtyRepos []git.RepoStatus
	var aheadRepos []git.RepoStatus

	for _, st := range statuses {
		if st.Error != nil {
			continue
		}
		if st.IsDirty() {
			dirtyRepos = append(dirtyRepos, st)
		}
		if st.HasUnpushed() {
			aheadRepos = append(aheadRepos, st)
		}
	}

	// Auto-commit dirty repos if requested
	if task.AutoCommit && len(dirtyRepos) > 0 {
		fmt.Println()
		fmt.Println("Committing changes...")
		fmt.Println()

		for _, repo := range dirtyRepos {
			_, handled, err := s.Core().PERFORM(agentic.TaskCommit{
				Path: repo.Path,
				Name: repo.Name,
			})
			if !handled {
				// Agentic service not available - skip silently
				fmt.Printf("  - %s: agentic service not available\n", repo.Name)
				continue
			}
			if err != nil {
				fmt.Printf("  x %s: %s\n", repo.Name, err)
			} else {
				fmt.Printf("  v %s\n", repo.Name)
			}
		}

		// Re-query status after commits
		result, _, _ = s.Core().QUERY(git.QueryStatus{
			Paths: paths,
			Names: names,
		})
		statuses = result.([]git.RepoStatus)

		// Rebuild ahead repos list
		aheadRepos = nil
		for _, st := range statuses {
			if st.Error == nil && st.HasUnpushed() {
				aheadRepos = append(aheadRepos, st)
			}
		}
	}

	// If status only, we're done
	if task.StatusOnly {
		if len(dirtyRepos) > 0 && !task.AutoCommit {
			fmt.Println()
			fmt.Println("Use --commit flag to auto-commit dirty repos")
		}
		return nil
	}

	// Push repos with unpushed commits
	if len(aheadRepos) == 0 {
		fmt.Println()
		fmt.Println("All repositories are up to date")
		return nil
	}

	fmt.Println()
	fmt.Printf("%d repos with unpushed commits:\n", len(aheadRepos))
	for _, st := range aheadRepos {
		fmt.Printf("  %s: %d commits\n", st.Name, st.Ahead)
	}

	fmt.Println()
	fmt.Print("Push all? [y/N] ")
	var answer string
	fmt.Scanln(&answer)
	if strings.ToLower(answer) != "y" {
		fmt.Println("Aborted")
		return nil
	}

	fmt.Println()

	// Push each repo
	for _, st := range aheadRepos {
		_, handled, err := s.Core().PERFORM(git.TaskPush{
			Path: st.Path,
			Name: st.Name,
		})
		if !handled {
			fmt.Printf("  x %s: git service not available\n", st.Name)
			continue
		}
		if err != nil {
			if git.IsNonFastForward(err) {
				fmt.Printf("  ! %s: branch has diverged\n", st.Name)
			} else {
				fmt.Printf("  x %s: %s\n", st.Name, err)
			}
		} else {
			fmt.Printf("  v %s\n", st.Name)
		}
	}

	return nil
}

func (s *Service) runStatus(task TaskStatus) error {
	paths, names, err := s.loadRegistry(task.RegistryPath)
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		fmt.Println("No git repositories found")
		return nil
	}

	result, handled, err := s.Core().QUERY(git.QueryStatus{
		Paths: paths,
		Names: names,
	})
	if !handled {
		return fmt.Errorf("git service not available")
	}
	if err != nil {
		return err
	}

	statuses := result.([]git.RepoStatus)
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	s.printStatusTable(statuses)
	return nil
}

func (s *Service) loadRegistry(registryPath string) ([]string, map[string]string, error) {
	var reg *repos.Registry
	var err error

	if registryPath != "" {
		reg, err = repos.LoadRegistry(registryPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load registry: %w", err)
		}
		fmt.Printf("Registry: %s\n\n", registryPath)
	} else {
		registryPath, err = repos.FindRegistry()
		if err == nil {
			reg, err = repos.LoadRegistry(registryPath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load registry: %w", err)
			}
			fmt.Printf("Registry: %s\n\n", registryPath)
		} else {
			// Fallback: scan current directory
			cwd, _ := os.Getwd()
			reg, err = repos.ScanDirectory(cwd)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to scan directory: %w", err)
			}
			fmt.Printf("Scanning: %s\n\n", cwd)
		}
	}

	var paths []string
	names := make(map[string]string)

	for _, repo := range reg.List() {
		if repo.IsGitRepo() {
			paths = append(paths, repo.Path)
			names[repo.Path] = repo.Name
		}
	}

	return paths, names, nil
}

func (s *Service) printStatusTable(statuses []git.RepoStatus) {
	// Calculate column widths
	nameWidth := 4 // "Repo"
	for _, st := range statuses {
		if len(st.Name) > nameWidth {
			nameWidth = len(st.Name)
		}
	}

	// Print header
	fmt.Printf("%-*s  %8s  %9s  %6s  %5s\n",
		nameWidth, "Repo", "Modified", "Untracked", "Staged", "Ahead")

	// Print separator
	fmt.Println(strings.Repeat("-", nameWidth+2+10+11+8+7))

	// Print rows
	for _, st := range statuses {
		if st.Error != nil {
			fmt.Printf("%-*s  error: %s\n", nameWidth, st.Name, st.Error)
			continue
		}

		fmt.Printf("%-*s  %8d  %9d  %6d  %5d\n",
			nameWidth, st.Name,
			st.Modified, st.Untracked, st.Staged, st.Ahead)
	}
}
