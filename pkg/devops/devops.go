// Package devops provides a portable development environment using LinuxKit images.
package devops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/host-uk/core/pkg/container"
	"github.com/host-uk/core/pkg/io"
)

// DevOps manages the portable development environment.
type DevOps struct {
	medium    io.Medium
	config    *Config
	images    *ImageManager
	container *container.LinuxKitManager
}

// New creates a new DevOps instance using the provided medium.
func New(m io.Medium) (*DevOps, error) {
	cfg, err := LoadConfig(m)
	if err != nil {
		return nil, fmt.Errorf("devops.New: failed to load config: %w", err)
	}

	images, err := NewImageManager(m, cfg)
	if err != nil {
		return nil, fmt.Errorf("devops.New: failed to create image manager: %w", err)
	}

	mgr, err := container.NewLinuxKitManager()
	if err != nil {
		return nil, fmt.Errorf("devops.New: failed to create container manager: %w", err)
	}

	return &DevOps{
		medium:    m,
		config:    cfg,
		images:    images,
		container: mgr,
	}, nil
}

// ImageName returns the platform-specific image name.
func ImageName() string {
	return fmt.Sprintf("core-devops-%s-%s.qcow2", runtime.GOOS, runtime.GOARCH)
}

// ImagesDir returns the path to the images directory.
func ImagesDir() (string, error) {
	if dir := os.Getenv("CORE_IMAGES_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".core", "images"), nil
}

// ImagePath returns the full path to the platform-specific image.
func ImagePath() (string, error) {
	dir, err := ImagesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ImageName()), nil
}

// IsInstalled checks if the dev image is installed.
func (d *DevOps) IsInstalled() bool {
	path, err := ImagePath()
	if err != nil {
		return false
	}
	return d.medium.IsFile(path)
}

// Install downloads and installs the dev image.
func (d *DevOps) Install(ctx context.Context, progress func(downloaded, total int64)) error {
	return d.images.Install(ctx, progress)
}

// CheckUpdate checks if an update is available.
func (d *DevOps) CheckUpdate(ctx context.Context) (current, latest string, hasUpdate bool, err error) {
	return d.images.CheckUpdate(ctx)
}

// BootOptions configures how to boot the dev environment.
type BootOptions struct {
	Memory int    // MB, default 4096
	CPUs   int    // default 2
	Name   string // container name
	Fresh  bool   // destroy existing and start fresh
}

// DefaultBootOptions returns sensible defaults.
func DefaultBootOptions() BootOptions {
	return BootOptions{
		Memory: 4096,
		CPUs:   2,
		Name:   "core-dev",
	}
}

// Boot starts the dev environment.
func (d *DevOps) Boot(ctx context.Context, opts BootOptions) error {
	if !d.images.IsInstalled() {
		return fmt.Errorf("dev image not installed (run 'core dev install' first)")
	}

	// Check if already running
	if !opts.Fresh {
		running, err := d.IsRunning(ctx)
		if err == nil && running {
			return fmt.Errorf("dev environment already running (use 'core dev stop' first or --fresh)")
		}
	}

	// Stop existing if fresh
	if opts.Fresh {
		_ = d.Stop(ctx)
	}

	imagePath, err := ImagePath()
	if err != nil {
		return err
	}

	// Build run options for LinuxKitManager
	runOpts := container.RunOptions{
		Name:    opts.Name,
		Memory:  opts.Memory,
		CPUs:    opts.CPUs,
		SSHPort: 2222,
		Detach:  true,
	}

	_, err = d.container.Run(ctx, imagePath, runOpts)
	return err
}

// Stop stops the dev environment.
func (d *DevOps) Stop(ctx context.Context) error {
	c, err := d.findContainer(ctx, "core-dev")
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("dev environment not found")
	}
	return d.container.Stop(ctx, c.ID)
}

// IsRunning checks if the dev environment is running.
func (d *DevOps) IsRunning(ctx context.Context) (bool, error) {
	c, err := d.findContainer(ctx, "core-dev")
	if err != nil {
		return false, err
	}
	return c != nil && c.Status == container.StatusRunning, nil
}

// findContainer finds a container by name.
func (d *DevOps) findContainer(ctx context.Context, name string) (*container.Container, error) {
	containers, err := d.container.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, nil
}

// DevStatus returns information about the dev environment.
type DevStatus struct {
	Installed    bool
	Running      bool
	ImageVersion string
	ContainerID  string
	Memory       int
	CPUs         int
	SSHPort      int
	Uptime       time.Duration
}

// Status returns the current dev environment status.
func (d *DevOps) Status(ctx context.Context) (*DevStatus, error) {
	status := &DevStatus{
		Installed: d.images.IsInstalled(),
		SSHPort:   2222,
	}

	if info, ok := d.images.manifest.Images[ImageName()]; ok {
		status.ImageVersion = info.Version
	}

	c, _ := d.findContainer(ctx, "core-dev")
	if c != nil {
		status.Running = c.Status == container.StatusRunning
		status.ContainerID = c.ID
		status.Memory = c.Memory
		status.CPUs = c.CPUs
		if status.Running {
			status.Uptime = time.Since(c.StartedAt)
		}
	}

	return status, nil
}
