package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"golang.org/x/oauth2"
)

// Repo represents a repository from the GitHub API.
type Repo struct {
	CloneURL string `json:"clone_url"` // The URL to clone the repository.
}

// ReleaseAsset represents a single asset from a GitHub release.
type ReleaseAsset struct {
	Name        string `json:"name"`                  // The name of the asset.
	DownloadURL string `json:"browser_download_url"` // The URL to download the asset.
}

// Release represents a GitHub release.
type Release struct {
	TagName    string         `json:"tag_name"`    // The name of the tag for the release.
	PreRelease bool           `json:"prerelease"`  // Indicates if the release is a pre-release.
	Assets     []ReleaseAsset `json:"assets"`      // A list of assets associated with the release.
}

// GithubClient defines the interface for interacting with the GitHub API.
// This allows for mocking the client in tests.
type GithubClient interface {
	// GetPublicRepos fetches the public repositories for a user or organization.
	GetPublicRepos(ctx context.Context, userOrOrg string) ([]string, error)
	// GetLatestRelease fetches the latest release for a given repository and channel.
	GetLatestRelease(ctx context.Context, owner, repo, channel string) (*Release, error)
	// GetReleaseByPullRequest fetches a release associated with a specific pull request number.
	GetReleaseByPullRequest(ctx context.Context, owner, repo string, prNumber int) (*Release, error)
}

type githubClient struct{}

// NewAuthenticatedClient creates a new HTTP client that authenticates with the GitHub API.
// It uses the GITHUB_TOKEN environment variable for authentication.
// If the token is not set, it returns the default HTTP client.
var NewAuthenticatedClient = func(ctx context.Context) *http.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return http.DefaultClient
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(ctx, ts)
}

func (g *githubClient) GetPublicRepos(ctx context.Context, userOrOrg string) ([]string, error) {
	return g.getPublicReposWithAPIURL(ctx, "https://api.github.com", userOrOrg)
}

func (g *githubClient) getPublicReposWithAPIURL(ctx context.Context, apiURL, userOrOrg string) ([]string, error) {
	client := NewAuthenticatedClient(ctx)
	var allCloneURLs []string
	url := fmt.Sprintf("%s/users/%s/repos", apiURL, userOrOrg)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Borg-Data-Collector")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			// Try organization endpoint
			url = fmt.Sprintf("%s/orgs/%s/repos", apiURL, userOrOrg)
			req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("User-Agent", "Borg-Data-Collector")
			resp, err = client.Do(req)
			if err != nil {
				return nil, err
			}
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to fetch repos: %s", resp.Status)
		}

		var repos []Repo
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		for _, repo := range repos {
			allCloneURLs = append(allCloneURLs, repo.CloneURL)
		}

		linkHeader := resp.Header.Get("Link")
		if linkHeader == "" {
			break
		}
		nextURL := g.findNextURL(linkHeader)
		if nextURL == "" {
			break
		}
		url = nextURL
	}

	return allCloneURLs, nil
}

func (g *githubClient) findNextURL(linkHeader string) string {
	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(link, ";")
		if len(parts) == 2 && strings.TrimSpace(parts[1]) == `rel="next"` {
			return strings.Trim(strings.TrimSpace(parts[0]), "<>")
		}
	}
	return ""
}

// GetLatestRelease fetches the latest release for a given repository and channel.
// The channel can be "stable", "beta", or "alpha".
func (g *githubClient) GetLatestRelease(ctx context.Context, owner, repo, channel string) (*Release, error) {
	client := NewAuthenticatedClient(ctx)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Borg-Data-Collector")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch releases: %s", resp.Status)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return filterReleases(releases, channel), nil
}

// filterReleases filters releases based on the specified channel.
func filterReleases(releases []Release, channel string) *Release {
	for _, release := range releases {
		releaseChannel := determineChannel(release.TagName, release.PreRelease)
		if releaseChannel == channel {
			return &release
		}
	}
	return nil
}

// determineChannel determines the stability channel of a release based on its tag and PreRelease flag.
func determineChannel(tagName string, isPreRelease bool) string {
	tagLower := strings.ToLower(tagName)
	if strings.Contains(tagLower, "alpha") {
		return "alpha"
	}
	if strings.Contains(tagLower, "beta") {
		return "beta"
	}
	if isPreRelease { // A pre-release without alpha/beta is treated as beta
		return "beta"
	}
	return "stable"
}

// GetReleaseByPullRequest fetches a release associated with a specific pull request number.
func (g *githubClient) GetReleaseByPullRequest(ctx context.Context, owner, repo string, prNumber int) (*Release, error) {
	client := NewAuthenticatedClient(ctx)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Borg-Data-Collector")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch releases: %s", resp.Status)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	// The pr number is included in the tag name with the format `vX.Y.Z-alpha.pr.123` or `vX.Y.Z-beta.pr.123`
	prTagSuffix := fmt.Sprintf(".pr.%d", prNumber)
	for _, release := range releases {
		if strings.Contains(release.TagName, prTagSuffix) {
			return &release, nil
		}
	}

	return nil, nil // No release found for the given PR number
}

// GetDownloadURL finds the appropriate download URL for the current operating system and architecture.
//
// It supports two modes of operation:
//  1. Using a 'releaseURLFormat' template: If 'releaseURLFormat' is provided,
//     it will be used to construct the download URL. The template can contain
//     placeholders for the release tag '{tag}', operating system '{os}', and
//     architecture '{arch}'.
//  2. Automatic detection: If 'releaseURLFormat' is empty, the function will
//     inspect the assets of the release to find a suitable download URL. It
//     searches for an asset name that contains both the current OS and architecture
//     (e.g., "my-app-linux-amd64"). If no match is found, it falls back to
//     matching only the OS.
//
// Example with releaseURLFormat:
//
//	release := &updater.Release{TagName: "v1.2.3"}
//	url, err := updater.GetDownloadURL(release, "https://example.com/downloads/{tag}/{os}/{arch}")
//	if err != nil {
//		// handle error
//	}
//	fmt.Println(url) // "https://example.com/downloads/v1.2.3/linux/amd64" (on a Linux AMD64 system)
//
// Example with automatic detection:
//
//	release := &updater.Release{
//		Assets: []updater.ReleaseAsset{
//			{Name: "my-app-linux-amd64", DownloadURL: "https://example.com/download/linux-amd64"},
//			{Name: "my-app-windows-amd64", DownloadURL: "https://example.com/download/windows-amd64"},
//		},
//	}
//	url, err := updater.GetDownloadURL(release, "")
//	if err != nil {
//		// handle error
//	}
//	fmt.Println(url) // "https://example.com/download/linux-amd64" (on a Linux AMD64 system)
func GetDownloadURL(release *Release, releaseURLFormat string) (string, error) {
	if release == nil {
		return "", fmt.Errorf("no release provided")
	}

	if releaseURLFormat != "" {
		// Replace {tag}, {os}, and {arch} placeholders
		r := strings.NewReplacer(
			"{tag}", release.TagName,
			"{os}", runtime.GOOS,
			"{arch}", runtime.GOARCH,
		)
		return r.Replace(releaseURLFormat), nil
	}

	osName := runtime.GOOS
	archName := runtime.GOARCH

	for _, asset := range release.Assets {
		assetNameLower := strings.ToLower(asset.Name)
		// Match asset that contains both OS and architecture
		if strings.Contains(assetNameLower, osName) && strings.Contains(assetNameLower, archName) {
			return asset.DownloadURL, nil
		}
	}

	// Fallback for OS only if no asset matched both OS and arch
	for _, asset := range release.Assets {
		assetNameLower := strings.ToLower(asset.Name)
		if strings.Contains(assetNameLower, osName) {
			return asset.DownloadURL, nil
		}
	}

	return "", fmt.Errorf("no suitable download asset found for %s/%s", osName, archName)
}
