package agentic

import (
	"context"
	"os"
	"os/exec"

	"github.com/host-uk/core/pkg/framework"
)

// Actions for AI service IPC

// ActionCommit requests Claude to create a commit.
type ActionCommit struct {
	Path    string
	Name    string
	CanEdit bool // allow Write/Edit tools
}

// ActionPrompt sends a custom prompt to Claude.
type ActionPrompt struct {
	Prompt       string
	WorkDir      string
	AllowedTools []string
}

// ServiceOptions for configuring the AI service.
type ServiceOptions struct {
	DefaultTools []string
}

// DefaultServiceOptions returns sensible defaults.
func DefaultServiceOptions() ServiceOptions {
	return ServiceOptions{
		DefaultTools: []string{"Bash", "Read", "Glob", "Grep"},
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

// OnStartup registers action handlers.
func (s *Service) OnStartup(ctx context.Context) error {
	s.Core().RegisterAction(s.handle)
	return nil
}

func (s *Service) handle(c *framework.Core, msg framework.Message) error {
	switch m := msg.(type) {
	case ActionCommit:
		return s.handleCommit(m)
	case ActionPrompt:
		return s.handlePrompt(m)
	}
	return nil
}

func (s *Service) handleCommit(action ActionCommit) error {
	prompt := Prompt("commit")

	tools := "Bash,Read,Glob,Grep"
	if action.CanEdit {
		tools = "Bash,Read,Write,Edit,Glob,Grep"
	}

	cmd := exec.CommandContext(context.Background(), "claude", "-p", prompt, "--allowedTools", tools)
	cmd.Dir = action.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (s *Service) handlePrompt(action ActionPrompt) error {
	tools := "Bash,Read,Glob,Grep"
	if len(action.AllowedTools) > 0 {
		tools = ""
		for i, t := range action.AllowedTools {
			if i > 0 {
				tools += ","
			}
			tools += t
		}
	}

	cmd := exec.CommandContext(context.Background(), "claude", "-p", action.Prompt, "--allowedTools", tools)
	if action.WorkDir != "" {
		cmd.Dir = action.WorkDir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
