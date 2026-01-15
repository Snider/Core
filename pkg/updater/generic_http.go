package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GenericUpdateInfo holds the information from a latest.json file.
// This file is expected to be at the root of a generic HTTP update server.
type GenericUpdateInfo struct {
	Version string `json:"version"` // The version number of the update.
	URL     string `json:"url"`     // The URL to download the update from.
}

// GetLatestUpdateFromURL fetches and parses a latest.json file from a base URL.
// The server at the baseURL should host a 'latest.json' file that contains
// the version and download URL for the latest update.
//
// Example of latest.json:
//
//	{
//	  "version": "1.2.3",
//	  "url": "https://your-server.com/path/to/release-asset"
//	}
func GetLatestUpdateFromURL(baseURL string) (*GenericUpdateInfo, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	// Append latest.json to the path
	u.Path += "/latest.json"

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest.json: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest.json: status code %d", resp.StatusCode)
	}

	var info GenericUpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to parse latest.json: %w", err)
	}

	if info.Version == "" || info.URL == "" {
		return nil, fmt.Errorf("invalid latest.json content: version or url is missing")
	}

	return &info, nil
}
