package sources

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCDNSource_Good_Available(t *testing.T) {
	src := NewCDNSource(SourceConfig{
		CDNURL:    "https://images.example.com",
		ImageName: "core-devops-darwin-arm64.qcow2",
	})

	assert.Equal(t, "cdn", src.Name())
	assert.True(t, src.Available())
}

func TestCDNSource_Bad_NoURL(t *testing.T) {
	src := NewCDNSource(SourceConfig{
		ImageName: "core-devops-darwin-arm64.qcow2",
	})

	assert.False(t, src.Available())
}

func TestCDNSource_LatestVersion_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manifest.json" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"version": "1.2.3"}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	src := NewCDNSource(SourceConfig{
		CDNURL:    server.URL,
		ImageName: "test.img",
	})

	version, err := src.LatestVersion(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "latest", version) // Current impl always returns "latest"
}

func TestCDNSource_Download_Good(t *testing.T) {
	content := "fake image data"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test.img" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, content)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	dest := t.TempDir()
	imageName := "test.img"
	src := NewCDNSource(SourceConfig{
		CDNURL:    server.URL,
		ImageName: imageName,
	})

	var progressCalled bool
	err := src.Download(context.Background(), dest, func(downloaded, total int64) {
		progressCalled = true
	})

	assert.NoError(t, err)
	assert.True(t, progressCalled)

	// Verify file content
	data, err := os.ReadFile(filepath.Join(dest, imageName))
	assert.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestCDNSource_Download_Bad(t *testing.T) {
	t.Run("HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		dest := t.TempDir()
		src := NewCDNSource(SourceConfig{
			CDNURL:    server.URL,
			ImageName: "test.img",
		})

		err := src.Download(context.Background(), dest, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 500")
	})

	t.Run("Invalid URL", func(t *testing.T) {
		dest := t.TempDir()
		src := NewCDNSource(SourceConfig{
			CDNURL:    "http://invalid-url-that-should-fail",
			ImageName: "test.img",
		})

		err := src.Download(context.Background(), dest, nil)
		assert.Error(t, err)
	})
}

func TestCDNSource_LatestVersion_Bad_NoManifest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	src := NewCDNSource(SourceConfig{
		CDNURL:    server.URL,
		ImageName: "test.img",
	})

	version, err := src.LatestVersion(context.Background())
	assert.NoError(t, err) // Should not error, just return "latest"
	assert.Equal(t, "latest", version)
}

func TestCDNSource_LatestVersion_Bad_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	src := NewCDNSource(SourceConfig{
		CDNURL:    server.URL,
		ImageName: "test.img",
	})

	version, err := src.LatestVersion(context.Background())
	assert.NoError(t, err) // Falls back to "latest"
	assert.Equal(t, "latest", version)
}

func TestCDNSource_Download_Good_NoProgress(t *testing.T) {
	content := "test content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, content)
	}))
	defer server.Close()

	dest := t.TempDir()
	src := NewCDNSource(SourceConfig{
		CDNURL:    server.URL,
		ImageName: "test.img",
	})

	// nil progress callback should be handled gracefully
	err := src.Download(context.Background(), dest, nil)
	assert.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dest, "test.img"))
	assert.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestCDNSource_Download_Good_LargeFile(t *testing.T) {
	// Create content larger than buffer size (32KB)
	content := make([]byte, 64*1024) // 64KB
	for i := range content {
		content[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	dest := t.TempDir()
	src := NewCDNSource(SourceConfig{
		CDNURL:    server.URL,
		ImageName: "large.img",
	})

	var progressCalls int
	var lastDownloaded int64
	err := src.Download(context.Background(), dest, func(downloaded, total int64) {
		progressCalls++
		lastDownloaded = downloaded
	})

	assert.NoError(t, err)
	assert.Greater(t, progressCalls, 1) // Should be called multiple times for large file
	assert.Equal(t, int64(len(content)), lastDownloaded)
}

func TestCDNSource_Download_Bad_HTTPErrorCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{"Bad Request", http.StatusBadRequest},
		{"Unauthorized", http.StatusUnauthorized},
		{"Forbidden", http.StatusForbidden},
		{"Not Found", http.StatusNotFound},
		{"Service Unavailable", http.StatusServiceUnavailable},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			dest := t.TempDir()
			src := NewCDNSource(SourceConfig{
				CDNURL:    server.URL,
				ImageName: "test.img",
			})

			err := src.Download(context.Background(), dest, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("HTTP %d", tc.statusCode))
		})
	}
}

func TestCDNSource_InterfaceCompliance(t *testing.T) {
	// Verify CDNSource implements ImageSource
	var _ ImageSource = (*CDNSource)(nil)
}

func TestCDNSource_Config(t *testing.T) {
	cfg := SourceConfig{
		CDNURL:    "https://cdn.example.com",
		ImageName: "my-image.qcow2",
	}
	src := NewCDNSource(cfg)

	assert.Equal(t, "https://cdn.example.com", src.config.CDNURL)
	assert.Equal(t, "my-image.qcow2", src.config.ImageName)
}

func TestNewCDNSource_Good(t *testing.T) {
	cfg := SourceConfig{
		GitHubRepo:    "host-uk/core-images",
		RegistryImage: "ghcr.io/host-uk/core-devops",
		CDNURL:        "https://cdn.example.com",
		ImageName:     "core-devops-darwin-arm64.qcow2",
	}

	src := NewCDNSource(cfg)
	assert.NotNil(t, src)
	assert.Equal(t, "cdn", src.Name())
	assert.Equal(t, cfg.CDNURL, src.config.CDNURL)
}

func TestCDNSource_Download_Good_CreatesDestDir(t *testing.T) {
	content := "test content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, content)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "nested", "dir")
	// dest doesn't exist yet

	src := NewCDNSource(SourceConfig{
		CDNURL:    server.URL,
		ImageName: "test.img",
	})

	err := src.Download(context.Background(), dest, nil)
	assert.NoError(t, err)

	// Verify nested dir was created
	info, err := os.Stat(dest)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestSourceConfig_Struct(t *testing.T) {
	cfg := SourceConfig{
		GitHubRepo:    "owner/repo",
		RegistryImage: "ghcr.io/owner/image",
		CDNURL:        "https://cdn.example.com",
		ImageName:     "image.qcow2",
	}

	assert.Equal(t, "owner/repo", cfg.GitHubRepo)
	assert.Equal(t, "ghcr.io/owner/image", cfg.RegistryImage)
	assert.Equal(t, "https://cdn.example.com", cfg.CDNURL)
	assert.Equal(t, "image.qcow2", cfg.ImageName)
}