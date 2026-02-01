package updater

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLatestUpdateFromURL(t *testing.T) {
	testCases := []struct {
		name            string
		handler         http.HandlerFunc
		expectError     bool
		expectedVersion string
		expectedURL     string
	}{
		{
			name: "Valid latest.json",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"version": "v1.1.0", "url": "http://example.com/release.zip"}`)
			},
			expectedVersion: "v1.1.0",
			expectedURL:     "http://example.com/release.zip",
		},
		{
			name: "Invalid JSON",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"version": "v1.1.0", "url": "http://example.com/release.zip"`) // Missing closing brace
			},
			expectError: true,
		},
		{
			name: "Missing version",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"url": "http://example.com/release.zip"}`)
			},
			expectError: true,
		},
		{
			name: "Missing URL",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"version": "v1.1.0"}`)
			},
			expectError: true,
		},
		{
			name: "Server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			info, err := GetLatestUpdateFromURL(server.URL)

			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}

			if !tc.expectError {
				if info.Version != tc.expectedVersion {
					t.Errorf("Expected version: %s, got: %s", tc.expectedVersion, info.Version)
				}
				if info.URL != tc.expectedURL {
					t.Errorf("Expected URL: %s, got: %s", tc.expectedURL, info.URL)
				}
			}
		})
	}
}
