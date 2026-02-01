package php

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
)

// CoolifyClient is an HTTP client for the Coolify API.
type CoolifyClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// CoolifyConfig holds configuration loaded from environment.
type CoolifyConfig struct {
	URL          string
	Token        string
	AppID        string
	StagingAppID string
}

// CoolifyDeployment represents a deployment from the Coolify API.
type CoolifyDeployment struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	CommitSHA   string    `json:"commit_sha,omitempty"`
	CommitMsg   string    `json:"commit_message,omitempty"`
	Branch      string    `json:"branch,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	FinishedAt  time.Time `json:"finished_at,omitempty"`
	Log         string    `json:"log,omitempty"`
	DeployedURL string    `json:"deployed_url,omitempty"`
}

// CoolifyApp represents an application from the Coolify API.
type CoolifyApp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	FQDN        string `json:"fqdn,omitempty"`
	Status      string `json:"status,omitempty"`
	Repository  string `json:"repository,omitempty"`
	Branch      string `json:"branch,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// NewCoolifyClient creates a new Coolify API client.
func NewCoolifyClient(baseURL, token string) *CoolifyClient {
	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &CoolifyClient{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// LoadCoolifyConfig loads Coolify configuration from .env file in the given directory.
func LoadCoolifyConfig(dir string) (*CoolifyConfig, error) {
	envPath := filepath.Join(dir, ".env")
	return LoadCoolifyConfigFromFile(envPath)
}

// LoadCoolifyConfigFromFile loads Coolify configuration from a specific .env file.
func LoadCoolifyConfigFromFile(path string) (*CoolifyConfig, error) {
	config := &CoolifyConfig{}

	// First try environment variables
	config.URL = os.Getenv("COOLIFY_URL")
	config.Token = os.Getenv("COOLIFY_TOKEN")
	config.AppID = os.Getenv("COOLIFY_APP_ID")
	config.StagingAppID = os.Getenv("COOLIFY_STAGING_APP_ID")

	// Then try .env file
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No .env file, just use env vars
			return validateCoolifyConfig(config)
		}
		return nil, cli.WrapVerb(err, "open", ".env file")
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, cli.WrapVerb(err, "read", ".env file")
	}

	// Parse .env file
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Only override if not already set from env
		switch key {
		case "COOLIFY_URL":
			if config.URL == "" {
				config.URL = value
			}
		case "COOLIFY_TOKEN":
			if config.Token == "" {
				config.Token = value
			}
		case "COOLIFY_APP_ID":
			if config.AppID == "" {
				config.AppID = value
			}
		case "COOLIFY_STAGING_APP_ID":
			if config.StagingAppID == "" {
				config.StagingAppID = value
			}
		}
	}

	return validateCoolifyConfig(config)
}

// validateCoolifyConfig checks that required fields are set.
func validateCoolifyConfig(config *CoolifyConfig) (*CoolifyConfig, error) {
	if config.URL == "" {
		return nil, cli.Err("COOLIFY_URL is not set")
	}
	if config.Token == "" {
		return nil, cli.Err("COOLIFY_TOKEN is not set")
	}
	return config, nil
}

// TriggerDeploy triggers a deployment for the specified application.
func (c *CoolifyClient) TriggerDeploy(ctx context.Context, appID string, force bool) (*CoolifyDeployment, error) {
	endpoint := cli.Sprintf("%s/api/v1/applications/%s/deploy", c.BaseURL, appID)

	payload := map[string]interface{}{}
	if force {
		payload["force"] = true
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, cli.WrapVerb(err, "marshal", "request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, cli.WrapVerb(err, "create", "request")
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, cli.Wrap(err, "request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, c.parseError(resp)
	}

	var deployment CoolifyDeployment
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		// Some Coolify versions return minimal response
		return &CoolifyDeployment{
			Status:    "queued",
			CreatedAt: time.Now(),
		}, nil
	}

	return &deployment, nil
}

// GetDeployment retrieves a specific deployment by ID.
func (c *CoolifyClient) GetDeployment(ctx context.Context, appID, deploymentID string) (*CoolifyDeployment, error) {
	endpoint := cli.Sprintf("%s/api/v1/applications/%s/deployments/%s", c.BaseURL, appID, deploymentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, cli.WrapVerb(err, "create", "request")
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, cli.Wrap(err, "request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var deployment CoolifyDeployment
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return nil, cli.WrapVerb(err, "decode", "response")
	}

	return &deployment, nil
}

// ListDeployments retrieves deployments for an application.
func (c *CoolifyClient) ListDeployments(ctx context.Context, appID string, limit int) ([]CoolifyDeployment, error) {
	endpoint := cli.Sprintf("%s/api/v1/applications/%s/deployments", c.BaseURL, appID)
	if limit > 0 {
		endpoint = cli.Sprintf("%s?limit=%d", endpoint, limit)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, cli.WrapVerb(err, "create", "request")
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, cli.Wrap(err, "request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var deployments []CoolifyDeployment
	if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
		return nil, cli.WrapVerb(err, "decode", "response")
	}

	return deployments, nil
}

// Rollback triggers a rollback to a previous deployment.
func (c *CoolifyClient) Rollback(ctx context.Context, appID, deploymentID string) (*CoolifyDeployment, error) {
	endpoint := cli.Sprintf("%s/api/v1/applications/%s/rollback", c.BaseURL, appID)

	payload := map[string]interface{}{
		"deployment_id": deploymentID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, cli.WrapVerb(err, "marshal", "request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, cli.WrapVerb(err, "create", "request")
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, cli.Wrap(err, "request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, c.parseError(resp)
	}

	var deployment CoolifyDeployment
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return &CoolifyDeployment{
			Status:    "rolling_back",
			CreatedAt: time.Now(),
		}, nil
	}

	return &deployment, nil
}

// GetApp retrieves application details.
func (c *CoolifyClient) GetApp(ctx context.Context, appID string) (*CoolifyApp, error) {
	endpoint := cli.Sprintf("%s/api/v1/applications/%s", c.BaseURL, appID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, cli.WrapVerb(err, "create", "request")
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, cli.Wrap(err, "request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var app CoolifyApp
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, cli.WrapVerb(err, "decode", "response")
	}

	return &app, nil
}

// setHeaders sets common headers for API requests.
func (c *CoolifyClient) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}

// parseError extracts error information from an API response.
func (c *CoolifyClient) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Message != "" {
			return cli.Err("API error (%d): %s", resp.StatusCode, errResp.Message)
		}
		if errResp.Error != "" {
			return cli.Err("API error (%d): %s", resp.StatusCode, errResp.Error)
		}
	}

	return cli.Err("API error (%d): %s", resp.StatusCode, string(body))
}
