package gitea

import (
	"os"

	"github.com/host-uk/core/pkg/config"
	"github.com/host-uk/core/pkg/log"
)

const (
	// ConfigKeyURL is the config key for the Gitea instance URL.
	ConfigKeyURL = "gitea.url"
	// ConfigKeyToken is the config key for the Gitea API token.
	ConfigKeyToken = "gitea.token"

	// DefaultURL is the default Gitea instance URL.
	DefaultURL = "https://gitea.snider.dev"
)

// NewFromConfig creates a Gitea client using the standard config resolution:
//
//  1. ~/.core/config.yaml keys: gitea.token, gitea.url
//  2. GITEA_TOKEN + GITEA_URL environment variables (override config file)
//  3. Provided flag overrides (highest priority; pass empty to skip)
func NewFromConfig(flagURL, flagToken string) (*Client, error) {
	url, token, err := ResolveConfig(flagURL, flagToken)
	if err != nil {
		return nil, err
	}

	if token == "" {
		return nil, log.E("gitea.NewFromConfig", "no API token configured (set GITEA_TOKEN or run: core gitea config --token TOKEN)", nil)
	}

	return New(url, token)
}

// ResolveConfig resolves the Gitea URL and token from all config sources.
// Flag values take highest priority, then env vars, then config file.
func ResolveConfig(flagURL, flagToken string) (url, token string, err error) {
	// Start with config file values
	cfg, cfgErr := config.New()
	if cfgErr == nil {
		_ = cfg.Get(ConfigKeyURL, &url)
		_ = cfg.Get(ConfigKeyToken, &token)
	}

	// Overlay environment variables
	if envURL := os.Getenv("GITEA_URL"); envURL != "" {
		url = envURL
	}
	if envToken := os.Getenv("GITEA_TOKEN"); envToken != "" {
		token = envToken
	}

	// Overlay flag values (highest priority)
	if flagURL != "" {
		url = flagURL
	}
	if flagToken != "" {
		token = flagToken
	}

	// Default URL if nothing configured
	if url == "" {
		url = DefaultURL
	}

	return url, token, nil
}

// SaveConfig persists the Gitea URL and/or token to the config file.
func SaveConfig(url, token string) error {
	cfg, err := config.New()
	if err != nil {
		return log.E("gitea.SaveConfig", "failed to load config", err)
	}

	if url != "" {
		if err := cfg.Set(ConfigKeyURL, url); err != nil {
			return log.E("gitea.SaveConfig", "failed to save URL", err)
		}
	}

	if token != "" {
		if err := cfg.Set(ConfigKeyToken, token); err != nil {
			return log.E("gitea.SaveConfig", "failed to save token", err)
		}
	}

	return nil
}
