package dev

import (
	"context"
	"sort"
	"strings"

	"github.com/host-uk/core/pkg/agentic"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/framework"
	"github.com/host-uk/core/pkg/git"
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
		cli.Println("No git repositories found")
		return nil
	}

	// QUERY git status
	result, handled, err := s.Core().QUERY(git.QueryStatus{
		Paths: paths,
		Names: names,
	})
	if !handled {
		return cli.Err("git service not available")
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
		cli.Blank()
		cli.Println("Committing changes...")
		cli.Blank()

		for _, repo := range dirtyRepos {
			_, handled, err := s.Core().PERFORM(agentic.TaskCommit{
				Path: repo.Path,
				Name: repo.Name,
			})
			if !handled {
				// Agentic service not available - skip silently
				cli.Print("  - %s: agentic service not available\n", repo.Name)
				continue
			}
			if err != nil {
				cli.Print("  x %s: %s\n", repo.Name, err)
			} else {
				cli.Print("  v %s\n", repo.Name)
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
			cli.Blank()
			cli.Println("Use --commit flag to auto-commit dirty repos")
		}
		return nil
	}

	// Push repos with unpushed commits
	if len(aheadRepos) == 0 {
		cli.Blank()
		cli.Println("All repositories are up to date")
		return nil
	}

	cli.Blank()
	cli.Print("%d repos with unpushed commits:\n", len(aheadRepos))
	for _, st := range aheadRepos {
		cli.Print("  %s: %d commits\n", st.Name, st.Ahead)
	}

	cli.Blank()
	cli.Print("Push all? [y/N] ")
	var answer string
	cli.Scanln(&answer)
	if strings.ToLower(answer) != "y" {
		cli.Println("Aborted")
		return nil
	}

	cli.Blank()

	// Push each repo
	for _, st := range aheadRepos {
		_, handled, err := s.Core().PERFORM(git.TaskPush{
			Path: st.Path,
			Name: st.Name,
		})
		if !handled {
			cli.Print("  x %s: git service not available\n", st.Name)
			continue
		}
		if err != nil {
			if git.IsNonFastForward(err) {
				cli.Print("  ! %s: branch has diverged\n", st.Name)
			} else {
				cli.Print("  x %s: %s\n", st.Name, err)
			}
		} else {
			cli.Print("  v %s\n", st.Name)
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
		cli.Println("No git repositories found")
		return nil
	}

	result, handled, err := s.Core().QUERY(git.QueryStatus{
		Paths: paths,
		Names: names,
	})
	if !handled {
		return cli.Err("git service not available")
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
	reg, _, err := loadRegistryWithConfig(registryPath)
	if err != nil {
		return nil, nil, err
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
	cli.Print("%-*s  %8s  %9s  %6s  %5s\n",
		nameWidth, "Repo", "Modified", "Untracked", "Staged", "Ahead")

	// Print separator
	cli.Text(strings.Repeat("-", nameWidth+2+10+11+8+7))

	// Print rows
	for _, st := range statuses {
		if st.Error != nil {
			cli.Print("%-*s  error: %s\n", nameWidth, st.Name, st.Error)
			continue
		}

		cli.Print("%-*s  %8d  %9d  %6d  %5d\n",
			nameWidth, st.Name,
			st.Modified, st.Untracked, st.Staged, st.Ahead)
	}
}
