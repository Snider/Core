// Package updater provides auto-update functionality for BugSETI.
package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	// GitHubReleasesAPI is the GitHub API endpoint for releases.
	GitHubReleasesAPI = "https://api.github.com/repos/%s/%s/releases"

	// DefaultOwner is the default GitHub repository owner.
	DefaultOwner = "host-uk"

	// DefaultRepo is the default GitHub repository name.
	DefaultRepo = "core"

	// DefaultCheckInterval is the default interval between update checks.
	DefaultCheckInterval = 6 * time.Hour
)

// GitHubRelease represents a GitHub release from the API.
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	PublishedAt time.Time     `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
	HTMLURL     string        `json:"html_url"`
}

// GitHubAsset represents a release asset from the GitHub API.
type GitHubAsset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
}

// ReleaseInfo contains information about an available release.
type ReleaseInfo struct {
	Version     string    `json:"version"`
	Channel     Channel   `json:"channel"`
	Tag         string    `json:"tag"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"publishedAt"`
	HTMLURL     string    `json:"htmlUrl"`
	BinaryURL   string    `json:"binaryUrl"`
	ArchiveURL  string    `json:"archiveUrl"`
	ChecksumURL string    `json:"checksumUrl"`
	Size        int64     `json:"size"`
}

// UpdateCheckResult contains the result of an update check.
type UpdateCheckResult struct {
	Available      bool         `json:"available"`
	CurrentVersion string       `json:"currentVersion"`
	LatestVersion  string       `json:"latestVersion"`
	Release        *ReleaseInfo `json:"release,omitempty"`
	Error          string       `json:"error,omitempty"`
	CheckedAt      time.Time    `json:"checkedAt"`
}

// Checker checks for available updates.
type Checker struct {
	owner      string
	repo       string
	httpClient *http.Client
}

// NewChecker creates a new update checker.
func NewChecker() *Checker {
	return &Checker{
		owner: DefaultOwner,
		repo:  DefaultRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckForUpdate checks if a newer version is available.
func (c *Checker) CheckForUpdate(ctx context.Context, currentVersion string, channel Channel) (*UpdateCheckResult, error) {
	result := &UpdateCheckResult{
		CurrentVersion: currentVersion,
		CheckedAt:      time.Now(),
	}

	// Fetch releases from GitHub
	releases, err := c.fetchReleases(ctx)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Find the latest release for the channel
	latest := c.findLatestRelease(releases, channel)
	if latest == nil {
		result.LatestVersion = currentVersion
		return result, nil
	}

	result.LatestVersion = latest.Version
	result.Release = latest

	// Compare versions
	if c.isNewerVersion(currentVersion, latest.Version, channel) {
		result.Available = true
	}

	return result, nil
}

// fetchReleases fetches all releases from GitHub.
func (c *Checker) fetchReleases(ctx context.Context) ([]GitHubRelease, error) {
	url := fmt.Sprintf(GitHubReleasesAPI, c.owner, c.repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "BugSETI-Updater")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}

	return releases, nil
}

// findLatestRelease finds the latest release for the given channel.
func (c *Checker) findLatestRelease(releases []GitHubRelease, channel Channel) *ReleaseInfo {
	var candidates []ReleaseInfo

	for _, release := range releases {
		// Skip drafts
		if release.Draft {
			continue
		}

		// Check if the tag matches our BugSETI release pattern
		if !strings.HasPrefix(release.TagName, "bugseti-") {
			continue
		}

		// Determine the channel for this release
		releaseChannel := c.determineChannel(release.TagName)
		if releaseChannel == "" {
			continue
		}

		// Check if this release should be considered for the requested channel
		if !channel.IncludesChannel(releaseChannel) {
			continue
		}

		// Extract version
		version := releaseChannel.ExtractVersion(release.TagName)
		if version == "" {
			continue
		}

		// Find the appropriate asset for this platform
		binaryName := c.getBinaryName()
		archiveName := c.getArchiveName()
		checksumName := archiveName + ".sha256"

		var binaryURL, archiveURL, checksumURL string
		var size int64

		for _, asset := range release.Assets {
			switch asset.Name {
			case binaryName:
				binaryURL = asset.BrowserDownloadURL
				size = asset.Size
			case archiveName:
				archiveURL = asset.BrowserDownloadURL
				if size == 0 {
					size = asset.Size
				}
			case checksumName:
				checksumURL = asset.BrowserDownloadURL
			}
		}

		// Skip if no binary available for this platform
		if binaryURL == "" && archiveURL == "" {
			continue
		}

		candidates = append(candidates, ReleaseInfo{
			Version:     version,
			Channel:     releaseChannel,
			Tag:         release.TagName,
			Name:        release.Name,
			Body:        release.Body,
			PublishedAt: release.PublishedAt,
			HTMLURL:     release.HTMLURL,
			BinaryURL:   binaryURL,
			ArchiveURL:  archiveURL,
			ChecksumURL: checksumURL,
			Size:        size,
		})
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort by version (newest first)
	sort.Slice(candidates, func(i, j int) bool {
		return c.compareVersions(candidates[i].Version, candidates[j].Version, channel) > 0
	})

	return &candidates[0]
}

// determineChannel determines the channel from a release tag.
func (c *Checker) determineChannel(tag string) Channel {
	for _, ch := range AllChannels() {
		if ch.MatchesTag(tag) {
			return ch
		}
	}
	return ""
}

// getBinaryName returns the binary name for the current platform.
func (c *Checker) getBinaryName() string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("bugseti-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext)
}

// getArchiveName returns the archive name for the current platform.
func (c *Checker) getArchiveName() string {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("bugseti-%s-%s.%s", runtime.GOOS, runtime.GOARCH, ext)
}

// isNewerVersion returns true if newVersion is newer than currentVersion.
func (c *Checker) isNewerVersion(currentVersion, newVersion string, channel Channel) bool {
	// Handle nightly versions (date-based)
	if channel == ChannelNightly {
		return newVersion > currentVersion
	}

	// Handle dev builds
	if currentVersion == "dev" {
		return true
	}

	// Use semver comparison
	current := c.normalizeSemver(currentVersion)
	new := c.normalizeSemver(newVersion)

	return semver.Compare(new, current) > 0
}

// compareVersions compares two versions.
func (c *Checker) compareVersions(v1, v2 string, channel Channel) int {
	// Handle nightly versions (date-based)
	if channel == ChannelNightly {
		if v1 > v2 {
			return 1
		} else if v1 < v2 {
			return -1
		}
		return 0
	}

	// Use semver comparison
	return semver.Compare(c.normalizeSemver(v1), c.normalizeSemver(v2))
}

// normalizeSemver ensures a version string has the 'v' prefix for semver.
func (c *Checker) normalizeSemver(version string) string {
	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}

// GetAllReleases returns all BugSETI releases from GitHub.
func (c *Checker) GetAllReleases(ctx context.Context) ([]ReleaseInfo, error) {
	releases, err := c.fetchReleases(ctx)
	if err != nil {
		return nil, err
	}

	var result []ReleaseInfo
	for _, release := range releases {
		if release.Draft {
			continue
		}

		if !strings.HasPrefix(release.TagName, "bugseti-") {
			continue
		}

		releaseChannel := c.determineChannel(release.TagName)
		if releaseChannel == "" {
			continue
		}

		version := releaseChannel.ExtractVersion(release.TagName)
		if version == "" {
			continue
		}

		binaryName := c.getBinaryName()
		archiveName := c.getArchiveName()
		checksumName := archiveName + ".sha256"

		var binaryURL, archiveURL, checksumURL string
		var size int64

		for _, asset := range release.Assets {
			switch asset.Name {
			case binaryName:
				binaryURL = asset.BrowserDownloadURL
				size = asset.Size
			case archiveName:
				archiveURL = asset.BrowserDownloadURL
				if size == 0 {
					size = asset.Size
				}
			case checksumName:
				checksumURL = asset.BrowserDownloadURL
			}
		}

		result = append(result, ReleaseInfo{
			Version:     version,
			Channel:     releaseChannel,
			Tag:         release.TagName,
			Name:        release.Name,
			Body:        release.Body,
			PublishedAt: release.PublishedAt,
			HTMLURL:     release.HTMLURL,
			BinaryURL:   binaryURL,
			ArchiveURL:  archiveURL,
			ChecksumURL: checksumURL,
			Size:        size,
		})
	}

	return result, nil
}
