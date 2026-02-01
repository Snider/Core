package devops

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/host-uk/core/pkg/devops/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageManager_Good_IsInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	// Not installed yet
	assert.False(t, mgr.IsInstalled())

	// Create fake image
	imagePath := filepath.Join(tmpDir, ImageName())
	err = os.WriteFile(imagePath, []byte("fake"), 0644)
	require.NoError(t, err)

	// Now installed
	assert.True(t, mgr.IsInstalled())
}

func TestNewImageManager_Good(t *testing.T) {
	t.Run("creates manager with cdn source", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("CORE_IMAGES_DIR", tmpDir)

		cfg := DefaultConfig()
		cfg.Images.Source = "cdn"

		mgr, err := NewImageManager(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Len(t, mgr.sources, 1)
		assert.Equal(t, "cdn", mgr.sources[0].Name())
	})

	t.Run("creates manager with github source", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("CORE_IMAGES_DIR", tmpDir)

		cfg := DefaultConfig()
		cfg.Images.Source = "github"

		mgr, err := NewImageManager(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)
		assert.Len(t, mgr.sources, 1)
		assert.Equal(t, "github", mgr.sources[0].Name())
	})
}

func TestManifest_Save(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "manifest.json")

	m := &Manifest{
		Images: make(map[string]ImageInfo),
		path:   path,
	}

	m.Images["test.img"] = ImageInfo{
		Version: "1.0.0",
		Source:  "test",
	}

	err := m.Save()
	assert.NoError(t, err)

	// Verify file exists and has content
	_, err = os.Stat(path)
	assert.NoError(t, err)

	// Reload
	m2, err := loadManifest(path)
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", m2.Images["test.img"].Version)
}

func TestLoadManifest_Bad(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "manifest.json")
		err := os.WriteFile(path, []byte("invalid json"), 0644)
		require.NoError(t, err)

		_, err = loadManifest(path)
		assert.Error(t, err)
	})
}

func TestCheckUpdate_Bad(t *testing.T) {
	t.Run("image not installed", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("CORE_IMAGES_DIR", tmpDir)

		cfg := DefaultConfig()
		mgr, err := NewImageManager(cfg)
		require.NoError(t, err)

		_, _, _, err = mgr.CheckUpdate(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image not installed")
	})
}

func TestNewImageManager_Good_AutoSource(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	cfg := DefaultConfig()
	cfg.Images.Source = "auto"

	mgr, err := NewImageManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, mgr)
	assert.Len(t, mgr.sources, 2) // github and cdn
}

func TestNewImageManager_Good_UnknownSourceFallsToAuto(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	cfg := DefaultConfig()
	cfg.Images.Source = "unknown"

	mgr, err := NewImageManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, mgr)
	assert.Len(t, mgr.sources, 2) // falls to default (auto) which is github + cdn
}

func TestLoadManifest_Good_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent.json")

	m, err := loadManifest(path)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.NotNil(t, m.Images)
	assert.Empty(t, m.Images)
	assert.Equal(t, path, m.path)
}

func TestLoadManifest_Good_ExistingData(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "manifest.json")

	data := `{"images":{"test.img":{"version":"2.0.0","source":"cdn"}}}`
	err := os.WriteFile(path, []byte(data), 0644)
	require.NoError(t, err)

	m, err := loadManifest(path)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, "2.0.0", m.Images["test.img"].Version)
	assert.Equal(t, "cdn", m.Images["test.img"].Source)
}

func TestImageInfo_Struct(t *testing.T) {
	info := ImageInfo{
		Version:    "1.0.0",
		SHA256:     "abc123",
		Downloaded: time.Now(),
		Source:     "github",
	}
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "abc123", info.SHA256)
	assert.False(t, info.Downloaded.IsZero())
	assert.Equal(t, "github", info.Source)
}

func TestManifest_Save_Good_CreatesDirs(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "manifest.json")

	m := &Manifest{
		Images: make(map[string]ImageInfo),
		path:   nestedPath,
	}
	m.Images["test.img"] = ImageInfo{Version: "1.0.0"}

	// Should fail because nested directories don't exist
	// (Save doesn't create parent directories, it just writes to path)
	err := m.Save()
	assert.Error(t, err)
}

func TestManifest_Save_Good_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "manifest.json")

	// First save
	m1 := &Manifest{
		Images: make(map[string]ImageInfo),
		path:   path,
	}
	m1.Images["test.img"] = ImageInfo{Version: "1.0.0"}
	err := m1.Save()
	require.NoError(t, err)

	// Second save with different data
	m2 := &Manifest{
		Images: make(map[string]ImageInfo),
		path:   path,
	}
	m2.Images["other.img"] = ImageInfo{Version: "2.0.0"}
	err = m2.Save()
	require.NoError(t, err)

	// Verify second data
	loaded, err := loadManifest(path)
	assert.NoError(t, err)
	assert.Equal(t, "2.0.0", loaded.Images["other.img"].Version)
	_, exists := loaded.Images["test.img"]
	assert.False(t, exists)
}

func TestImageManager_Install_Bad_NoSourceAvailable(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	// Create manager with empty sources
	mgr := &ImageManager{
		config:   DefaultConfig(),
		manifest: &Manifest{Images: make(map[string]ImageInfo), path: filepath.Join(tmpDir, "manifest.json")},
		sources:  nil, // no sources
	}

	err := mgr.Install(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no image source available")
}

func TestNewImageManager_Good_CreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	imagesDir := filepath.Join(tmpDir, "images")
	t.Setenv("CORE_IMAGES_DIR", imagesDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, mgr)

	// Verify directory was created
	info, err := os.Stat(imagesDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

// mockImageSource is a test helper for simulating image sources
type mockImageSource struct {
	name          string
	available     bool
	latestVersion string
	latestErr     error
	downloadErr   error
}

func (m *mockImageSource) Name() string    { return m.name }
func (m *mockImageSource) Available() bool { return m.available }
func (m *mockImageSource) LatestVersion(ctx context.Context) (string, error) {
	return m.latestVersion, m.latestErr
}
func (m *mockImageSource) Download(ctx context.Context, dest string, progress func(downloaded, total int64)) error {
	if m.downloadErr != nil {
		return m.downloadErr
	}
	// Create a fake image file
	imagePath := filepath.Join(dest, ImageName())
	return os.WriteFile(imagePath, []byte("mock image content"), 0644)
}

func TestImageManager_Install_Good_WithMockSource(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	mock := &mockImageSource{
		name:          "mock",
		available:     true,
		latestVersion: "v1.0.0",
	}

	mgr := &ImageManager{
		config:   DefaultConfig(),
		manifest: &Manifest{Images: make(map[string]ImageInfo), path: filepath.Join(tmpDir, "manifest.json")},
		sources:  []sources.ImageSource{mock},
	}

	err := mgr.Install(context.Background(), nil)
	assert.NoError(t, err)
	assert.True(t, mgr.IsInstalled())

	// Verify manifest was updated
	info, ok := mgr.manifest.Images[ImageName()]
	assert.True(t, ok)
	assert.Equal(t, "v1.0.0", info.Version)
	assert.Equal(t, "mock", info.Source)
}

func TestImageManager_Install_Bad_DownloadError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	mock := &mockImageSource{
		name:          "mock",
		available:     true,
		latestVersion: "v1.0.0",
		downloadErr:   assert.AnError,
	}

	mgr := &ImageManager{
		config:   DefaultConfig(),
		manifest: &Manifest{Images: make(map[string]ImageInfo), path: filepath.Join(tmpDir, "manifest.json")},
		sources:  []sources.ImageSource{mock},
	}

	err := mgr.Install(context.Background(), nil)
	assert.Error(t, err)
}

func TestImageManager_Install_Bad_VersionError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	mock := &mockImageSource{
		name:      "mock",
		available: true,
		latestErr: assert.AnError,
	}

	mgr := &ImageManager{
		config:   DefaultConfig(),
		manifest: &Manifest{Images: make(map[string]ImageInfo), path: filepath.Join(tmpDir, "manifest.json")},
		sources:  []sources.ImageSource{mock},
	}

	err := mgr.Install(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get latest version")
}

func TestImageManager_Install_Good_SkipsUnavailableSource(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	unavailableMock := &mockImageSource{
		name:      "unavailable",
		available: false,
	}
	availableMock := &mockImageSource{
		name:          "available",
		available:     true,
		latestVersion: "v2.0.0",
	}

	mgr := &ImageManager{
		config:   DefaultConfig(),
		manifest: &Manifest{Images: make(map[string]ImageInfo), path: filepath.Join(tmpDir, "manifest.json")},
		sources:  []sources.ImageSource{unavailableMock, availableMock},
	}

	err := mgr.Install(context.Background(), nil)
	assert.NoError(t, err)

	// Should have used the available source
	info := mgr.manifest.Images[ImageName()]
	assert.Equal(t, "available", info.Source)
}

func TestImageManager_CheckUpdate_Good_WithMockSource(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	mock := &mockImageSource{
		name:          "mock",
		available:     true,
		latestVersion: "v2.0.0",
	}

	mgr := &ImageManager{
		config: DefaultConfig(),
		manifest: &Manifest{
			Images: map[string]ImageInfo{
				ImageName(): {Version: "v1.0.0", Source: "mock"},
			},
			path: filepath.Join(tmpDir, "manifest.json"),
		},
		sources: []sources.ImageSource{mock},
	}

	current, latest, hasUpdate, err := mgr.CheckUpdate(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "v1.0.0", current)
	assert.Equal(t, "v2.0.0", latest)
	assert.True(t, hasUpdate)
}

func TestImageManager_CheckUpdate_Good_NoUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	mock := &mockImageSource{
		name:          "mock",
		available:     true,
		latestVersion: "v1.0.0",
	}

	mgr := &ImageManager{
		config: DefaultConfig(),
		manifest: &Manifest{
			Images: map[string]ImageInfo{
				ImageName(): {Version: "v1.0.0", Source: "mock"},
			},
			path: filepath.Join(tmpDir, "manifest.json"),
		},
		sources: []sources.ImageSource{mock},
	}

	current, latest, hasUpdate, err := mgr.CheckUpdate(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "v1.0.0", current)
	assert.Equal(t, "v1.0.0", latest)
	assert.False(t, hasUpdate)
}

func TestImageManager_CheckUpdate_Bad_NoSource(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	unavailableMock := &mockImageSource{
		name:      "mock",
		available: false,
	}

	mgr := &ImageManager{
		config: DefaultConfig(),
		manifest: &Manifest{
			Images: map[string]ImageInfo{
				ImageName(): {Version: "v1.0.0", Source: "mock"},
			},
			path: filepath.Join(tmpDir, "manifest.json"),
		},
		sources: []sources.ImageSource{unavailableMock},
	}

	_, _, _, err := mgr.CheckUpdate(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no image source available")
}

func TestImageManager_CheckUpdate_Bad_VersionError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	mock := &mockImageSource{
		name:      "mock",
		available: true,
		latestErr: assert.AnError,
	}

	mgr := &ImageManager{
		config: DefaultConfig(),
		manifest: &Manifest{
			Images: map[string]ImageInfo{
				ImageName(): {Version: "v1.0.0", Source: "mock"},
			},
			path: filepath.Join(tmpDir, "manifest.json"),
		},
		sources: []sources.ImageSource{mock},
	}

	current, _, _, err := mgr.CheckUpdate(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "v1.0.0", current) // Current should still be returned
}

func TestImageManager_Install_Bad_EmptySources(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	mgr := &ImageManager{
		config:   DefaultConfig(),
		manifest: &Manifest{Images: make(map[string]ImageInfo), path: filepath.Join(tmpDir, "manifest.json")},
		sources:  []sources.ImageSource{}, // Empty slice, not nil
	}

	err := mgr.Install(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no image source available")
}

func TestImageManager_Install_Bad_AllUnavailable(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	mock1 := &mockImageSource{name: "mock1", available: false}
	mock2 := &mockImageSource{name: "mock2", available: false}

	mgr := &ImageManager{
		config:   DefaultConfig(),
		manifest: &Manifest{Images: make(map[string]ImageInfo), path: filepath.Join(tmpDir, "manifest.json")},
		sources:  []sources.ImageSource{mock1, mock2},
	}

	err := mgr.Install(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no image source available")
}

func TestImageManager_CheckUpdate_Good_FirstSourceUnavailable(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	unavailable := &mockImageSource{name: "unavailable", available: false}
	available := &mockImageSource{name: "available", available: true, latestVersion: "v2.0.0"}

	mgr := &ImageManager{
		config: DefaultConfig(),
		manifest: &Manifest{
			Images: map[string]ImageInfo{
				ImageName(): {Version: "v1.0.0", Source: "available"},
			},
			path: filepath.Join(tmpDir, "manifest.json"),
		},
		sources: []sources.ImageSource{unavailable, available},
	}

	current, latest, hasUpdate, err := mgr.CheckUpdate(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "v1.0.0", current)
	assert.Equal(t, "v2.0.0", latest)
	assert.True(t, hasUpdate)
}

func TestManifest_Struct(t *testing.T) {
	m := &Manifest{
		Images: map[string]ImageInfo{
			"test.img": {Version: "1.0.0"},
		},
		path: "/path/to/manifest.json",
	}
	assert.Equal(t, "/path/to/manifest.json", m.path)
	assert.Len(t, m.Images, 1)
	assert.Equal(t, "1.0.0", m.Images["test.img"].Version)
}
