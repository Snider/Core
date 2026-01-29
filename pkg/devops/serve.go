package devops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ServeOptions configures the dev server.
type ServeOptions struct {
	Port int    // Port to serve on (default 8000)
	Path string // Subdirectory to serve (default: current dir)
}

// Serve mounts the project and starts a dev server.
func (d *DevOps) Serve(ctx context.Context, projectDir string, opts ServeOptions) error {
	running, err := d.IsRunning(ctx)
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("dev environment not running (run 'core dev boot' first)")
	}

	if opts.Port == 0 {
		opts.Port = 8000
	}

	servePath := projectDir
	if opts.Path != "" {
		servePath = filepath.Join(projectDir, opts.Path)
	}

	// Mount project directory via SSHFS
	if err := d.mountProject(ctx, servePath); err != nil {
		return fmt.Errorf("failed to mount project: %w", err)
	}

	// Detect and run serve command
	serveCmd := DetectServeCommand(servePath)
	fmt.Printf("Starting server: %s\n", serveCmd)
	fmt.Printf("Listening on http://localhost:%d\n", opts.Port)

	// Run serve command via SSH
	return d.sshShell(ctx, []string{"cd", "/app", "&&", serveCmd})
}

// mountProject mounts a directory into the VM via SSHFS.
func (d *DevOps) mountProject(ctx context.Context, path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Use reverse SSHFS mount
	// The VM connects back to host to mount the directory
	cmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-R", "10000:localhost:22", // Reverse tunnel for SSHFS
		"-p", "2222",
		"root@localhost",
		fmt.Sprintf("mkdir -p /app && sshfs -p 10000 %s@localhost:%s /app -o allow_other", os.Getenv("USER"), absPath),
	)
	return cmd.Run()
}

// DetectServeCommand auto-detects the serve command for a project.
func DetectServeCommand(projectDir string) string {
	// Laravel/Octane
	if hasFile(projectDir, "artisan") {
		return "php artisan octane:start --host=0.0.0.0 --port=8000"
	}

	// Node.js with dev script
	if hasFile(projectDir, "package.json") {
		if hasPackageScript(projectDir, "dev") {
			return "npm run dev -- --host 0.0.0.0"
		}
		if hasPackageScript(projectDir, "start") {
			return "npm start"
		}
	}

	// PHP with composer
	if hasFile(projectDir, "composer.json") {
		return "frankenphp php-server -l :8000"
	}

	// Go
	if hasFile(projectDir, "go.mod") {
		if hasFile(projectDir, "main.go") {
			return "go run ."
		}
	}

	// Python Django
	if hasFile(projectDir, "manage.py") {
		return "python manage.py runserver 0.0.0.0:8000"
	}

	// Fallback: simple HTTP server
	return "python3 -m http.server 8000"
}
