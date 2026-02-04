package devops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/host-uk/core/pkg/devops/sources"
	"github.com/host-uk/core/pkg/io"
)

// ImageManager handles image downloads and updates.
type ImageManager struct {
	medium   io.Medium
	config   *Config
	manifest *Manifest
	sources  []sources.ImageSource
}

// Manifest tracks installed images.
type Manifest struct {
	medium io.Medium
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
func NewImageManager(m io.Medium, cfg *Config) (*ImageManager, error) {
	imagesDir, err := ImagesDir()
	if err != nil {
		return nil, err
	}

	// Ensure images directory exists
	if err := m.EnsureDir(imagesDir); err != nil {
		return nil, err
	}

	// Load or create manifest
	manifestPath := filepath.Join(imagesDir, "manifest.json")
	manifest, err := loadManifest(m, manifestPath)
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
		medium:   m,
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
	return m.medium.IsFile(path)
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
	if err := src.Download(ctx, m.medium, imagesDir, progress); err != nil {
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

func loadManifest(m io.Medium, path string) (*Manifest, error) {
	manifest := &Manifest{
		medium: m,
		Images: make(map[string]ImageInfo),
		path:   path,
	}

	content, err := m.Read(path)
	if err != nil {
		if os.IsNotExist(err) {
			return manifest, nil
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(content), manifest); err != nil {
		return nil, err
	}
	manifest.medium = m
	manifest.path = path

	return manifest, nil
}

// Save writes the manifest to disk.
func (m *Manifest) Save() error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return m.medium.Write(m.path, string(data))
}
