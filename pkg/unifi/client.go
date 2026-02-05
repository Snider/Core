package unifi

import (
	"crypto/tls"
	"net/http"

	uf "github.com/unpoller/unifi/v5"

	"github.com/host-uk/core/pkg/log"
)

// Client wraps the unpoller UniFi client with config-based auth.
type Client struct {
	api *uf.Unifi
	url string
}

// New creates a new UniFi API client for the given controller URL and credentials.
func New(url, user, pass, apikey string, verifyTLS bool) (*Client, error) {
	cfg := &uf.Config{
		URL:    url,
		User:   user,
		Pass:   pass,
		APIKey: apikey,
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if !verifyTLS {
		// Only disable verification if explicitly requested (e.g. for home lab self-signed certs)
		tlsConfig.InsecureSkipVerify = true //nolint:gosec
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	api, err := uf.NewUnifi(cfg)
	if err != nil {
		return nil, log.E("unifi.New", "failed to create client", err)
	}

	// Override the HTTP client to skip TLS verification
	api.Client = httpClient

	return &Client{api: api, url: url}, nil
}

// API exposes the underlying SDK client for direct access.
func (c *Client) API() *uf.Unifi { return c.api }

// URL returns the UniFi controller URL.
func (c *Client) URL() string { return c.url }
