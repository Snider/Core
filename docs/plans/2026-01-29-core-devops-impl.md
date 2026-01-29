# Core DevOps CLI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement `core dev` commands for portable development environment using core-devops LinuxKit images.

**Architecture:** `pkg/devops` package handles image management, config, and orchestration. Reuses `pkg/container.LinuxKitManager` for VM lifecycle. Image sources (GitHub, Registry, CDN) implement common interface. CLI in `cmd/core/cmd/dev.go`.

**Tech Stack:** Go, pkg/container, golang.org/x/crypto/ssh, os/exec for gh CLI, YAML config

---

### Task 1: Create DevOps Package Structure

**Files:**
- Create: `pkg/devops/devops.go`
- Create: `pkg/devops/go.mod`

**Step 1: Create go.mod**

```go
module github.com/host-uk/core/pkg/devops

go 1.25

require (
	github.com/host-uk/core/pkg/container v0.0.0
	golang.org/x/crypto v0.32.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/host-uk/core/pkg/container => ../container
```

**Step 2: Create devops.go with core types**

```go
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
```

**Step 3: Add to go.work**

Run: `cd /Users/snider/Code/Core && echo "	./pkg/devops" >> go.work && go work sync`

**Step 4: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/devops/...`
Expected: Error (missing Config, ImageManager) - that's OK for now

**Step 5: Commit**

```bash
git add pkg/devops/
git add go.work go.work.sum
git commit -m "feat(devops): add package structure

Initial pkg/devops setup with DevOps type and path helpers.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 2: Implement Config Loading

**Files:**
- Create: `pkg/devops/config.go`
- Create: `pkg/devops/config_test.go`

**Step 1: Write the failing test**

```go
package devops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Good_Default(t *testing.T) {
	// Use temp home dir
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Images.Source != "auto" {
		t.Errorf("expected source 'auto', got %q", cfg.Images.Source)
	}
}

func TestLoadConfig_Good_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".core")
	os.MkdirAll(configDir, 0755)

	configContent := `version: 1
images:
  source: github
  github:
    repo: myorg/images
`
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0644)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Images.Source != "github" {
		t.Errorf("expected source 'github', got %q", cfg.Images.Source)
	}
	if cfg.Images.GitHub.Repo != "myorg/images" {
		t.Errorf("expected repo 'myorg/images', got %q", cfg.Images.GitHub.Repo)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/... -run TestLoadConfig -v`
Expected: FAIL (LoadConfig not defined)

**Step 3: Write implementation**

```go
package devops

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds global devops configuration from ~/.core/config.yaml.
type Config struct {
	Version int          `yaml:"version"`
	Images  ImagesConfig `yaml:"images"`
}

// ImagesConfig holds image source configuration.
type ImagesConfig struct {
	Source   string         `yaml:"source"` // auto, github, registry, cdn
	GitHub   GitHubConfig   `yaml:"github,omitempty"`
	Registry RegistryConfig `yaml:"registry,omitempty"`
	CDN      CDNConfig      `yaml:"cdn,omitempty"`
}

// GitHubConfig holds GitHub Releases configuration.
type GitHubConfig struct {
	Repo string `yaml:"repo"` // owner/repo format
}

// RegistryConfig holds container registry configuration.
type RegistryConfig struct {
	Image string `yaml:"image"` // e.g., ghcr.io/host-uk/core-devops
}

// CDNConfig holds CDN/S3 configuration.
type CDNConfig struct {
	URL string `yaml:"url"` // base URL for downloads
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Images: ImagesConfig{
			Source: "auto",
			GitHub: GitHubConfig{
				Repo: "host-uk/core-images",
			},
			Registry: RegistryConfig{
				Image: "ghcr.io/host-uk/core-devops",
			},
		},
	}
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".core", "config.yaml"), nil
}

// LoadConfig loads configuration from ~/.core/config.yaml.
// Returns default config if file doesn't exist.
func LoadConfig() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/... -run TestLoadConfig -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/devops/config.go pkg/devops/config_test.go
git commit -m "feat(devops): add config loading

Loads ~/.core/config.yaml with image source preferences.
Defaults to auto-detection with host-uk/core-images.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 3: Implement ImageSource Interface

**Files:**
- Create: `pkg/devops/sources/source.go`

**Step 1: Create source interface**

```go
// Package sources provides image download sources for core-devops.
package sources

import (
	"context"
)

// ImageSource defines the interface for downloading dev images.
type ImageSource interface {
	// Name returns the source identifier.
	Name() string
	// Available checks if this source can be used.
	Available() bool
	// LatestVersion returns the latest available version.
	LatestVersion(ctx context.Context) (string, error)
	// Download downloads the image to the destination path.
	// Reports progress via the callback if provided.
	Download(ctx context.Context, dest string, progress func(downloaded, total int64)) error
}

// SourceConfig holds configuration for a source.
type SourceConfig struct {
	// GitHub configuration
	GitHubRepo string
	// Registry configuration
	RegistryImage string
	// CDN configuration
	CDNURL string
	// Image name (e.g., core-devops-darwin-arm64.qcow2)
	ImageName string
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/devops/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/devops/sources/source.go
git commit -m "feat(devops): add ImageSource interface

Defines common interface for GitHub, Registry, and CDN sources.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 4: Implement GitHub Source

**Files:**
- Create: `pkg/devops/sources/github.go`
- Create: `pkg/devops/sources/github_test.go`

**Step 1: Write the failing test**

```go
package sources

import (
	"testing"
)

func TestGitHubSource_Good_Available(t *testing.T) {
	src := NewGitHubSource(SourceConfig{
		GitHubRepo: "host-uk/core-images",
		ImageName:  "core-devops-darwin-arm64.qcow2",
	})

	if src.Name() != "github" {
		t.Errorf("expected name 'github', got %q", src.Name())
	}

	// Available depends on gh CLI being installed
	_ = src.Available()
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/sources/... -run TestGitHubSource -v`
Expected: FAIL

**Step 3: Write implementation**

```go
package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GitHubSource downloads images from GitHub Releases.
type GitHubSource struct {
	config SourceConfig
}

// NewGitHubSource creates a new GitHub source.
func NewGitHubSource(cfg SourceConfig) *GitHubSource {
	return &GitHubSource{config: cfg}
}

// Name returns "github".
func (s *GitHubSource) Name() string {
	return "github"
}

// Available checks if gh CLI is installed and authenticated.
func (s *GitHubSource) Available() bool {
	_, err := exec.LookPath("gh")
	if err != nil {
		return false
	}
	// Check if authenticated
	cmd := exec.Command("gh", "auth", "status")
	return cmd.Run() == nil
}

// LatestVersion returns the latest release tag.
func (s *GitHubSource) LatestVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", "release", "view",
		"-R", s.config.GitHubRepo,
		"--json", "tagName",
		"-q", ".tagName",
	)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("github.LatestVersion: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Download downloads the image from the latest release.
func (s *GitHubSource) Download(ctx context.Context, dest string, progress func(downloaded, total int64)) error {
	// Get release assets to find our image
	cmd := exec.CommandContext(ctx, "gh", "release", "download",
		"-R", s.config.GitHubRepo,
		"-p", s.config.ImageName,
		"-D", dest,
		"--clobber",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("github.Download: %w", err)
	}
	return nil
}

// releaseAsset represents a GitHub release asset.
type releaseAsset struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	URL  string `json:"url"`
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/sources/... -run TestGitHubSource -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/devops/sources/github.go pkg/devops/sources/github_test.go
git commit -m "feat(devops): add GitHub Releases source

Downloads core-devops images from GitHub Releases using gh CLI.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 5: Implement CDN Source

**Files:**
- Create: `pkg/devops/sources/cdn.go`
- Create: `pkg/devops/sources/cdn_test.go`

**Step 1: Write the failing test**

```go
package sources

import (
	"testing"
)

func TestCDNSource_Good_Available(t *testing.T) {
	src := NewCDNSource(SourceConfig{
		CDNURL:    "https://images.example.com",
		ImageName: "core-devops-darwin-arm64.qcow2",
	})

	if src.Name() != "cdn" {
		t.Errorf("expected name 'cdn', got %q", src.Name())
	}

	// CDN is available if URL is configured
	if !src.Available() {
		t.Error("expected Available() to be true when URL is set")
	}
}

func TestCDNSource_Bad_NoURL(t *testing.T) {
	src := NewCDNSource(SourceConfig{
		ImageName: "core-devops-darwin-arm64.qcow2",
	})

	if src.Available() {
		t.Error("expected Available() to be false when URL is empty")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/sources/... -run TestCDNSource -v`
Expected: FAIL

**Step 3: Write implementation**

```go
package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// CDNSource downloads images from a CDN or S3 bucket.
type CDNSource struct {
	config SourceConfig
}

// NewCDNSource creates a new CDN source.
func NewCDNSource(cfg SourceConfig) *CDNSource {
	return &CDNSource{config: cfg}
}

// Name returns "cdn".
func (s *CDNSource) Name() string {
	return "cdn"
}

// Available checks if CDN URL is configured.
func (s *CDNSource) Available() bool {
	return s.config.CDNURL != ""
}

// LatestVersion fetches version from manifest or returns "latest".
func (s *CDNSource) LatestVersion(ctx context.Context) (string, error) {
	// Try to fetch manifest.json for version info
	url := fmt.Sprintf("%s/manifest.json", s.config.CDNURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "latest", nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "latest", nil
	}
	defer resp.Body.Close()

	// For now, just return latest - could parse manifest for version
	return "latest", nil
}

// Download downloads the image from CDN.
func (s *CDNSource) Download(ctx context.Context, dest string, progress func(downloaded, total int64)) error {
	url := fmt.Sprintf("%s/%s", s.config.CDNURL, s.config.ImageName)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("cdn.Download: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cdn.Download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("cdn.Download: HTTP %d", resp.StatusCode)
	}

	// Ensure dest directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("cdn.Download: %w", err)
	}

	// Create destination file
	destPath := filepath.Join(dest, s.config.ImageName)
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("cdn.Download: %w", err)
	}
	defer f.Close()

	// Copy with progress
	total := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return fmt.Errorf("cdn.Download: %w", werr)
			}
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("cdn.Download: %w", err)
		}
	}

	return nil
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/sources/... -run TestCDNSource -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/devops/sources/cdn.go pkg/devops/sources/cdn_test.go
git commit -m "feat(devops): add CDN/S3 source

Downloads core-devops images from custom CDN with progress reporting.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 6: Implement ImageManager

**Files:**
- Create: `pkg/devops/images.go`
- Create: `pkg/devops/images_test.go`

**Step 1: Write the failing test**

```go
package devops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImageManager_Good_IsInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Not installed yet
	if mgr.IsInstalled() {
		t.Error("expected IsInstalled() to be false")
	}

	// Create fake image
	imagePath := filepath.Join(tmpDir, ImageName())
	os.WriteFile(imagePath, []byte("fake"), 0644)

	// Now installed
	if !mgr.IsInstalled() {
		t.Error("expected IsInstalled() to be true")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/... -run TestImageManager -v`
Expected: FAIL

**Step 3: Write implementation**

```go
package devops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/host-uk/core/pkg/devops/sources"
)

// ImageManager handles image downloads and updates.
type ImageManager struct {
	config   *Config
	manifest *Manifest
	sources  []sources.ImageSource
}

// Manifest tracks installed images.
type Manifest struct {
	Images map[string]ImageInfo `json:"images"`
	path   string
}

// ImageInfo holds metadata about an installed image.
type ImageInfo struct {
	Version    string    `json:"version"`
	SHA256     string    `json:"sha256,omitempty"`
	Downloaded time.Time `json:"downloaded"`
	Source     string    `json:"source"`
}

// NewImageManager creates a new image manager.
func NewImageManager(cfg *Config) (*ImageManager, error) {
	imagesDir, err := ImagesDir()
	if err != nil {
		return nil, err
	}

	// Ensure images directory exists
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return nil, err
	}

	// Load or create manifest
	manifestPath := filepath.Join(imagesDir, "manifest.json")
	manifest, err := loadManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	// Build source list based on config
	imageName := ImageName()
	sourceCfg := sources.SourceConfig{
		GitHubRepo:    cfg.Images.GitHub.Repo,
		RegistryImage: cfg.Images.Registry.Image,
		CDNURL:        cfg.Images.CDN.URL,
		ImageName:     imageName,
	}

	var srcs []sources.ImageSource
	switch cfg.Images.Source {
	case "github":
		srcs = []sources.ImageSource{sources.NewGitHubSource(sourceCfg)}
	case "cdn":
		srcs = []sources.ImageSource{sources.NewCDNSource(sourceCfg)}
	default: // "auto"
		srcs = []sources.ImageSource{
			sources.NewGitHubSource(sourceCfg),
			sources.NewCDNSource(sourceCfg),
		}
	}

	return &ImageManager{
		config:   cfg,
		manifest: manifest,
		sources:  srcs,
	}, nil
}

// IsInstalled checks if the dev image is installed.
func (m *ImageManager) IsInstalled() bool {
	path, err := ImagePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Install downloads and installs the dev image.
func (m *ImageManager) Install(ctx context.Context, progress func(downloaded, total int64)) error {
	imagesDir, err := ImagesDir()
	if err != nil {
		return err
	}

	// Find first available source
	var src sources.ImageSource
	for _, s := range m.sources {
		if s.Available() {
			src = s
			break
		}
	}
	if src == nil {
		return fmt.Errorf("no image source available")
	}

	// Get version
	version, err := src.LatestVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	fmt.Printf("Downloading %s from %s...\n", ImageName(), src.Name())

	// Download
	if err := src.Download(ctx, imagesDir, progress); err != nil {
		return err
	}

	// Update manifest
	m.manifest.Images[ImageName()] = ImageInfo{
		Version:    version,
		Downloaded: time.Now(),
		Source:     src.Name(),
	}

	return m.manifest.Save()
}

// CheckUpdate checks if an update is available.
func (m *ImageManager) CheckUpdate(ctx context.Context) (current, latest string, hasUpdate bool, err error) {
	info, ok := m.manifest.Images[ImageName()]
	if !ok {
		return "", "", false, fmt.Errorf("image not installed")
	}
	current = info.Version

	// Find first available source
	var src sources.ImageSource
	for _, s := range m.sources {
		if s.Available() {
			src = s
			break
		}
	}
	if src == nil {
		return current, "", false, fmt.Errorf("no image source available")
	}

	latest, err = src.LatestVersion(ctx)
	if err != nil {
		return current, "", false, err
	}

	hasUpdate = current != latest
	return current, latest, hasUpdate, nil
}

func loadManifest(path string) (*Manifest, error) {
	m := &Manifest{
		Images: make(map[string]ImageInfo),
		path:   path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, m); err != nil {
		return nil, err
	}
	m.path = path

	return m, nil
}

// Save writes the manifest to disk.
func (m *Manifest) Save() error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0644)
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/... -run TestImageManager -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/devops/images.go pkg/devops/images_test.go
git commit -m "feat(devops): add ImageManager

Manages image downloads, manifest tracking, and update checking.
Tries sources in priority order (GitHub, CDN).

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 7: Implement Boot/Stop/Status

**Files:**
- Modify: `pkg/devops/devops.go`
- Create: `pkg/devops/devops_test.go`

**Step 1: Add boot/stop/status methods to devops.go**

```go
// Add to devops.go

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

	runOpts := container.RunOptions{
		Name:    opts.Name,
		Detach:  true,
		Memory:  opts.Memory,
		CPUs:    opts.CPUs,
		SSHPort: 2222,
	}

	_, err = d.container.Run(ctx, imagePath, runOpts)
	return err
}

// Stop stops the dev environment.
func (d *DevOps) Stop(ctx context.Context) error {
	containers, err := d.container.List(ctx)
	if err != nil {
		return err
	}

	for _, c := range containers {
		if c.Name == "core-dev" && c.Status == container.StatusRunning {
			return d.container.Stop(ctx, c.ID)
		}
	}

	return nil
}

// IsRunning checks if the dev environment is running.
func (d *DevOps) IsRunning(ctx context.Context) (bool, error) {
	containers, err := d.container.List(ctx)
	if err != nil {
		return false, err
	}

	for _, c := range containers {
		if c.Name == "core-dev" && c.Status == container.StatusRunning {
			return true, nil
		}
	}

	return false, nil
}

// Status returns information about the dev environment.
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
	}

	if info, ok := d.images.manifest.Images[ImageName()]; ok {
		status.ImageVersion = info.Version
	}

	containers, err := d.container.List(ctx)
	if err != nil {
		return status, nil
	}

	for _, c := range containers {
		if c.Name == "core-dev" && c.Status == container.StatusRunning {
			status.Running = true
			status.ContainerID = c.ID
			status.Memory = c.Memory
			status.CPUs = c.CPUs
			status.SSHPort = 2222
			status.Uptime = time.Since(c.StartedAt)
			break
		}
	}

	return status, nil
}
```

**Step 2: Add missing import to devops.go**

```go
import (
	"time"
	// ... other imports
)
```

**Step 3: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/devops/...`
Expected: No errors

**Step 4: Commit**

```bash
git add pkg/devops/devops.go
git commit -m "feat(devops): add Boot/Stop/Status methods

Manages dev VM lifecycle using LinuxKitManager.
Supports fresh boot, status checking, graceful stop.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 8: Implement Shell Command

**Files:**
- Create: `pkg/devops/shell.go`

**Step 1: Create shell.go**

```go
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
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-A", // Agent forwarding
		"-p", "2222",
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
	containers, err := d.container.List(ctx)
	if err != nil {
		return err
	}

	for _, c := range containers {
		if c.Name == "core-dev" {
			// Use socat to connect to the console socket
			socketPath := fmt.Sprintf("/tmp/core-%s-console.sock", c.ID)
			cmd := exec.CommandContext(ctx, "socat", "-,raw,echo=0", "unix-connect:"+socketPath)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
	}

	return fmt.Errorf("console not available")
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/devops/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/devops/shell.go
git commit -m "feat(devops): add Shell for SSH and console access

Connects to dev VM via SSH (default) or serial console (--console).
Supports SSH agent forwarding for credential access.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 9: Implement Test Detection

**Files:**
- Create: `pkg/devops/test.go`
- Create: `pkg/devops/test_test.go`

**Step 1: Write the failing test**

```go
package devops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectTestCommand_Good_ComposerJSON(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(`{"scripts":{"test":"pest"}}`), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "composer test" {
		t.Errorf("expected 'composer test', got %q", cmd)
	}
}

func TestDetectTestCommand_Good_PackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"scripts":{"test":"vitest"}}`), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "npm test" {
		t.Errorf("expected 'npm test', got %q", cmd)
	}
}

func TestDetectTestCommand_Good_GoMod(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example"), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "go test ./..." {
		t.Errorf("expected 'go test ./...', got %q", cmd)
	}
}

func TestDetectTestCommand_Good_CoreTestYaml(t *testing.T) {
	tmpDir := t.TempDir()
	coreDir := filepath.Join(tmpDir, ".core")
	os.MkdirAll(coreDir, 0755)
	os.WriteFile(filepath.Join(coreDir, "test.yaml"), []byte("command: custom-test"), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "custom-test" {
		t.Errorf("expected 'custom-test', got %q", cmd)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/... -run TestDetectTestCommand -v`
Expected: FAIL

**Step 3: Write implementation**

```go
package devops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TestConfig holds test configuration from .core/test.yaml.
type TestConfig struct {
	Version  int           `yaml:"version"`
	Command  string        `yaml:"command,omitempty"`
	Commands []TestCommand `yaml:"commands,omitempty"`
	Env      map[string]string `yaml:"env,omitempty"`
}

// TestCommand is a named test command.
type TestCommand struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

// TestOptions configures test execution.
type TestOptions struct {
	Name    string   // Run specific named command from .core/test.yaml
	Command []string // Override command (from -- args)
}

// Test runs tests in the dev environment.
func (d *DevOps) Test(ctx context.Context, projectDir string, opts TestOptions) error {
	running, err := d.IsRunning(ctx)
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("dev environment not running (run 'core dev boot' first)")
	}

	var cmd string

	// Priority: explicit command > named command > auto-detect
	if len(opts.Command) > 0 {
		cmd = joinCommand(opts.Command)
	} else if opts.Name != "" {
		cfg, err := LoadTestConfig(projectDir)
		if err != nil {
			return err
		}
		for _, c := range cfg.Commands {
			if c.Name == opts.Name {
				cmd = c.Run
				break
			}
		}
		if cmd == "" {
			return fmt.Errorf("test command %q not found in .core/test.yaml", opts.Name)
		}
	} else {
		cmd = DetectTestCommand(projectDir)
		if cmd == "" {
			return fmt.Errorf("could not detect test command (create .core/test.yaml)")
		}
	}

	// Run via SSH
	return d.sshShell(ctx, []string{"cd", "/app", "&&", cmd})
}

// DetectTestCommand auto-detects the test command for a project.
func DetectTestCommand(projectDir string) string {
	// 1. Check .core/test.yaml
	cfg, err := LoadTestConfig(projectDir)
	if err == nil && cfg.Command != "" {
		return cfg.Command
	}

	// 2. Check composer.json
	if hasFile(projectDir, "composer.json") {
		return "composer test"
	}

	// 3. Check package.json
	if hasFile(projectDir, "package.json") {
		return "npm test"
	}

	// 4. Check go.mod
	if hasFile(projectDir, "go.mod") {
		return "go test ./..."
	}

	// 5. Check pytest
	if hasFile(projectDir, "pytest.ini") || hasFile(projectDir, "pyproject.toml") {
		return "pytest"
	}

	// 6. Check Taskfile
	if hasFile(projectDir, "Taskfile.yaml") || hasFile(projectDir, "Taskfile.yml") {
		return "task test"
	}

	return ""
}

// LoadTestConfig loads .core/test.yaml.
func LoadTestConfig(projectDir string) (*TestConfig, error) {
	path := filepath.Join(projectDir, ".core", "test.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg TestConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func hasFile(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func joinCommand(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/... -run TestDetectTestCommand -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/devops/test.go pkg/devops/test_test.go
git commit -m "feat(devops): add test detection and execution

Auto-detects test framework from project files.
Supports .core/test.yaml for custom configuration.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 10: Implement Serve with Mount

**Files:**
- Create: `pkg/devops/serve.go`

**Step 1: Create serve.go**

```go
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
		"-R", "10000:localhost:22", // Reverse tunnel for SSHFS
		"-p", "2222",
		"root@localhost",
		"mkdir -p /app && sshfs -p 10000 "+os.Getenv("USER")+"@localhost:"+absPath+" /app -o allow_other",
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

	// Python
	if hasFile(projectDir, "manage.py") {
		return "python manage.py runserver 0.0.0.0:8000"
	}

	// Fallback: simple HTTP server
	return "python3 -m http.server 8000"
}

func hasPackageScript(projectDir, script string) bool {
	data, err := os.ReadFile(filepath.Join(projectDir, "package.json"))
	if err != nil {
		return false
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	_, ok := pkg.Scripts[script]
	return ok
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/devops/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/devops/serve.go
git commit -m "feat(devops): add Serve with project mounting

Mounts project via SSHFS and runs auto-detected dev server.
Supports Laravel, Node.js, PHP, Go, Python projects.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 11: Implement Claude Sandbox

**Files:**
- Create: `pkg/devops/claude.go`

**Step 1: Create claude.go**

```go
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

	// Add environment variables
	for _, env := range envVars {
		args = append(args, "-o", "SendEnv="+strings.Split(env, "=")[0])
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
	cmd.Env = append(os.Environ(), envVars...)

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
		"-P", "2222",
		"-r", ghConfigDir,
		"root@localhost:/root/.config/",
	)
	return cmd.Run()
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/devops/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/devops/claude.go
git commit -m "feat(devops): add Claude sandbox session

Starts Claude in immutable dev environment with auth forwarding.
Auto-boots VM, mounts project, forwards credentials.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 12: Add CLI Commands

**Files:**
- Create: `cmd/core/cmd/dev.go`
- Modify: `cmd/core/cmd/root.go`

**Step 1: Create dev.go**

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/devops"
	"github.com/leaanthony/clir"
)

var (
	devHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#3b82f6"))

	devSuccessStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#22c55e"))

	devErrorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ef4444"))

	devDimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280"))
)

// AddDevCommand adds the dev command group.
func AddDevCommand(app *clir.Cli) {
	devCmd := app.NewSubCommand("dev", "Portable development environment")
	devCmd.LongDescription("Manage the core-devops portable development environment.\n" +
		"A sandboxed, immutable Linux VM with 100+ development tools.")

	addDevInstallCommand(devCmd)
	addDevBootCommand(devCmd)
	addDevStopCommand(devCmd)
	addDevStatusCommand(devCmd)
	addDevShellCommand(devCmd)
	addDevServeCommand(devCmd)
	addDevTestCommand(devCmd)
	addDevClaudeCommand(devCmd)
	addDevUpdateCommand(devCmd)
}

func addDevInstallCommand(parent *clir.Cli) {
	var source string
	cmd := parent.NewSubCommand("install", "Download the dev environment image")
	cmd.StringFlag("source", "Image source: auto, github, registry, cdn", &source)

	cmd.Action(func() error {
		ctx := context.Background()
		d, err := devops.New()
		if err != nil {
			return err
		}

		if d.IsInstalled() {
			fmt.Printf("%s Dev image already installed\n", devSuccessStyle.Render("OK:"))
			fmt.Println("Use 'core dev update' to check for updates")
			return nil
		}

		fmt.Printf("%s Downloading dev image...\n", devHeaderStyle.Render("Install:"))

		progress := func(downloaded, total int64) {
			if total > 0 {
				pct := float64(downloaded) / float64(total) * 100
				fmt.Printf("\r  %.1f%% (%d / %d MB)", pct, downloaded/1024/1024, total/1024/1024)
			}
		}

		if err := d.Install(ctx, progress); err != nil {
			return err
		}

		fmt.Println()
		fmt.Printf("%s Dev image installed\n", devSuccessStyle.Render("Success:"))
		return nil
	})
}

func addDevBootCommand(parent *clir.Cli) {
	var memory, cpus int
	var fresh bool

	cmd := parent.NewSubCommand("boot", "Start the dev environment")
	cmd.IntFlag("memory", "Memory in MB (default: 4096)", &memory)
	cmd.IntFlag("cpus", "Number of CPUs (default: 2)", &cpus)
	cmd.BoolFlag("fresh", "Destroy existing and start fresh", &fresh)

	cmd.Action(func() error {
		ctx := context.Background()
		d, err := devops.New()
		if err != nil {
			return err
		}

		opts := devops.DefaultBootOptions()
		if memory > 0 {
			opts.Memory = memory
		}
		if cpus > 0 {
			opts.CPUs = cpus
		}
		opts.Fresh = fresh

		fmt.Printf("%s Starting dev environment...\n", devHeaderStyle.Render("Boot:"))

		if err := d.Boot(ctx, opts); err != nil {
			return err
		}

		fmt.Printf("%s Dev environment running\n", devSuccessStyle.Render("Success:"))
		fmt.Printf("  Memory: %d MB\n", opts.Memory)
		fmt.Printf("  CPUs:   %d\n", opts.CPUs)
		fmt.Printf("  SSH:    ssh -p 2222 root@localhost\n")
		return nil
	})
}

func addDevStopCommand(parent *clir.Cli) {
	cmd := parent.NewSubCommand("stop", "Stop the dev environment")
	cmd.Action(func() error {
		ctx := context.Background()
		d, err := devops.New()
		if err != nil {
			return err
		}

		fmt.Printf("%s Stopping dev environment...\n", devHeaderStyle.Render("Stop:"))

		if err := d.Stop(ctx); err != nil {
			return err
		}

		fmt.Printf("%s Dev environment stopped\n", devSuccessStyle.Render("Success:"))
		return nil
	})
}

func addDevStatusCommand(parent *clir.Cli) {
	cmd := parent.NewSubCommand("status", "Show dev environment status")
	cmd.Action(func() error {
		ctx := context.Background()
		d, err := devops.New()
		if err != nil {
			return err
		}

		status, err := d.Status(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("%s Dev Environment\n\n", devHeaderStyle.Render("Status:"))

		if status.Installed {
			fmt.Printf("  Image:   %s\n", devSuccessStyle.Render("installed"))
			fmt.Printf("  Version: %s\n", status.ImageVersion)
		} else {
			fmt.Printf("  Image:   %s\n", devDimStyle.Render("not installed"))
		}

		if status.Running {
			fmt.Printf("  Status:  %s\n", devSuccessStyle.Render("running"))
			fmt.Printf("  ID:      %s\n", status.ContainerID[:8])
			fmt.Printf("  Memory:  %d MB\n", status.Memory)
			fmt.Printf("  CPUs:    %d\n", status.CPUs)
			fmt.Printf("  SSH:     port %d\n", status.SSHPort)
			fmt.Printf("  Uptime:  %s\n", status.Uptime.Round(1000000000))
		} else {
			fmt.Printf("  Status:  %s\n", devDimStyle.Render("stopped"))
		}

		return nil
	})
}

func addDevShellCommand(parent *clir.Cli) {
	var console bool
	cmd := parent.NewSubCommand("shell", "Open a shell in the dev environment")
	cmd.BoolFlag("console", "Use serial console instead of SSH", &console)

	cmd.Action(func() error {
		ctx := context.Background()
		d, err := devops.New()
		if err != nil {
			return err
		}

		return d.Shell(ctx, devops.ShellOptions{Console: console})
	})
}

func addDevServeCommand(parent *clir.Cli) {
	var port int
	var path string

	cmd := parent.NewSubCommand("serve", "Mount project and start dev server")
	cmd.IntFlag("port", "Port to serve on (default: 8000)", &port)
	cmd.StringFlag("path", "Subdirectory to serve", &path)

	cmd.Action(func() error {
		ctx := context.Background()
		d, err := devops.New()
		if err != nil {
			return err
		}

		projectDir, _ := os.Getwd()
		return d.Serve(ctx, projectDir, devops.ServeOptions{Port: port, Path: path})
	})
}

func addDevTestCommand(parent *clir.Cli) {
	var name string

	cmd := parent.NewSubCommand("test", "Run tests in dev environment")
	cmd.StringFlag("name", "Run specific named test from .core/test.yaml", &name)

	cmd.Action(func() error {
		ctx := context.Background()
		d, err := devops.New()
		if err != nil {
			return err
		}

		projectDir, _ := os.Getwd()
		args := cmd.OtherArgs()

		return d.Test(ctx, projectDir, devops.TestOptions{
			Name:    name,
			Command: args,
		})
	})
}

func addDevClaudeCommand(parent *clir.Cli) {
	var noAuth bool
	var auth string
	var model string

	cmd := parent.NewSubCommand("claude", "Start Claude in sandboxed dev environment")
	cmd.BoolFlag("no-auth", "Don't forward any credentials", &noAuth)
	cmd.StringFlag("auth", "Selective auth forwarding: gh,anthropic,ssh,git", &auth)
	cmd.StringFlag("model", "Model to use: opus, sonnet", &model)

	cmd.Action(func() error {
		ctx := context.Background()
		d, err := devops.New()
		if err != nil {
			return err
		}

		projectDir, _ := os.Getwd()

		var authList []string
		if auth != "" {
			authList = strings.Split(auth, ",")
		}

		return d.Claude(ctx, projectDir, devops.ClaudeOptions{
			NoAuth: noAuth,
			Auth:   authList,
			Model:  model,
		})
	})
}

func addDevUpdateCommand(parent *clir.Cli) {
	var force bool
	cmd := parent.NewSubCommand("update", "Check for and download image updates")
	cmd.BoolFlag("force", "Force download even if up to date", &force)

	cmd.Action(func() error {
		ctx := context.Background()
		d, err := devops.New()
		if err != nil {
			return err
		}

		if !d.IsInstalled() {
			return fmt.Errorf("dev image not installed (run 'core dev install' first)")
		}

		fmt.Printf("%s Checking for updates...\n", devHeaderStyle.Render("Update:"))

		current, latest, hasUpdate, err := d.CheckUpdate(ctx)
		if err != nil {
			return err
		}

		if !hasUpdate && !force {
			fmt.Printf("%s Already up to date (%s)\n", devSuccessStyle.Render("OK:"), current)
			return nil
		}

		fmt.Printf("  Current: %s\n", current)
		fmt.Printf("  Latest:  %s\n", latest)

		progress := func(downloaded, total int64) {
			if total > 0 {
				pct := float64(downloaded) / float64(total) * 100
				fmt.Printf("\r  Downloading: %.1f%%", pct)
			}
		}

		if err := d.Install(ctx, progress); err != nil {
			return err
		}

		fmt.Println()
		fmt.Printf("%s Updated to %s\n", devSuccessStyle.Render("Success:"), latest)
		return nil
	})
}
```

**Step 2: Add to root.go**

Add after other command registrations:
```go
AddDevCommand(app)
```

**Step 3: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./cmd/core/...`
Expected: No errors

**Step 4: Commit**

```bash
git add cmd/core/cmd/dev.go cmd/core/cmd/root.go
git commit -m "feat(cli): add dev command group

Commands:
- core dev install/boot/stop/status
- core dev shell/serve/test
- core dev claude (sandboxed AI session)
- core dev update

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 13: Final Integration Test

**Step 1: Build CLI**

Run: `cd /Users/snider/Code/Core && go build -o bin/core ./cmd/core`
Expected: No errors

**Step 2: Test help output**

Run: `./bin/core dev --help`
Expected: Shows all dev subcommands

**Step 3: Run package tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/devops/... -v`
Expected: All tests pass

**Step 4: Update TODO.md**

Mark S4.6 tasks as complete in tasks/TODO.md

**Step 5: Final commit**

```bash
git add -A
git commit -m "chore(devops): finalize S4.6 core-devops CLI

All dev commands implemented:
- install/boot/stop/status
- shell/serve/test
- claude (sandboxed AI session)
- update

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Summary

13 tasks covering:
1. Package structure
2. Config loading
3. ImageSource interface
4. GitHub source
5. CDN source
6. ImageManager
7. Boot/Stop/Status
8. Shell command
9. Test detection
10. Serve with mount
11. Claude sandbox
12. CLI commands
13. Integration test
