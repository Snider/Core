package devops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ClaudeOptions configures the Claude sandbox session.
type ClaudeOptions struct {
	NoAuth bool     // Don't forward any auth
	Auth   []string // Selective auth: "gh", "anthropic", "ssh", "git"
	Model  string   // Model to use: opus, sonnet
}

// Claude starts a sandboxed Claude session in the dev environment.
func (d *DevOps) Claude(ctx context.Context, projectDir string, opts ClaudeOptions) error {
	// Auto-boot if not running
	running, err := d.IsRunning(ctx)
	if err != nil {
		return err
	}
	if !running {
		fmt.Println("Dev environment not running, booting...")
		if err := d.Boot(ctx, DefaultBootOptions()); err != nil {
			return fmt.Errorf("failed to boot: %w", err)
		}
	}

	// Mount project
	if err := d.mountProject(ctx, projectDir); err != nil {
		return fmt.Errorf("failed to mount project: %w", err)
	}

	// Prepare environment variables to forward
	envVars := []string{}

	if !opts.NoAuth {
		authTypes := opts.Auth
		if len(authTypes) == 0 {
			authTypes = []string{"gh", "anthropic", "ssh", "git"}
		}

		for _, auth := range authTypes {
			switch auth {
			case "anthropic":
				if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
					envVars = append(envVars, "ANTHROPIC_API_KEY="+key)
				}
			case "git":
				// Forward git config
				name, _ := exec.Command("git", "config", "user.name").Output()
				email, _ := exec.Command("git", "config", "user.email").Output()
				if len(name) > 0 {
					envVars = append(envVars, "GIT_AUTHOR_NAME="+strings.TrimSpace(string(name)))
					envVars = append(envVars, "GIT_COMMITTER_NAME="+strings.TrimSpace(string(name)))
				}
				if len(email) > 0 {
					envVars = append(envVars, "GIT_AUTHOR_EMAIL="+strings.TrimSpace(string(email)))
					envVars = append(envVars, "GIT_COMMITTER_EMAIL="+strings.TrimSpace(string(email)))
				}
			}
		}
	}

	// Build SSH command with agent forwarding
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-A", // SSH agent forwarding
		"-p", "2222",
	}

	args = append(args, "root@localhost")

	// Build command to run inside
	claudeCmd := "cd /app && claude"
	if opts.Model != "" {
		claudeCmd += " --model " + opts.Model
	}
	args = append(args, claudeCmd)

	// Set environment for SSH
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Pass environment variables through SSH
	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			cmd.Env = append(os.Environ(), env)
		}
	}

	fmt.Println("Starting Claude in sandboxed environment...")
	fmt.Println("Project mounted at /app")
	fmt.Println("Auth forwarded: SSH agent" + formatAuthList(opts))
	fmt.Println()

	return cmd.Run()
}

func formatAuthList(opts ClaudeOptions) string {
	if opts.NoAuth {
		return " (none)"
	}
	if len(opts.Auth) == 0 {
		return ", gh, anthropic, git"
	}
	return ", " + strings.Join(opts.Auth, ", ")
}

// CopyGHAuth copies GitHub CLI auth to the VM.
func (d *DevOps) CopyGHAuth(ctx context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	ghConfigDir := filepath.Join(home, ".config", "gh")
	if _, err := os.Stat(ghConfigDir); os.IsNotExist(err) {
		return nil // No gh config to copy
	}

	// Use scp to copy gh config
	cmd := exec.CommandContext(ctx, "scp",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-P", "2222",
		"-r", ghConfigDir,
		"root@localhost:/root/.config/",
	)
	return cmd.Run()
}
