package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/minio/selfupdate"
	"golang.org/x/mod/semver"
)

// Version holds the current version of the application.
// It is set at build time via ldflags or fallback to the version in package.json.
var Version = PkgVersion

// NewGithubClient is a variable that holds a function to create a new GithubClient.
// This can be replaced in tests to inject a mock client.
//
// Example:
//
//	updater.NewGithubClient = func() updater.GithubClient {
//		return &mockClient{} // or your mock implementation
//	}
var NewGithubClient = func() GithubClient {
	return &githubClient{}
}

// DoUpdate is a variable that holds the function to perform the actual update.
// This can be replaced in tests to prevent actual updates.
var DoUpdate = func(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}(resp.Body)

	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("failed to rollback from failed update: %v", rerr)
		}
		return fmt.Errorf("update failed: %v", err)
	}

	fmt.Println("Update applied successfully.")
	return nil
}

// CheckForNewerVersion checks if a newer version of the application is available on GitHub.
// It fetches the latest release for the given owner, repository, and channel, and compares its tag
// with the current application version.
var CheckForNewerVersion = func(owner, repo, channel string, forceSemVerPrefix bool) (*Release, bool, error) {
	client := NewGithubClient()
	ctx := context.Background()

	release, err := client.GetLatestRelease(ctx, owner, repo, channel)
	if err != nil {
		return nil, false, fmt.Errorf("error fetching latest release: %w", err)
	}

	if release == nil {
		return nil, false, nil // No release found
	}

	// Always normalize to 'v' prefix for semver comparison
	vCurrent := formatVersionForComparison(Version)
	vLatest := formatVersionForComparison(release.TagName)

	if semver.Compare(vCurrent, vLatest) >= 0 {
		return release, false, nil // Current version is up-to-date or newer
	}

	return release, true, nil // A newer version is available
}

// CheckForUpdates checks for new updates on GitHub and applies them if a newer version is found.
// It uses the provided owner, repository, and channel to find the latest release.
var CheckForUpdates = func(owner, repo, channel string, forceSemVerPrefix bool, releaseURLFormat string) error {
	release, updateAvailable, err := CheckForNewerVersion(owner, repo, channel, forceSemVerPrefix)
	if err != nil {
		return err
	}

	if !updateAvailable {
		if release != nil {
			fmt.Printf("Current version %s is up-to-date with latest release %s.\n",
				formatVersionForDisplay(Version, forceSemVerPrefix),
				formatVersionForDisplay(release.TagName, forceSemVerPrefix))
		} else {
			fmt.Println("No releases found.")
		}
		return nil
	}

	fmt.Printf("Newer version %s found (current: %s). Applying update...\n",
		formatVersionForDisplay(release.TagName, forceSemVerPrefix),
		formatVersionForDisplay(Version, forceSemVerPrefix))

	downloadURL, err := GetDownloadURL(release, releaseURLFormat)
	if err != nil {
		return fmt.Errorf("error getting download URL: %w", err)
	}

	return DoUpdate(downloadURL)
}

// CheckOnly checks for new updates on GitHub without applying them.
// It prints a message indicating if a new release is available.
var CheckOnly = func(owner, repo, channel string, forceSemVerPrefix bool, releaseURLFormat string) error {
	release, updateAvailable, err := CheckForNewerVersion(owner, repo, channel, forceSemVerPrefix)
	if err != nil {
		return err
	}

	if !updateAvailable {
		if release != nil {
			fmt.Printf("Current version %s is up-to-date with latest release %s.\n",
				formatVersionForDisplay(Version, forceSemVerPrefix),
				formatVersionForDisplay(release.TagName, forceSemVerPrefix))
		} else {
			fmt.Println("No new release found.")
		}
		return nil
	}

	fmt.Printf("New release found: %s (current version: %s)\n",
		formatVersionForDisplay(release.TagName, forceSemVerPrefix),
		formatVersionForDisplay(Version, forceSemVerPrefix))
	return nil
}

// CheckForUpdatesByTag checks for and applies updates from GitHub based on the channel
// determined by the current application's version tag (e.g., 'stable' or 'prerelease').
var CheckForUpdatesByTag = func(owner, repo string) error {
	channel := determineChannel(Version, false) // isPreRelease is false for current version
	return CheckForUpdates(owner, repo, channel, true, "")
}

// CheckOnlyByTag checks for updates from GitHub based on the channel determined by the
// current version tag, without applying them.
var CheckOnlyByTag = func(owner, repo string) error {
	channel := determineChannel(Version, false) // isPreRelease is false for current version
	return CheckOnly(owner, repo, channel, true, "")
}

// CheckForUpdatesByPullRequest finds a release associated with a specific pull request number
// on GitHub and applies the update.
var CheckForUpdatesByPullRequest = func(owner, repo string, prNumber int, releaseURLFormat string) error {
	client := NewGithubClient()
	ctx := context.Background()

	release, err := client.GetReleaseByPullRequest(ctx, owner, repo, prNumber)
	if err != nil {
		return fmt.Errorf("error fetching release for pull request: %w", err)
	}

	if release == nil {
		fmt.Printf("No release found for PR #%d.\n", prNumber)
		return nil
	}

	fmt.Printf("Release %s found for PR #%d. Applying update...\n", release.TagName, prNumber)

	downloadURL, err := GetDownloadURL(release, releaseURLFormat)
	if err != nil {
		return fmt.Errorf("error getting download URL: %w", err)
	}

	return DoUpdate(downloadURL)
}

// CheckForUpdatesHTTP checks for and applies updates from a generic HTTP endpoint.
// The endpoint is expected to provide update information in a structured format.
var CheckForUpdatesHTTP = func(baseURL string) error {
	info, err := GetLatestUpdateFromURL(baseURL)
	if err != nil {
		return err
	}

	vCurrent := formatVersionForComparison(Version)
	vLatest := formatVersionForComparison(info.Version)

	if semver.Compare(vCurrent, vLatest) >= 0 {
		fmt.Printf("Current version %s is up-to-date with latest release %s.\n", Version, info.Version)
		return nil
	}

	fmt.Printf("Newer version %s found (current: %s). Applying update...\n", info.Version, Version)
	return DoUpdate(info.URL)
}

// CheckOnlyHTTP checks for updates from a generic HTTP endpoint without applying them.
// It prints a message if a new version is available.
var CheckOnlyHTTP = func(baseURL string) error {
	info, err := GetLatestUpdateFromURL(baseURL)
	if err != nil {
		return err
	}

	vCurrent := formatVersionForComparison(Version)
	vLatest := formatVersionForComparison(info.Version)

	if semver.Compare(vCurrent, vLatest) >= 0 {
		fmt.Printf("Current version %s is up-to-date with latest release %s.\n", Version, info.Version)
		return nil
	}

	fmt.Printf("New release found: %s (current version: %s)\n", info.Version, Version)
	return nil
}

// formatVersionForComparison ensures the version string has a 'v' prefix for semver comparison.
func formatVersionForComparison(version string) string {
	if version != "" && !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}

// formatVersionForDisplay ensures the version string has the correct 'v' prefix based on the forceSemVerPrefix flag.
func formatVersionForDisplay(version string, forceSemVerPrefix bool) string {
	hasV := strings.HasPrefix(version, "v")
	if forceSemVerPrefix && !hasV {
		return "v" + version
	}
	if !forceSemVerPrefix && hasV {
		return strings.TrimPrefix(version, "v")
	}
	return version
}
