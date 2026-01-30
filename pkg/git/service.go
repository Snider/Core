package git

import (
	"context"

	"github.com/host-uk/core/pkg/framework"
)

// Actions for git service IPC

// ActionStatus requests git status for paths.
type ActionStatus struct {
	Paths []string
	Names map[string]string
}

// ActionPush requests git push for a path.
type ActionPush struct{ Path, Name string }

// ActionPull requests git pull for a path.
type ActionPull struct{ Path, Name string }

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

// OnStartup registers action handlers.
func (s *Service) OnStartup(ctx context.Context) error {
	s.Core().RegisterAction(s.handle)
	return nil
}

func (s *Service) handle(c *framework.Core, msg framework.Message) error {
	switch m := msg.(type) {
	case ActionStatus:
		return s.handleStatus(m)
	case ActionPush:
		return s.handlePush(m)
	case ActionPull:
		return s.handlePull(m)
	}
	return nil
}

func (s *Service) handleStatus(action ActionStatus) error {
	ctx := context.Background()
	statuses := Status(ctx, StatusOptions{
		Paths: action.Paths,
		Names: action.Names,
	})
	s.lastStatus = statuses
	return nil
}

func (s *Service) handlePush(action ActionPush) error {
	return Push(context.Background(), action.Path)
}

func (s *Service) handlePull(action ActionPull) error {
	return Pull(context.Background(), action.Path)
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
