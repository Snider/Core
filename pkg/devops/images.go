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
