package git

import (
	"context"

	"github.com/host-uk/core/pkg/framework"
)

// Queries for git service

// QueryStatus requests git status for paths.
type QueryStatus struct {
	Paths []string
	Names map[string]string
}

// QueryDirtyRepos requests repos with uncommitted changes.
type QueryDirtyRepos struct{}

// QueryAheadRepos requests repos with unpushed commits.
type QueryAheadRepos struct{}

// Tasks for git service

// TaskPush requests git push for a path.
type TaskPush struct {
	Path string
	Name string
}

// TaskPull requests git pull for a path.
type TaskPull struct {
	Path string
	Name string
}

// TaskPushMultiple requests git push for multiple paths.
type TaskPushMultiple struct {
	Paths []string
	Names map[string]string
}

// ServiceOptions for configuring the git service.
type ServiceOptions struct {
	WorkDir string
}

// Service provides git operations as a Core service.
type Service struct {
	*framework.ServiceRuntime[ServiceOptions]
	lastStatus []RepoStatus
}

// NewService creates a git service factory.
func NewService(opts ServiceOptions) func(*framework.Core) (any, error) {
	return func(c *framework.Core) (any, error) {
		return &Service{
			ServiceRuntime: framework.NewServiceRuntime(c, opts),
		}, nil
	}
}

// OnStartup registers query and task handlers.
func (s *Service) OnStartup(ctx context.Context) error {
	s.Core().RegisterQuery(s.handleQuery)
	s.Core().RegisterTask(s.handleTask)
	return nil
}

func (s *Service) handleQuery(c *framework.Core, q framework.Query) (any, bool, error) {
	switch m := q.(type) {
	case QueryStatus:
		statuses := Status(context.Background(), StatusOptions(m))
		s.lastStatus = statuses
		return statuses, true, nil

	case QueryDirtyRepos:
		return s.DirtyRepos(), true, nil

	case QueryAheadRepos:
		return s.AheadRepos(), true, nil
	}
	return nil, false, nil
}

func (s *Service) handleTask(c *framework.Core, t framework.Task) (any, bool, error) {
	switch m := t.(type) {
	case TaskPush:
		err := Push(context.Background(), m.Path)
		return nil, true, err

	case TaskPull:
		err := Pull(context.Background(), m.Path)
		return nil, true, err

	case TaskPushMultiple:
		results := PushMultiple(context.Background(), m.Paths, m.Names)
		return results, true, nil
	}
	return nil, false, nil
}

// Status returns last status result.
func (s *Service) Status() []RepoStatus { return s.lastStatus }

// DirtyRepos returns repos with uncommitted changes.
func (s *Service) DirtyRepos() []RepoStatus {
	var dirty []RepoStatus
	for _, st := range s.lastStatus {
		if st.Error == nil && st.IsDirty() {
			dirty = append(dirty, st)
		}
	}
	return dirty
}

// AheadRepos returns repos with unpushed commits.
func (s *Service) AheadRepos() []RepoStatus {
	var ahead []RepoStatus
	for _, st := range s.lastStatus {
		if st.Error == nil && st.HasUnpushed() {
			ahead = append(ahead, st)
		}
	}
	return ahead
}
