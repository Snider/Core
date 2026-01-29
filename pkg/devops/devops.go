// Package devops provides a portable development environment using LinuxKit images.
package devops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/host-uk/core/pkg/container"
)

// DevOps manages the portable development environment.
type DevOps struct {
	config    *Config
	images    *ImageManager
	container *container.LinuxKitManager
}

// New creates a new DevOps instance.
func New() (*DevOps, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("devops.New: failed to load config: %w", err)
	}

	images, err := NewImageManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("devops.New: failed to create image manager: %w", err)
	}

	mgr, err := container.NewLinuxKitManager()
	if err != nil {
		return nil, fmt.Errorf("devops.New: failed to create container manager: %w", err)
	}

	return &DevOps{
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
	_, err = os.Stat(path)
	return err == nil
}

// Install downloads and installs the dev image.
func (d *DevOps) Install(ctx context.Context, progress func(downloaded, total int64)) error {
	return d.images.Install(ctx, progress)
}

// CheckUpdate checks if an update is available.
func (d *DevOps) CheckUpdate(ctx context.Context) (current, latest string, hasUpdate bool, err error) {
	return d.images.CheckUpdate(ctx)
}
