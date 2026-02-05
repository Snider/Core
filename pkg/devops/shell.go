package devops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// ShellOptions configures the shell connection.
type ShellOptions struct {
	Console bool     // Use serial console instead of SSH
	Command []string // Command to run (empty = interactive shell)
}

// Shell connects to the dev environment.
func (d *DevOps) Shell(ctx context.Context, opts ShellOptions) error {
	running, err := d.IsRunning(ctx)
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("dev environment not running (run 'core dev boot' first)")
	}

	if opts.Console {
		return d.serialConsole(ctx)
	}

	return d.sshShell(ctx, opts.Command)
}

// sshShell connects via SSH.
func (d *DevOps) sshShell(ctx context.Context, command []string) error {
	args := []string{
		"-o", "StrictHostKeyChecking=yes",
		"-o", "UserKnownHostsFile=~/.core/known_hosts",
		"-o", "LogLevel=ERROR",
		"-A", // Agent forwarding
		"-p", fmt.Sprintf("%d", DefaultSSHPort),
		"root@localhost",
	}

	if len(command) > 0 {
		args = append(args, command...)
	}

	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// serialConsole attaches to the QEMU serial console.
func (d *DevOps) serialConsole(ctx context.Context) error {
	// Find the container to get its console socket
	c, err := d.findContainer(ctx, "core-dev")
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("console not available: container not found")
	}

	// Use socat to connect to the console socket
	socketPath := fmt.Sprintf("/tmp/core-%s-console.sock", c.ID)
	cmd := exec.CommandContext(ctx, "socat", "-,raw,echo=0", "unix-connect:"+socketPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
