package updater

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewUpdateService(t *testing.T) {
	testCases := []struct {
		name        string
		config      UpdateServiceConfig
		expectError bool
		isGitHub    bool
	}{
		{
			name: "Valid GitHub URL",
			config: UpdateServiceConfig{
				RepoURL: "https://github.com/owner/repo",
			},
			isGitHub: true,
		},
		{
			name: "Valid non-GitHub URL",
			config: UpdateServiceConfig{
				RepoURL: "https://example.com/updates",
			},
			isGitHub: false,
		},
		{
			name: "Invalid GitHub URL",
			config: UpdateServiceConfig{
				RepoURL: "https://github.com/owner",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, err := NewUpdateService(tc.config)
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}
			if err == nil && service.isGitHub != tc.isGitHub {
				t.Errorf("Expected isGitHub: %v, got: %v", tc.isGitHub, service.isGitHub)
			}
		})
	}
}

func TestUpdateService_Start(t *testing.T) {
	// Setup a mock server for HTTP tests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"version": "v1.1.0", "url": "http://example.com/release.zip"}`))
	}))
	defer server.Close()

	testCases := []struct {
		name                string
		config              UpdateServiceConfig
		checkOnlyGitHub     int
		checkAndDoGitHub    int
		checkOnlyHTTPCalls  int
		checkAndDoHTTPCalls int
		expectError         bool
	}{
		{
			name: "GitHub: NoCheck",
			config: UpdateServiceConfig{
				RepoURL:        "https://github.com/owner/repo",
				CheckOnStartup: NoCheck,
			},
		},
		{
			name: "GitHub: CheckOnStartup",
			config: UpdateServiceConfig{
				RepoURL:        "https://github.com/owner/repo",
				CheckOnStartup: CheckOnStartup,
			},
			checkOnlyGitHub: 1,
		},
		{
			name: "GitHub: CheckAndUpdateOnStartup",
			config: UpdateServiceConfig{
				RepoURL:        "https://github.com/owner/repo",
				CheckOnStartup: CheckAndUpdateOnStartup,
			},
			checkAndDoGitHub: 1,
		},
		{
			name: "HTTP: NoCheck",
			config: UpdateServiceConfig{
				RepoURL:        server.URL,
				CheckOnStartup: NoCheck,
			},
		},
		{
			name: "HTTP: CheckOnStartup",
			config: UpdateServiceConfig{
				RepoURL:        server.URL,
				CheckOnStartup: CheckOnStartup,
			},
			checkOnlyHTTPCalls: 1,
		},
		{
			name: "HTTP: CheckAndUpdateOnStartup",
			config: UpdateServiceConfig{
				RepoURL:        server.URL,
				CheckOnStartup: CheckAndUpdateOnStartup,
			},
			checkAndDoHTTPCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var checkOnlyGitHub, checkAndDoGitHub, checkOnlyHTTP, checkAndDoHTTP int

			// Mock GitHub functions
			originalCheckOnly := CheckOnly
			CheckOnly = func(owner, repo, channel string, forceSemVerPrefix bool, releaseURLFormat string) error {
				checkOnlyGitHub++
				return nil
			}
			defer func() { CheckOnly = originalCheckOnly }()

			originalCheckForUpdates := CheckForUpdates
			CheckForUpdates = func(owner, repo, channel string, forceSemVerPrefix bool, releaseURLFormat string) error {
				checkAndDoGitHub++
				return nil
			}
			defer func() { CheckForUpdates = originalCheckForUpdates }()

			// Mock HTTP functions
			originalCheckOnlyHTTP := CheckOnlyHTTP
			CheckOnlyHTTP = func(baseURL string) error {
				checkOnlyHTTP++
				return nil
			}
			defer func() { CheckOnlyHTTP = originalCheckOnlyHTTP }()

			originalCheckForUpdatesHTTP := CheckForUpdatesHTTP
			CheckForUpdatesHTTP = func(baseURL string) error {
				checkAndDoHTTP++
				return nil
			}
			defer func() { CheckForUpdatesHTTP = originalCheckForUpdatesHTTP }()

			service, _ := NewUpdateService(tc.config)
			err := service.Start()

			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}
			if checkOnlyGitHub != tc.checkOnlyGitHub {
				t.Errorf("Expected GitHub CheckOnly calls: %d, got: %d", tc.checkOnlyGitHub, checkOnlyGitHub)
			}
			if checkAndDoGitHub != tc.checkAndDoGitHub {
				t.Errorf("Expected GitHub CheckForUpdates calls: %d, got: %d", tc.checkAndDoGitHub, checkAndDoGitHub)
			}
			if checkOnlyHTTP != tc.checkOnlyHTTPCalls {
				t.Errorf("Expected HTTP CheckOnly calls: %d, got: %d", tc.checkOnlyHTTPCalls, checkOnlyHTTP)
			}
			if checkAndDoHTTP != tc.checkAndDoHTTPCalls {
				t.Errorf("Expected HTTP CheckForUpdates calls: %d, got: %d", tc.checkAndDoHTTPCalls, checkAndDoHTTP)
			}
		})
	}
}
