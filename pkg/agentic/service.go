package agentic

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/host-uk/core/pkg/framework"
)

// Tasks for AI service

// TaskCommit requests Claude to create a commit.
type TaskCommit struct {
	Path    string
	Name    string
	CanEdit bool // allow Write/Edit tools
}

// TaskPrompt sends a custom prompt to Claude.
type TaskPrompt struct {
	Prompt       string
	WorkDir      string
	AllowedTools []string
}

// ServiceOptions for configuring the AI service.
type ServiceOptions struct {
	DefaultTools []string
	AllowEdit    bool // global permission for Write/Edit tools
}

// DefaultServiceOptions returns sensible defaults.
func DefaultServiceOptions() ServiceOptions {
	return ServiceOptions{
		DefaultTools: []string{"Bash", "Read", "Glob", "Grep"},
		AllowEdit:    false,
	}
}

// Service provides AI/Claude operations as a Core service.
type Service struct {
	*framework.ServiceRuntime[ServiceOptions]
}

// NewService creates an AI service factory.
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
	case TaskCommit:
		err := s.doCommit(m)
		return nil, true, err

	case TaskPrompt:
		err := s.doPrompt(m)
		return nil, true, err
	}
	return nil, false, nil
}

func (s *Service) doCommit(task TaskCommit) error {
	prompt := Prompt("commit")

	tools := []string{"Bash", "Read", "Glob", "Grep"}
	if task.CanEdit {
		tools = []string{"Bash", "Read", "Write", "Edit", "Glob", "Grep"}
	}

	cmd := exec.CommandContext(context.Background(), "claude", "-p", prompt, "--allowedTools", strings.Join(tools, ","))
	cmd.Dir = task.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (s *Service) doPrompt(task TaskPrompt) error {
	opts := s.Opts()
	tools := opts.DefaultTools
	if len(tools) == 0 {
		tools = []string{"Bash", "Read", "Glob", "Grep"}
	}

	if len(task.AllowedTools) > 0 {
		tools = task.AllowedTools
	}

	cmd := exec.CommandContext(context.Background(), "claude", "-p", task.Prompt, "--allowedTools", strings.Join(tools, ","))
	if task.WorkDir != "" {
		cmd.Dir = task.WorkDir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
