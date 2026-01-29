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

// Compile-time interface check.
var _ ImageSource = (*CDNSource)(nil)

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
	if err := os.MkdirAll(dest, 0755); err != nil {
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
