package updater

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
)

// mockGithubClient is a mock implementation of the GithubClient interface for testing.
type mockGithubClient struct {
	getLatestRelease      func(ctx context.Context, owner, repo, channel string) (*Release, error)
	getReleaseByPR        func(ctx context.Context, owner, repo string, prNumber int) (*Release, error)
	getPublicRepos        func(ctx context.Context, userOrOrg string) ([]string, error)
	getLatestReleaseCount int
	getReleaseByPRCount   int
	getPublicReposCount   int
}

func (m *mockGithubClient) GetLatestRelease(ctx context.Context, owner, repo, channel string) (*Release, error) {
	m.getLatestReleaseCount++
	return m.getLatestRelease(ctx, owner, repo, channel)
}

func (m *mockGithubClient) GetReleaseByPullRequest(ctx context.Context, owner, repo string, prNumber int) (*Release, error) {
	m.getReleaseByPRCount++
	return m.getReleaseByPR(ctx, owner, repo, prNumber)
}

func (m *mockGithubClient) GetPublicRepos(ctx context.Context, userOrOrg string) ([]string, error) {
	m.getPublicReposCount++
	if m.getPublicRepos != nil {
		return m.getPublicRepos(ctx, userOrOrg)
	}
	return nil, fmt.Errorf("GetPublicRepos not implemented")
}

func ExampleCheckForNewerVersion() {
	originalNewGithubClient := NewGithubClient
	defer func() { NewGithubClient = originalNewGithubClient }()

	NewGithubClient = func() GithubClient {
		return &mockGithubClient{
			getLatestRelease: func(ctx context.Context, owner, repo, channel string) (*Release, error) {
				return &Release{TagName: "v1.1.0"}, nil
			},
		}
	}

	Version = "1.0.0"
	release, available, err := CheckForNewerVersion("owner", "repo", "stable", true)
	if err != nil {
		log.Fatalf("CheckForNewerVersion failed: %v", err)
	}

	if available {
		fmt.Printf("Newer version available: %s", release.TagName)
	} else {
		fmt.Println("No newer version available.")
	}
	// Output: Newer version available: v1.1.0
}

func ExampleCheckForUpdates() {
	// Mock the functions to prevent actual updates and network calls
	originalDoUpdate := DoUpdate
	originalNewGithubClient := NewGithubClient
	defer func() {
		DoUpdate = originalDoUpdate
		NewGithubClient = originalNewGithubClient
	}()

	NewGithubClient = func() GithubClient {
		return &mockGithubClient{
			getLatestRelease: func(ctx context.Context, owner, repo, channel string) (*Release, error) {
				return &Release{
					TagName: "v1.1.0",
					Assets:  []ReleaseAsset{{Name: fmt.Sprintf("test-asset-%s-%s", "linux", "amd64"), DownloadURL: "http://example.com/asset"}},
				}, nil
			},
		}
	}

	DoUpdate = func(url string) error {
		fmt.Printf("Update would be applied from: %s", url)
		return nil
	}

	Version = "1.0.0"
	err := CheckForUpdates("owner", "repo", "stable", true, "")
	if err != nil {
		log.Fatalf("CheckForUpdates failed: %v", err)
	}
	// Output:
	// Newer version v1.1.0 found (current: v1.0.0). Applying update...
	// Update would be applied from: http://example.com/asset
}

func ExampleCheckOnly() {
	originalNewGithubClient := NewGithubClient
	defer func() { NewGithubClient = originalNewGithubClient }()

	NewGithubClient = func() GithubClient {
		return &mockGithubClient{
			getLatestRelease: func(ctx context.Context, owner, repo, channel string) (*Release, error) {
				return &Release{TagName: "v1.1.0"}, nil
			},
		}
	}

	Version = "1.0.0"
	err := CheckOnly("owner", "repo", "stable", true, "")
	if err != nil {
		log.Fatalf("CheckOnly failed: %v", err)
	}
	// Output: New release found: v1.1.0 (current version: v1.0.0)
}

func ExampleCheckForUpdatesByTag() {
	// Mock the functions to prevent actual updates and network calls
	originalDoUpdate := DoUpdate
	originalNewGithubClient := NewGithubClient
	defer func() {
		DoUpdate = originalDoUpdate
		NewGithubClient = originalNewGithubClient
	}()

	NewGithubClient = func() GithubClient {
		return &mockGithubClient{
			getLatestRelease: func(ctx context.Context, owner, repo, channel string) (*Release, error) {
				if channel == "stable" {
					return &Release{
						TagName: "v1.1.0",
						Assets:  []ReleaseAsset{{Name: fmt.Sprintf("test-asset-%s-%s", "linux", "amd64"), DownloadURL: "http://example.com/asset"}},
					}, nil
				}
				return nil, nil
			},
		}
	}

	DoUpdate = func(url string) error {
		fmt.Printf("Update would be applied from: %s", url)
		return nil
	}

	Version = "1.0.0" // A version that resolves to the "stable" channel
	err := CheckForUpdatesByTag("owner", "repo")
	if err != nil {
		log.Fatalf("CheckForUpdatesByTag failed: %v", err)
	}
	// Output:
	// Newer version v1.1.0 found (current: v1.0.0). Applying update...
	// Update would be applied from: http://example.com/asset
}

func ExampleCheckOnlyByTag() {
	originalNewGithubClient := NewGithubClient
	defer func() { NewGithubClient = originalNewGithubClient }()

	NewGithubClient = func() GithubClient {
		return &mockGithubClient{
			getLatestRelease: func(ctx context.Context, owner, repo, channel string) (*Release, error) {
				if channel == "stable" {
					return &Release{TagName: "v1.1.0"}, nil
				}
				return nil, nil
			},
		}
	}

	Version = "1.0.0" // A version that resolves to the "stable" channel
	err := CheckOnlyByTag("owner", "repo")
	if err != nil {
		log.Fatalf("CheckOnlyByTag failed: %v", err)
	}
	// Output: New release found: v1.1.0 (current version: v1.0.0)
}

func ExampleCheckForUpdatesByPullRequest() {
	// Mock the functions to prevent actual updates and network calls
	originalDoUpdate := DoUpdate
	originalNewGithubClient := NewGithubClient
	defer func() {
		DoUpdate = originalDoUpdate
		NewGithubClient = originalNewGithubClient
	}()

	NewGithubClient = func() GithubClient {
		return &mockGithubClient{
			getReleaseByPR: func(ctx context.Context, owner, repo string, prNumber int) (*Release, error) {
				if prNumber == 123 {
					return &Release{
						TagName: "v1.1.0-alpha.pr.123",
						Assets:  []ReleaseAsset{{Name: fmt.Sprintf("test-asset-%s-%s", "linux", "amd64"), DownloadURL: "http://example.com/asset-pr"}},
					}, nil
				}
				return nil, nil
			},
		}
	}

	DoUpdate = func(url string) error {
		fmt.Printf("Update would be applied from: %s", url)
		return nil
	}

	err := CheckForUpdatesByPullRequest("owner", "repo", 123, "")
	if err != nil {
		log.Fatalf("CheckForUpdatesByPullRequest failed: %v", err)
	}
	// Output:
	// Release v1.1.0-alpha.pr.123 found for PR #123. Applying update...
	// Update would be applied from: http://example.com/asset-pr
}

func ExampleCheckForUpdatesHTTP() {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest.json" {
			fmt.Fprintln(w, `{"version": "1.1.0", "url": "http://example.com/update"}`)
		}
	}))
	defer server.Close()

	// Mock the doUpdateFunc to prevent actual updates
	originalDoUpdate := DoUpdate
	defer func() { DoUpdate = originalDoUpdate }()
	DoUpdate = func(url string) error {
		fmt.Printf("Update would be applied from: %s", url)
		return nil
	}

	Version = "1.0.0"
	err := CheckForUpdatesHTTP(server.URL)
	if err != nil {
		log.Fatalf("CheckForUpdatesHTTP failed: %v", err)
	}
	// Output:
	// Newer version 1.1.0 found (current: 1.0.0). Applying update...
	// Update would be applied from: http://example.com/update
}

func ExampleCheckOnlyHTTP() {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest.json" {
			fmt.Fprintln(w, `{"version": "1.1.0", "url": "http://example.com/update"}`)
		}
	}))
	defer server.Close()

	Version = "1.0.0"
	err := CheckOnlyHTTP(server.URL)
	if err != nil {
		log.Fatalf("CheckOnlyHTTP failed: %v", err)
	}
	// Output: New release found: 1.1.0 (current version: 1.0.0)
}
