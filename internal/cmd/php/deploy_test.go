package php

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCoolifyConfig_Good(t *testing.T) {
	tests := []struct {
		name        string
		envContent  string
		wantURL     string
		wantToken   string
		wantAppID   string
		wantStaging string
	}{
		{
			name: "all values set",
			envContent: `COOLIFY_URL=https://coolify.example.com
COOLIFY_TOKEN=secret-token
COOLIFY_APP_ID=app-123
COOLIFY_STAGING_APP_ID=staging-456`,
			wantURL:     "https://coolify.example.com",
			wantToken:   "secret-token",
			wantAppID:   "app-123",
			wantStaging: "staging-456",
		},
		{
			name: "quoted values",
			envContent: `COOLIFY_URL="https://coolify.example.com"
COOLIFY_TOKEN='secret-token'
COOLIFY_APP_ID="app-123"`,
			wantURL:   "https://coolify.example.com",
			wantToken: "secret-token",
			wantAppID: "app-123",
		},
		{
			name: "with comments and blank lines",
			envContent: `# Coolify configuration
COOLIFY_URL=https://coolify.example.com

# API token
COOLIFY_TOKEN=secret-token
COOLIFY_APP_ID=app-123
# COOLIFY_STAGING_APP_ID=not-this`,
			wantURL:   "https://coolify.example.com",
			wantToken: "secret-token",
			wantAppID: "app-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir := t.TempDir()
			envPath := filepath.Join(dir, ".env")

			// Write .env file
			if err := os.WriteFile(envPath, []byte(tt.envContent), 0644); err != nil {
				t.Fatalf("failed to write .env: %v", err)
			}

			// Load config
			config, err := LoadCoolifyConfig(dir)
			if err != nil {
				t.Fatalf("LoadCoolifyConfig() error = %v", err)
			}

			if config.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", config.URL, tt.wantURL)
			}
			if config.Token != tt.wantToken {
				t.Errorf("Token = %q, want %q", config.Token, tt.wantToken)
			}
			if config.AppID != tt.wantAppID {
				t.Errorf("AppID = %q, want %q", config.AppID, tt.wantAppID)
			}
			if tt.wantStaging != "" && config.StagingAppID != tt.wantStaging {
				t.Errorf("StagingAppID = %q, want %q", config.StagingAppID, tt.wantStaging)
			}
		})
	}
}

func TestLoadCoolifyConfig_Bad(t *testing.T) {
	tests := []struct {
		name       string
		envContent string
		wantErr    string
	}{
		{
			name:       "missing URL",
			envContent: "COOLIFY_TOKEN=secret",
			wantErr:    "COOLIFY_URL is not set",
		},
		{
			name:       "missing token",
			envContent: "COOLIFY_URL=https://coolify.example.com",
			wantErr:    "COOLIFY_TOKEN is not set",
		},
		{
			name:       "empty values",
			envContent: "COOLIFY_URL=\nCOOLIFY_TOKEN=",
			wantErr:    "COOLIFY_URL is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir := t.TempDir()
			envPath := filepath.Join(dir, ".env")

			// Write .env file
			if err := os.WriteFile(envPath, []byte(tt.envContent), 0644); err != nil {
				t.Fatalf("failed to write .env: %v", err)
			}

			// Load config
			_, err := LoadCoolifyConfig(dir)
			if err == nil {
				t.Fatal("LoadCoolifyConfig() expected error, got nil")
			}

			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestGetAppIDForEnvironment_Good(t *testing.T) {
	config := &CoolifyConfig{
		URL:          "https://coolify.example.com",
		Token:        "token",
		AppID:        "prod-123",
		StagingAppID: "staging-456",
	}

	tests := []struct {
		name   string
		env    Environment
		wantID string
	}{
		{
			name:   "production environment",
			env:    EnvProduction,
			wantID: "prod-123",
		},
		{
			name:   "staging environment",
			env:    EnvStaging,
			wantID: "staging-456",
		},
		{
			name:   "empty defaults to production",
			env:    "",
			wantID: "prod-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := getAppIDForEnvironment(config, tt.env)
			if id != tt.wantID {
				t.Errorf("getAppIDForEnvironment() = %q, want %q", id, tt.wantID)
			}
		})
	}
}

func TestGetAppIDForEnvironment_FallbackToProduction(t *testing.T) {
	config := &CoolifyConfig{
		URL:   "https://coolify.example.com",
		Token: "token",
		AppID: "prod-123",
		// No staging app ID
	}

	// Staging should fall back to production
	id := getAppIDForEnvironment(config, EnvStaging)
	if id != "prod-123" {
		t.Errorf("getAppIDForEnvironment(EnvStaging) = %q, want %q (should fallback)", id, "prod-123")
	}
}

func TestIsDeploymentComplete_Good(t *testing.T) {
	completeStatuses := []string{"finished", "success", "failed", "error", "cancelled"}
	for _, status := range completeStatuses {
		if !IsDeploymentComplete(status) {
			t.Errorf("IsDeploymentComplete(%q) = false, want true", status)
		}
	}

	incompleteStatuses := []string{"queued", "building", "deploying", "pending", "rolling_back"}
	for _, status := range incompleteStatuses {
		if IsDeploymentComplete(status) {
			t.Errorf("IsDeploymentComplete(%q) = true, want false", status)
		}
	}
}

func TestIsDeploymentSuccessful_Good(t *testing.T) {
	successStatuses := []string{"finished", "success"}
	for _, status := range successStatuses {
		if !IsDeploymentSuccessful(status) {
			t.Errorf("IsDeploymentSuccessful(%q) = false, want true", status)
		}
	}

	failedStatuses := []string{"failed", "error", "cancelled", "queued", "building"}
	for _, status := range failedStatuses {
		if IsDeploymentSuccessful(status) {
			t.Errorf("IsDeploymentSuccessful(%q) = true, want false", status)
		}
	}
}

func TestNewCoolifyClient_Good(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		wantBaseURL string
	}{
		{
			name:        "URL without trailing slash",
			baseURL:     "https://coolify.example.com",
			wantBaseURL: "https://coolify.example.com",
		},
		{
			name:        "URL with trailing slash",
			baseURL:     "https://coolify.example.com/",
			wantBaseURL: "https://coolify.example.com",
		},
		{
			name:        "URL with api path",
			baseURL:     "https://coolify.example.com/api/",
			wantBaseURL: "https://coolify.example.com/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewCoolifyClient(tt.baseURL, "token")
			if client.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL = %q, want %q", client.BaseURL, tt.wantBaseURL)
			}
			if client.Token != "token" {
				t.Errorf("Token = %q, want %q", client.Token, "token")
			}
			if client.HTTPClient == nil {
				t.Error("HTTPClient is nil")
			}
		})
	}
}
