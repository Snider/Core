package agentic

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/config"
	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/log"
)

// Config holds the configuration for connecting to the core-agentic service.
type Config struct {
	// BaseURL is the URL of the core-agentic API server.
	BaseURL string `yaml:"base_url" json:"base_url" mapstructure:"base_url"`
	// Token is the authentication token for API requests.
	Token string `yaml:"token" json:"token" mapstructure:"token"`
	// DefaultProject is the project to use when none is specified.
	DefaultProject string `yaml:"default_project" json:"default_project" mapstructure:"default_project"`
	// AgentID is the identifier for this agent (optional, used for claiming tasks).
	AgentID string `yaml:"agent_id" json:"agent_id" mapstructure:"agent_id"`
}

// configFileName is the name of the YAML config file.
const configFileName = "agentic.yaml"

// envFileName is the name of the environment file.
const envFileName = ".env"

// DefaultBaseURL is the default API endpoint if none is configured.
const DefaultBaseURL = "https://api.core-agentic.dev"

// LoadConfig loads the agentic configuration from the specified directory.
// It uses the centralized config service.
//
// Environment variables take precedence (prefix: AGENTIC_):
//   - AGENTIC_BASE_URL: API base URL
//   - AGENTIC_TOKEN: Authentication token
//   - AGENTIC_PROJECT: Default project
//   - AGENTIC_AGENT_ID: Agent identifier
func LoadConfig(dir string) (*Config, error) {
	cfg := &Config{
		BaseURL: DefaultBaseURL,
	}

	// Try loading from .env file in the specified directory
	if dir != "" {
		envPath := filepath.Join(dir, envFileName)
		if err := loadEnvFile(envPath, cfg); err == nil {
			// Successfully loaded from .env
			applyEnvOverrides(cfg)
			if cfg.Token != "" {
				return cfg, nil
			}
		}
	}

	// Try loading from current directory .env
	cwd, err := os.Getwd()
	if err == nil {
		envPath := filepath.Join(cwd, envFileName)
		if err := loadEnvFile(envPath, cfg); err == nil {
			applyEnvOverrides(cfg)
			if cfg.Token != "" {
				return cfg, nil
			}
		}
	}

	// Try loading from ~/.core/agentic.yaml
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, log.E("agentic.LoadConfig", "failed to get home directory", err)
	}

	configPath := filepath.Join(homeDir, ".core", configFileName)
	if io.Local.IsFile(configPath) {
		// Use centralized config service to load the YAML file
		c, err := config.New(config.WithPath(configPath))
		if err != nil {
			return nil, log.E("agentic.LoadConfig", "failed to initialize config", err)
		}
		if err := c.Get("", cfg); err != nil {
			return nil, log.E("agentic.LoadConfig", "failed to load config", err)
		}
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Validate configuration
	if cfg.Token == "" {
		return nil, log.E("agentic.LoadConfig", "no authentication token configured", nil)
	}

	return cfg, nil
}

// loadEnvFile reads a .env file and extracts agentic configuration.
func loadEnvFile(path string, cfg *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		switch key {
		case "AGENTIC_BASE_URL":
			cfg.BaseURL = value
		case "AGENTIC_TOKEN":
			cfg.Token = value
		case "AGENTIC_PROJECT":
			cfg.DefaultProject = value
		case "AGENTIC_AGENT_ID":
			cfg.AgentID = value
		}
	}

	return scanner.Err()
}

// applyEnvOverrides applies environment variable overrides to the config.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("AGENTIC_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("AGENTIC_TOKEN"); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv("AGENTIC_PROJECT"); v != "" {
		cfg.DefaultProject = v
	}
	if v := os.Getenv("AGENTIC_AGENT_ID"); v != "" {
		cfg.AgentID = v
	}
}

// SaveConfig saves the configuration to ~/.core/agentic.yaml.
func SaveConfig(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data := make(map[string]any)
	data["base_url"] = cfg.BaseURL
	data["token"] = cfg.Token
	data["default_project"] = cfg.DefaultProject
	data["agent_id"] = cfg.AgentID

	return config.Save(io.Local, path, data)
}

// ConfigPath returns the path to the config file in the user's home directory.
func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", log.E("agentic.ConfigPath", "failed to get home directory", err)
	}
	return filepath.Join(homeDir, ".core", configFileName), nil
}
