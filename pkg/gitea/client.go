// Package gitea provides a thin wrapper around the Gitea Go SDK
// for managing repositories, issues, and pull requests on a Gitea instance.
//
// Authentication is resolved from config file, environment variables, or flag overrides:
//
//  1. ~/.core/config.yaml keys: gitea.token, gitea.url
//  2. GITEA_TOKEN + GITEA_URL environment variables (override config file)
//  3. Flag overrides via core gitea config --url/--token (highest priority)
package gitea

import (
	"code.gitea.io/sdk/gitea"

	"github.com/host-uk/core/pkg/log"
)

// Client wraps the Gitea SDK client with config-based auth.
type Client struct {
	api *gitea.Client
	url string
}

// New creates a new Gitea API client for the given URL and token.
func New(url, token string) (*Client, error) {
	api, err := gitea.NewClient(url, gitea.SetToken(token))
	if err != nil {
		return nil, log.E("gitea.New", "failed to create client", err)
	}

	return &Client{api: api, url: url}, nil
}

// API exposes the underlying SDK client for direct access.
func (c *Client) API() *gitea.Client { return c.api }

// URL returns the Gitea instance URL.
func (c *Client) URL() string { return c.url }
