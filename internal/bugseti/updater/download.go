// Package updater provides auto-update functionality for BugSETI.
package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DownloadProgress reports download progress.
type DownloadProgress struct {
	BytesDownloaded int64   `json:"bytesDownloaded"`
	TotalBytes      int64   `json:"totalBytes"`
	Percent         float64 `json:"percent"`
}

// DownloadResult contains the result of a download operation.
type DownloadResult struct {
	BinaryPath string `json:"binaryPath"`
	Version    string `json:"version"`
	Checksum   string `json:"checksum"`
	VerifiedOK bool   `json:"verifiedOK"`
}

// Downloader handles downloading and verifying updates.
type Downloader struct {
	httpClient *http.Client
	stagingDir string
	onProgress func(DownloadProgress)
}

// NewDownloader creates a new update downloader.
func NewDownloader() (*Downloader, error) {
	// Create staging directory in user's temp dir
	stagingDir := filepath.Join(os.TempDir(), "bugseti-updates")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create staging directory: %w", err)
	}

	return &Downloader{
		httpClient: &http.Client{},
		stagingDir: stagingDir,
	}, nil
}

// SetProgressCallback sets a callback for download progress updates.
func (d *Downloader) SetProgressCallback(cb func(DownloadProgress)) {
	d.onProgress = cb
}

// Download downloads a release and stages it for installation.
func (d *Downloader) Download(ctx context.Context, release *ReleaseInfo) (*DownloadResult, error) {
	result := &DownloadResult{
		Version: release.Version,
	}

	// Prefer archive download for extraction
	downloadURL := release.ArchiveURL
	if downloadURL == "" {
		downloadURL = release.BinaryURL
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("no download URL available for release %s", release.Version)
	}

	// Download the checksum first if available
	var expectedChecksum string
	if release.ChecksumURL != "" {
		checksum, err := d.downloadChecksum(ctx, release.ChecksumURL)
		if err != nil {
			// Log but don't fail - checksum verification is optional
			fmt.Printf("Warning: could not download checksum: %v\n", err)
		} else {
			expectedChecksum = checksum
		}
	}

	// Download the file
	downloadedPath, err := d.downloadFile(ctx, downloadURL, release.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to download update: %w", err)
	}

	// Verify checksum if available
	actualChecksum, err := d.calculateChecksum(downloadedPath)
	if err != nil {
		os.Remove(downloadedPath)
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	result.Checksum = actualChecksum

	if expectedChecksum != "" {
		if actualChecksum != expectedChecksum {
			os.Remove(downloadedPath)
			return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
		}
		result.VerifiedOK = true
	}

	// Extract if it's an archive
	var binaryPath string
	if strings.HasSuffix(downloadURL, ".tar.gz") {
		binaryPath, err = d.extractTarGz(downloadedPath)
	} else if strings.HasSuffix(downloadURL, ".zip") {
		binaryPath, err = d.extractZip(downloadedPath)
	} else {
		// It's a raw binary
		binaryPath = downloadedPath
	}

	if err != nil {
		os.Remove(downloadedPath)
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	// Make the binary executable (Unix only)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	result.BinaryPath = binaryPath
	return result, nil
}

// downloadChecksum downloads and parses a checksum file.
func (d *Downloader) downloadChecksum(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "BugSETI-Updater")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Checksum file format: "hash  filename" or just "hash"
	parts := strings.Fields(strings.TrimSpace(string(data)))
	if len(parts) == 0 {
		return "", fmt.Errorf("empty checksum file")
	}

	return parts[0], nil
}

// downloadFile downloads a file with progress reporting.
func (d *Downloader) downloadFile(ctx context.Context, url string, expectedSize int64) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "BugSETI-Updater")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Get total size from response or use expected size
	totalSize := resp.ContentLength
	if totalSize <= 0 {
		totalSize = expectedSize
	}

	// Create output file
	filename := filepath.Base(url)
	outPath := filepath.Join(d.stagingDir, filename)
	out, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Download with progress
	var downloaded int64
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		select {
		case <-ctx.Done():
			os.Remove(outPath)
			return "", ctx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				os.Remove(outPath)
				return "", writeErr
			}
			downloaded += int64(n)

			// Report progress
			if d.onProgress != nil && totalSize > 0 {
				d.onProgress(DownloadProgress{
					BytesDownloaded: downloaded,
					TotalBytes:      totalSize,
					Percent:         float64(downloaded) / float64(totalSize) * 100,
				})
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			os.Remove(outPath)
			return "", readErr
		}
	}

	return outPath, nil
}

// calculateChecksum calculates the SHA256 checksum of a file.
func (d *Downloader) calculateChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// extractTarGz extracts a .tar.gz archive and returns the path to the binary.
func (d *Downloader) extractTarGz(archivePath string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	extractDir := filepath.Join(d.stagingDir, "extracted")
	os.RemoveAll(extractDir)
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return "", err
	}

	var binaryPath string
	binaryName := "bugseti"
	if runtime.GOOS == "windows" {
		binaryName = "bugseti.exe"
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		target := filepath.Join(extractDir, header.Name)

		// Prevent directory traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(extractDir)) {
			return "", fmt.Errorf("invalid file path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			// Create parent directory
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", err
			}

			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", err
			}
			outFile.Close()

			// Check if this is the binary we're looking for
			if filepath.Base(header.Name) == binaryName {
				binaryPath = target
			}
		}
	}

	// Clean up archive
	os.Remove(archivePath)

	if binaryPath == "" {
		return "", fmt.Errorf("binary not found in archive")
	}

	return binaryPath, nil
}

// extractZip extracts a .zip archive and returns the path to the binary.
func (d *Downloader) extractZip(archivePath string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	extractDir := filepath.Join(d.stagingDir, "extracted")
	os.RemoveAll(extractDir)
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return "", err
	}

	var binaryPath string
	binaryName := "bugseti"
	if runtime.GOOS == "windows" {
		binaryName = "bugseti.exe"
	}

	for _, f := range r.File {
		target := filepath.Join(extractDir, f.Name)

		// Prevent directory traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(extractDir)) {
			return "", fmt.Errorf("invalid file path in archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			return "", err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return "", err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return "", err
		}

		// Check if this is the binary we're looking for
		if filepath.Base(f.Name) == binaryName {
			binaryPath = target
		}
	}

	// Clean up archive
	os.Remove(archivePath)

	if binaryPath == "" {
		return "", fmt.Errorf("binary not found in archive")
	}

	return binaryPath, nil
}

// Cleanup removes all staged files.
func (d *Downloader) Cleanup() error {
	return os.RemoveAll(d.stagingDir)
}

// GetStagingDir returns the staging directory path.
func (d *Downloader) GetStagingDir() string {
	return d.stagingDir
}
