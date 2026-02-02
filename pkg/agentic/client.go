package agentic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/log"
)

// Client is the API client for the core-agentic service.
type Client struct {
	// BaseURL is the base URL of the API server.
	BaseURL string
	// Token is the authentication token.
	Token string
	// HTTPClient is the HTTP client used for requests.
	HTTPClient *http.Client
	// AgentID is the identifier for this agent when claiming tasks.
	AgentID string
}

// NewClient creates a new agentic API client with the given base URL and token.
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientFromConfig creates a new client from a Config struct.
func NewClientFromConfig(cfg *Config) *Client {
	client := NewClient(cfg.BaseURL, cfg.Token)
	client.AgentID = cfg.AgentID
	return client
}

// ListTasks retrieves a list of tasks matching the given options.
func (c *Client) ListTasks(ctx context.Context, opts ListOptions) ([]Task, error) {
	const op = "agentic.Client.ListTasks"

	// Build query parameters
	params := url.Values{}
	if opts.Status != "" {
		params.Set("status", string(opts.Status))
	}
	if opts.Priority != "" {
		params.Set("priority", string(opts.Priority))
	}
	if opts.Project != "" {
		params.Set("project", opts.Project)
	}
	if opts.ClaimedBy != "" {
		params.Set("claimed_by", opts.ClaimedBy)
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if len(opts.Labels) > 0 {
		params.Set("labels", strings.Join(opts.Labels, ","))
	}

	endpoint := c.BaseURL + "/api/tasks"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, log.E(op, "failed to create request", err)
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, log.E(op, "request failed", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, log.E(op, "API error", err)
	}

	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, log.E(op, "failed to decode response", err)
	}

	return tasks, nil
}

// GetTask retrieves a single task by its ID.
func (c *Client) GetTask(ctx context.Context, id string) (*Task, error) {
	const op = "agentic.Client.GetTask"

	if id == "" {
		return nil, log.E(op, "task ID is required", nil)
	}

	endpoint := fmt.Sprintf("%s/api/tasks/%s", c.BaseURL, url.PathEscape(id))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, log.E(op, "failed to create request", err)
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, log.E(op, "request failed", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, log.E(op, "API error", err)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, log.E(op, "failed to decode response", err)
	}

	return &task, nil
}

// ClaimTask claims a task for the current agent.
func (c *Client) ClaimTask(ctx context.Context, id string) (*Task, error) {
	const op = "agentic.Client.ClaimTask"

	if id == "" {
		return nil, log.E(op, "task ID is required", nil)
	}

	endpoint := fmt.Sprintf("%s/api/tasks/%s/claim", c.BaseURL, url.PathEscape(id))

	// Include agent ID in the claim request if available
	var body io.Reader
	if c.AgentID != "" {
		data, _ := json.Marshal(map[string]string{"agent_id": c.AgentID})
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, log.E(op, "failed to create request", err)
	}

	c.setHeaders(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, log.E(op, "request failed", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, log.E(op, "API error", err)
	}

	// Read body once to allow multiple decode attempts
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, log.E(op, "failed to read response", err)
	}

	// Try decoding as ClaimResponse first
	var result ClaimResponse
	if err := json.Unmarshal(bodyData, &result); err == nil && result.Task != nil {
		return result.Task, nil
	}

	// Try decoding as just a Task for simpler API responses
	var task Task
	if err := json.Unmarshal(bodyData, &task); err != nil {
		return nil, log.E(op, "failed to decode response", err)
	}

	return &task, nil
}

// UpdateTask updates a task with new status, progress, or notes.
func (c *Client) UpdateTask(ctx context.Context, id string, update TaskUpdate) error {
	const op = "agentic.Client.UpdateTask"

	if id == "" {
		return log.E(op, "task ID is required", nil)
	}

	endpoint := fmt.Sprintf("%s/api/tasks/%s", c.BaseURL, url.PathEscape(id))

	data, err := json.Marshal(update)
	if err != nil {
		return log.E(op, "failed to marshal update", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, bytes.NewReader(data))
	if err != nil {
		return log.E(op, "failed to create request", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return log.E(op, "request failed", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return log.E(op, "API error", err)
	}

	return nil
}

// CompleteTask marks a task as completed with the given result.
func (c *Client) CompleteTask(ctx context.Context, id string, result TaskResult) error {
	const op = "agentic.Client.CompleteTask"

	if id == "" {
		return log.E(op, "task ID is required", nil)
	}

	endpoint := fmt.Sprintf("%s/api/tasks/%s/complete", c.BaseURL, url.PathEscape(id))

	data, err := json.Marshal(result)
	if err != nil {
		return log.E(op, "failed to marshal result", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return log.E(op, "failed to create request", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return log.E(op, "request failed", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return log.E(op, "API error", err)
	}

	return nil
}

// setHeaders adds common headers to the request.
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "core-agentic-client/1.0")
}

// checkResponse checks if the response indicates an error.
func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)

	// Try to parse as APIError
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Message != "" {
		apiErr.Code = resp.StatusCode
		return &apiErr
	}

	// Return generic error
	return &APIError{
		Code:    resp.StatusCode,
		Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		Details: string(body),
	}
}

// Ping tests the connection to the API server.
func (c *Client) Ping(ctx context.Context) error {
	const op = "agentic.Client.Ping"

	endpoint := c.BaseURL + "/api/health"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return log.E(op, "failed to create request", err)
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return log.E(op, "request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return log.E(op, fmt.Sprintf("server returned status %d", resp.StatusCode), nil)
	}

	return nil
}
