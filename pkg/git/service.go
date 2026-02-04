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

func (QueryStatus) Response() []RepoStatus { return nil }

// QueryDirtyRepos requests repos with uncommitted changes.
type QueryDirtyRepos struct{}

func (QueryDirtyRepos) Response() []RepoStatus { return nil }

// QueryAheadRepos requests repos with unpushed commits.
type QueryAheadRepos struct{}

func (QueryAheadRepos) Response() []RepoStatus { return nil }

// Tasks for git service

// TaskPush requests git push for a path.
type TaskPush struct {
	Path string
	Name string
}

func (TaskPush) Response() any { return nil }

// TaskPull requests git pull for a path.
type TaskPull struct {
	Path string
	Name string
}

func (TaskPull) Response() any { return nil }

// TaskPushMultiple requests git push for multiple paths.
type TaskPushMultiple struct {
	Paths []string
	Names map[string]string
}

func (TaskPushMultiple) Response() []PushResult { return nil }

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
	framework.RegisterQuery(s.Core(), s.handleQueryStatus)
	framework.RegisterQuery(s.Core(), s.handleQueryDirtyRepos)
	framework.RegisterQuery(s.Core(), s.handleQueryAheadRepos)
	framework.RegisterTask(s.Core(), s.handleTaskPush)
	framework.RegisterTask(s.Core(), s.handleTaskPull)
	framework.RegisterTask(s.Core(), s.handleTaskPushMultiple)
	return nil
}

func (s *Service) handleQueryStatus(c *framework.Core, q QueryStatus) ([]RepoStatus, bool, error) {
	statuses := Status(context.Background(), StatusOptions(q))
	s.lastStatus = statuses
	return statuses, true, nil
}

func (s *Service) handleQueryDirtyRepos(c *framework.Core, q QueryDirtyRepos) ([]RepoStatus, bool, error) {
	return s.DirtyRepos(), true, nil
}

func (s *Service) handleQueryAheadRepos(c *framework.Core, q QueryAheadRepos) ([]RepoStatus, bool, error) {
	return s.AheadRepos(), true, nil
}

func (s *Service) handleTaskPush(c *framework.Core, t TaskPush) (any, bool, error) {
	err := Push(context.Background(), t.Path)
	return nil, true, err
}

func (s *Service) handleTaskPull(c *framework.Core, t TaskPull) (any, bool, error) {
	err := Pull(context.Background(), t.Path)
	return nil, true, err
}

func (s *Service) handleTaskPushMultiple(c *framework.Core, t TaskPushMultiple) ([]PushResult, bool, error) {
	results := PushMultiple(context.Background(), t.Paths, t.Names)
	return results, true, nil
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
