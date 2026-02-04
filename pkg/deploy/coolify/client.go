package coolify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/host-uk/core/pkg/deploy/python"
)

// Client wraps the Python CoolifyClient for Go usage.
type Client struct {
	baseURL   string
	apiToken  string
	timeout   int
	verifySSL bool

	mu sync.Mutex
}

// Config holds Coolify client configuration.
type Config struct {
	BaseURL   string
	APIToken  string
	Timeout   int
	VerifySSL bool
}

// DefaultConfig returns default configuration from environment.
func DefaultConfig() Config {
	return Config{
		BaseURL:   os.Getenv("COOLIFY_URL"),
		APIToken:  os.Getenv("COOLIFY_TOKEN"),
		Timeout:   30,
		VerifySSL: true,
	}
}

// NewClient creates a new Coolify client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("COOLIFY_URL not set")
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("COOLIFY_TOKEN not set")
	}

	// Initialize Python runtime
	if err := python.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize Python: %w", err)
	}

	return &Client{
		baseURL:   cfg.BaseURL,
		apiToken:  cfg.APIToken,
		timeout:   cfg.Timeout,
		verifySSL: cfg.VerifySSL,
	}, nil
}

// Call invokes a Coolify API operation by operationId.
func (c *Client) Call(ctx context.Context, operationID string, params map[string]any) (map[string]any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if params == nil {
		params = map[string]any{}
	}

	// Generate and run Python script
	script, err := python.CoolifyScript(c.baseURL, c.apiToken, operationID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate script: %w", err)
	}
	output, err := python.RunScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("API call %s failed: %w", operationID, err)
	}

	// Parse JSON result
	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		// Try parsing as array
		var arrResult []any
		if err2 := json.Unmarshal([]byte(output), &arrResult); err2 == nil {
			return map[string]any{"result": arrResult}, nil
		}
		return nil, fmt.Errorf("failed to parse response: %w (output: %s)", err, output)
	}

	return result, nil
}

// ListServers returns all servers.
func (c *Client) ListServers(ctx context.Context) ([]map[string]any, error) {
	result, err := c.Call(ctx, "list-servers", nil)
	if err != nil {
		return nil, err
	}
	return extractArray(result)
}

// GetServer returns a server by UUID.
func (c *Client) GetServer(ctx context.Context, uuid string) (map[string]any, error) {
	return c.Call(ctx, "get-server-by-uuid", map[string]any{"uuid": uuid})
}

// ValidateServer validates a server by UUID.
func (c *Client) ValidateServer(ctx context.Context, uuid string) (map[string]any, error) {
	return c.Call(ctx, "validate-server-by-uuid", map[string]any{"uuid": uuid})
}

// ListProjects returns all projects.
func (c *Client) ListProjects(ctx context.Context) ([]map[string]any, error) {
	result, err := c.Call(ctx, "list-projects", nil)
	if err != nil {
		return nil, err
	}
	return extractArray(result)
}

// GetProject returns a project by UUID.
func (c *Client) GetProject(ctx context.Context, uuid string) (map[string]any, error) {
	return c.Call(ctx, "get-project-by-uuid", map[string]any{"uuid": uuid})
}

// CreateProject creates a new project.
func (c *Client) CreateProject(ctx context.Context, name, description string) (map[string]any, error) {
	return c.Call(ctx, "create-project", map[string]any{
		"name":        name,
		"description": description,
	})
}

// ListApplications returns all applications.
func (c *Client) ListApplications(ctx context.Context) ([]map[string]any, error) {
	result, err := c.Call(ctx, "list-applications", nil)
	if err != nil {
		return nil, err
	}
	return extractArray(result)
}

// GetApplication returns an application by UUID.
func (c *Client) GetApplication(ctx context.Context, uuid string) (map[string]any, error) {
	return c.Call(ctx, "get-application-by-uuid", map[string]any{"uuid": uuid})
}

// DeployApplication triggers deployment of an application.
func (c *Client) DeployApplication(ctx context.Context, uuid string) (map[string]any, error) {
	return c.Call(ctx, "deploy-by-tag-or-uuid", map[string]any{"uuid": uuid})
}

// ListDatabases returns all databases.
func (c *Client) ListDatabases(ctx context.Context) ([]map[string]any, error) {
	result, err := c.Call(ctx, "list-databases", nil)
	if err != nil {
		return nil, err
	}
	return extractArray(result)
}

// GetDatabase returns a database by UUID.
func (c *Client) GetDatabase(ctx context.Context, uuid string) (map[string]any, error) {
	return c.Call(ctx, "get-database-by-uuid", map[string]any{"uuid": uuid})
}

// ListServices returns all services.
func (c *Client) ListServices(ctx context.Context) ([]map[string]any, error) {
	result, err := c.Call(ctx, "list-services", nil)
	if err != nil {
		return nil, err
	}
	return extractArray(result)
}

// GetService returns a service by UUID.
func (c *Client) GetService(ctx context.Context, uuid string) (map[string]any, error) {
	return c.Call(ctx, "get-service-by-uuid", map[string]any{"uuid": uuid})
}

// ListEnvironments returns environments for a project.
func (c *Client) ListEnvironments(ctx context.Context, projectUUID string) ([]map[string]any, error) {
	result, err := c.Call(ctx, "get-environments", map[string]any{"project_uuid": projectUUID})
	if err != nil {
		return nil, err
	}
	return extractArray(result)
}

// GetTeam returns the current team.
func (c *Client) GetTeam(ctx context.Context) (map[string]any, error) {
	return c.Call(ctx, "get-current-team", nil)
}

// GetTeamMembers returns members of the current team.
func (c *Client) GetTeamMembers(ctx context.Context) ([]map[string]any, error) {
	result, err := c.Call(ctx, "get-current-team-members", nil)
	if err != nil {
		return nil, err
	}
	return extractArray(result)
}

// extractArray extracts an array from result["result"] or returns empty.
func extractArray(result map[string]any) ([]map[string]any, error) {
	if arr, ok := result["result"].([]any); ok {
		items := make([]map[string]any, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				items = append(items, m)
			}
		}
		return items, nil
	}
	return nil, nil
}
