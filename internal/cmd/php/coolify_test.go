package php

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoolifyClient_Good(t *testing.T) {
	t.Run("creates client with correct base URL", func(t *testing.T) {
		client := NewCoolifyClient("https://coolify.example.com", "token")

		assert.Equal(t, "https://coolify.example.com", client.BaseURL)
		assert.Equal(t, "token", client.Token)
		assert.NotNil(t, client.HTTPClient)
	})

	t.Run("strips trailing slash from base URL", func(t *testing.T) {
		client := NewCoolifyClient("https://coolify.example.com/", "token")
		assert.Equal(t, "https://coolify.example.com", client.BaseURL)
	})

	t.Run("http client has timeout", func(t *testing.T) {
		client := NewCoolifyClient("https://coolify.example.com", "token")
		assert.Equal(t, 30*time.Second, client.HTTPClient.Timeout)
	})
}

func TestCoolifyConfig_Good(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		config := CoolifyConfig{
			URL:          "https://coolify.example.com",
			Token:        "secret-token",
			AppID:        "app-123",
			StagingAppID: "staging-456",
		}

		assert.Equal(t, "https://coolify.example.com", config.URL)
		assert.Equal(t, "secret-token", config.Token)
		assert.Equal(t, "app-123", config.AppID)
		assert.Equal(t, "staging-456", config.StagingAppID)
	})
}

func TestCoolifyDeployment_Good(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		now := time.Now()
		deployment := CoolifyDeployment{
			ID:          "dep-123",
			Status:      "finished",
			CommitSHA:   "abc123",
			CommitMsg:   "Test commit",
			Branch:      "main",
			CreatedAt:   now,
			FinishedAt:  now.Add(5 * time.Minute),
			Log:         "Build successful",
			DeployedURL: "https://app.example.com",
		}

		assert.Equal(t, "dep-123", deployment.ID)
		assert.Equal(t, "finished", deployment.Status)
		assert.Equal(t, "abc123", deployment.CommitSHA)
		assert.Equal(t, "Test commit", deployment.CommitMsg)
		assert.Equal(t, "main", deployment.Branch)
	})
}

func TestCoolifyApp_Good(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		app := CoolifyApp{
			ID:          "app-123",
			Name:        "MyApp",
			FQDN:        "https://myapp.example.com",
			Status:      "running",
			Repository:  "https://github.com/user/repo",
			Branch:      "main",
			Environment: "production",
		}

		assert.Equal(t, "app-123", app.ID)
		assert.Equal(t, "MyApp", app.Name)
		assert.Equal(t, "https://myapp.example.com", app.FQDN)
		assert.Equal(t, "running", app.Status)
	})
}

func TestLoadCoolifyConfigFromFile_Good(t *testing.T) {
	t.Run("loads config from .env file", func(t *testing.T) {
		dir := t.TempDir()
		envContent := `COOLIFY_URL=https://coolify.example.com
COOLIFY_TOKEN=secret-token
COOLIFY_APP_ID=app-123
COOLIFY_STAGING_APP_ID=staging-456`

		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		config, err := LoadCoolifyConfigFromFile(filepath.Join(dir, ".env"))
		assert.NoError(t, err)
		assert.Equal(t, "https://coolify.example.com", config.URL)
		assert.Equal(t, "secret-token", config.Token)
		assert.Equal(t, "app-123", config.AppID)
		assert.Equal(t, "staging-456", config.StagingAppID)
	})

	t.Run("handles quoted values", func(t *testing.T) {
		dir := t.TempDir()
		envContent := `COOLIFY_URL="https://coolify.example.com"
COOLIFY_TOKEN='secret-token'`

		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		config, err := LoadCoolifyConfigFromFile(filepath.Join(dir, ".env"))
		assert.NoError(t, err)
		assert.Equal(t, "https://coolify.example.com", config.URL)
		assert.Equal(t, "secret-token", config.Token)
	})

	t.Run("ignores comments", func(t *testing.T) {
		dir := t.TempDir()
		envContent := `# This is a comment
COOLIFY_URL=https://coolify.example.com
# COOLIFY_TOKEN=wrong-token
COOLIFY_TOKEN=correct-token`

		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		config, err := LoadCoolifyConfigFromFile(filepath.Join(dir, ".env"))
		assert.NoError(t, err)
		assert.Equal(t, "correct-token", config.Token)
	})

	t.Run("ignores blank lines", func(t *testing.T) {
		dir := t.TempDir()
		envContent := `COOLIFY_URL=https://coolify.example.com

COOLIFY_TOKEN=secret-token`

		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		config, err := LoadCoolifyConfigFromFile(filepath.Join(dir, ".env"))
		assert.NoError(t, err)
		assert.Equal(t, "https://coolify.example.com", config.URL)
	})
}

func TestLoadCoolifyConfigFromFile_Bad(t *testing.T) {
	t.Run("fails when COOLIFY_URL missing", func(t *testing.T) {
		dir := t.TempDir()
		envContent := `COOLIFY_TOKEN=secret-token`

		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		_, err = LoadCoolifyConfigFromFile(filepath.Join(dir, ".env"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "COOLIFY_URL is not set")
	})

	t.Run("fails when COOLIFY_TOKEN missing", func(t *testing.T) {
		dir := t.TempDir()
		envContent := `COOLIFY_URL=https://coolify.example.com`

		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		_, err = LoadCoolifyConfigFromFile(filepath.Join(dir, ".env"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "COOLIFY_TOKEN is not set")
	})
}

func TestLoadCoolifyConfig_FromDirectory_Good(t *testing.T) {
	t.Run("loads from directory", func(t *testing.T) {
		dir := t.TempDir()
		envContent := `COOLIFY_URL=https://coolify.example.com
COOLIFY_TOKEN=secret-token`

		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		config, err := LoadCoolifyConfig(dir)
		assert.NoError(t, err)
		assert.Equal(t, "https://coolify.example.com", config.URL)
	})
}

func TestValidateCoolifyConfig_Bad(t *testing.T) {
	t.Run("returns error for empty URL", func(t *testing.T) {
		config := &CoolifyConfig{Token: "token"}
		_, err := validateCoolifyConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "COOLIFY_URL is not set")
	})

	t.Run("returns error for empty token", func(t *testing.T) {
		config := &CoolifyConfig{URL: "https://coolify.example.com"}
		_, err := validateCoolifyConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "COOLIFY_TOKEN is not set")
	})
}

func TestCoolifyClient_TriggerDeploy_Good(t *testing.T) {
	t.Run("triggers deployment successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/applications/app-123/deploy", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "Bearer secret-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			resp := CoolifyDeployment{
				ID:        "dep-456",
				Status:    "queued",
				CreatedAt: time.Now(),
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "secret-token")
		deployment, err := client.TriggerDeploy(context.Background(), "app-123", false)

		assert.NoError(t, err)
		assert.Equal(t, "dep-456", deployment.ID)
		assert.Equal(t, "queued", deployment.Status)
	})

	t.Run("triggers deployment with force", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, true, body["force"])

			resp := CoolifyDeployment{ID: "dep-456", Status: "queued"}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "secret-token")
		_, err := client.TriggerDeploy(context.Background(), "app-123", true)
		assert.NoError(t, err)
	})

	t.Run("handles minimal response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Return an invalid JSON response to trigger the fallback
			_, _ = w.Write([]byte("not json"))
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "secret-token")
		deployment, err := client.TriggerDeploy(context.Background(), "app-123", false)

		assert.NoError(t, err)
		// The fallback response should be returned
		assert.Equal(t, "queued", deployment.Status)
	})
}

func TestCoolifyClient_TriggerDeploy_Bad(t *testing.T) {
	t.Run("fails on HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "secret-token")
		_, err := client.TriggerDeploy(context.Background(), "app-123", false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})
}

func TestCoolifyClient_GetDeployment_Good(t *testing.T) {
	t.Run("gets deployment details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/applications/app-123/deployments/dep-456", r.URL.Path)
			assert.Equal(t, "GET", r.Method)

			resp := CoolifyDeployment{
				ID:        "dep-456",
				Status:    "finished",
				CommitSHA: "abc123",
				Branch:    "main",
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "secret-token")
		deployment, err := client.GetDeployment(context.Background(), "app-123", "dep-456")

		assert.NoError(t, err)
		assert.Equal(t, "dep-456", deployment.ID)
		assert.Equal(t, "finished", deployment.Status)
		assert.Equal(t, "abc123", deployment.CommitSHA)
	})
}

func TestCoolifyClient_GetDeployment_Bad(t *testing.T) {
	t.Run("fails on 404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "secret-token")
		_, err := client.GetDeployment(context.Background(), "app-123", "dep-456")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Not found")
	})
}

func TestCoolifyClient_ListDeployments_Good(t *testing.T) {
	t.Run("lists deployments", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/applications/app-123/deployments", r.URL.Path)
			assert.Equal(t, "10", r.URL.Query().Get("limit"))

			resp := []CoolifyDeployment{
				{ID: "dep-1", Status: "finished"},
				{ID: "dep-2", Status: "failed"},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "secret-token")
		deployments, err := client.ListDeployments(context.Background(), "app-123", 10)

		assert.NoError(t, err)
		assert.Len(t, deployments, 2)
		assert.Equal(t, "dep-1", deployments[0].ID)
		assert.Equal(t, "dep-2", deployments[1].ID)
	})

	t.Run("lists without limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "", r.URL.Query().Get("limit"))
			_ = json.NewEncoder(w).Encode([]CoolifyDeployment{})
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "secret-token")
		_, err := client.ListDeployments(context.Background(), "app-123", 0)
		assert.NoError(t, err)
	})
}

func TestCoolifyClient_Rollback_Good(t *testing.T) {
	t.Run("triggers rollback", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/applications/app-123/rollback", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "dep-old", body["deployment_id"])

			resp := CoolifyDeployment{
				ID:     "dep-new",
				Status: "rolling_back",
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "secret-token")
		deployment, err := client.Rollback(context.Background(), "app-123", "dep-old")

		assert.NoError(t, err)
		assert.Equal(t, "dep-new", deployment.ID)
		assert.Equal(t, "rolling_back", deployment.Status)
	})
}

func TestCoolifyClient_GetApp_Good(t *testing.T) {
	t.Run("gets app details", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/applications/app-123", r.URL.Path)
			assert.Equal(t, "GET", r.Method)

			resp := CoolifyApp{
				ID:     "app-123",
				Name:   "MyApp",
				FQDN:   "https://myapp.example.com",
				Status: "running",
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "secret-token")
		app, err := client.GetApp(context.Background(), "app-123")

		assert.NoError(t, err)
		assert.Equal(t, "app-123", app.ID)
		assert.Equal(t, "MyApp", app.Name)
		assert.Equal(t, "https://myapp.example.com", app.FQDN)
	})
}

func TestCoolifyClient_SetHeaders(t *testing.T) {
	t.Run("sets all required headers", func(t *testing.T) {
		client := NewCoolifyClient("https://coolify.example.com", "my-token")
		req, _ := http.NewRequest("GET", "https://coolify.example.com", nil)

		client.setHeaders(req)

		assert.Equal(t, "Bearer my-token", req.Header.Get("Authorization"))
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", req.Header.Get("Accept"))
	})
}

func TestCoolifyClient_ParseError(t *testing.T) {
	t.Run("parses message field", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Bad request message"})
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "token")
		_, err := client.GetApp(context.Background(), "app-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Bad request message")
	})

	t.Run("parses error field", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Error message"})
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "token")
		_, err := client.GetApp(context.Background(), "app-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Error message")
	})

	t.Run("returns raw body when no JSON fields", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Raw error message"))
		}))
		defer server.Close()

		client := NewCoolifyClient(server.URL, "token")
		_, err := client.GetApp(context.Background(), "app-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Raw error message")
	})
}

func TestEnvironmentVariablePriority(t *testing.T) {
	t.Run("env vars take precedence over .env file", func(t *testing.T) {
		dir := t.TempDir()
		envContent := `COOLIFY_URL=https://from-file.com
COOLIFY_TOKEN=file-token`

		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		// Set environment variables
		origURL := os.Getenv("COOLIFY_URL")
		origToken := os.Getenv("COOLIFY_TOKEN")
		defer func() {
			_ = os.Setenv("COOLIFY_URL", origURL)
			_ = os.Setenv("COOLIFY_TOKEN", origToken)
		}()

		_ = os.Setenv("COOLIFY_URL", "https://from-env.com")
		_ = os.Setenv("COOLIFY_TOKEN", "env-token")

		config, err := LoadCoolifyConfig(dir)
		assert.NoError(t, err)
		// Environment variables should take precedence
		assert.Equal(t, "https://from-env.com", config.URL)
		assert.Equal(t, "env-token", config.Token)
	})
}
