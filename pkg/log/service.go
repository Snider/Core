package log

import (
	"context"

	"github.com/host-uk/core/pkg/framework"
)

// Service wraps Logger for Core framework integration.
type Service struct {
	*framework.ServiceRuntime[Options]
	*Logger
}

// NewService creates a log service factory for Core.
func NewService(opts Options) func(*framework.Core) (any, error) {
	return func(c *framework.Core) (any, error) {
		logger := New(opts)

		return &Service{
			ServiceRuntime: framework.NewServiceRuntime(c, opts),
			Logger:         logger,
		}, nil
	}
}

// OnStartup registers query and task handlers.
func (s *Service) OnStartup(ctx context.Context) error {
	s.Core().RegisterQuery(s.handleQuery)
	s.Core().RegisterTask(s.handleTask)
	return nil
}

// QueryLevel returns the current log level.
type QueryLevel struct{}

// TaskSetLevel changes the log level.
type TaskSetLevel struct {
	Level Level
}

func (s *Service) handleQuery(c *framework.Core, q framework.Query) (any, bool, error) {
	switch q.(type) {
	case QueryLevel:
		return s.Level(), true, nil
	}
	return nil, false, nil
}

func (s *Service) handleTask(c *framework.Core, t framework.Task) (any, bool, error) {
	switch m := t.(type) {
	case TaskSetLevel:
		s.SetLevel(m.Level)
		return nil, true, nil
	}
	return nil, false, nil
}
