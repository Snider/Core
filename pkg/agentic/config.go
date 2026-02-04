package agentic

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/log"
	"gopkg.in/yaml.v3"
)

// Config holds the configuration for connecting to the core-agentic service.
type Config struct {
	// BaseURL is the URL of the core-agentic API server.
	BaseURL string `yaml:"base_url" json:"base_url"`
	// Token is the authentication token for API requests.
	Token string `yaml:"token" json:"token"`
	// DefaultProject is the project to use when none is specified.
	DefaultProject string `yaml:"default_project" json:"default_project"`
	// AgentID is the identifier for this agent (optional, used for claiming tasks).
	AgentID string `yaml:"agent_id" json:"agent_id"`
}

// configFileName is the name of the YAML config file.
const configFileName = "agentic.yaml"

// envFileName is the name of the environment file.
const envFileName = ".env"

// DefaultBaseURL is the default API endpoint if none is configured.
const DefaultBaseURL = "https://api.core-agentic.dev"

// LoadConfig loads the agentic configuration from the specified directory.
// It first checks for a .env file, then falls back to ~/.core/agentic.yaml.
// If dir is empty, it checks the current directory first.
//
// Environment variables take precedence:
//   - AGENTIC_BASE_URL: API base URL
//   - AGENTIC_TOKEN: Authentication token
//   - AGENTIC_PROJECT: Default project
//   - AGENTIC_AGENT_ID: Agent identifier
func LoadConfig(m io.Medium, dir string) (*Config, error) {
	cfg := &Config{
		BaseURL: DefaultBaseURL,
	}

	// Try loading from .env file in the specified directory
	if dir != "" {
		envPath := filepath.Join(dir, envFileName)
		if err := loadEnvFile(m, envPath, cfg); err == nil {
			// Successfully loaded from .env
			applyEnvOverrides(cfg)
			if cfg.Token != "" {
				return cfg, nil
			}
		}
	}

	// Try loading from current directory .env
	if dir == "" {
		cwd, err := os.Getwd()
		if err == nil {
			envPath := filepath.Join(cwd, envFileName)
			if err := loadEnvFile(m, envPath, cfg); err == nil {
				applyEnvOverrides(cfg)
				if cfg.Token != "" {
					return cfg, nil
				}
			}
		}
	}

	// Try loading from ~/.core/agentic.yaml
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, log.E("agentic.LoadConfig", "failed to get home directory", err)
	}

	configPath := filepath.Join(homeDir, ".core", configFileName)
	if err := loadYAMLConfig(m, configPath, cfg); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, log.E("agentic.LoadConfig", "failed to load config", err)
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
func loadEnvFile(m io.Medium, path string, cfg *Config) error {
	content, err := m.Read(path)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
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

// loadYAMLConfig reads configuration from a YAML file.
func loadYAMLConfig(m io.Medium, path string, cfg *Config) error {
	data, err := m.Read(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal([]byte(data), cfg)
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
func SaveConfig(m io.Medium, cfg *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return log.E("agentic.SaveConfig", "failed to get home directory", err)
	}

	configDir := filepath.Join(homeDir, ".core")
	if err := m.EnsureDir(configDir); err != nil {
		return log.E("agentic.SaveConfig", "failed to create config directory", err)
	}

	configPath := filepath.Join(configDir, configFileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return log.E("agentic.SaveConfig", "failed to marshal config", err)
	}

	if err := m.Write(configPath, string(data)); err != nil {
		return log.E("agentic.SaveConfig", "failed to write config file", err)
	}

	return nil
}

// ConfigPath returns the path to the config file in the user's home directory.
func ConfigPath(m io.Medium) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", log.E("agentic.ConfigPath", "failed to get home directory", err)
	}
	return filepath.Join(homeDir, ".core", configFileName), nil
}
