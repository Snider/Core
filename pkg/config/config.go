// Package config provides layered configuration management for the Core framework.
//
// Configuration values are resolved in priority order: defaults -> file -> env -> flags.
// Values are stored in a YAML file at ~/.core/config.yaml by default.
//
// Keys use dot notation for nested access:
//
//	cfg.Set("dev.editor", "vim")
//	var editor string
//	cfg.Get("dev.editor", &editor)
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	core "github.com/host-uk/core/pkg/framework/core"
	coreio "github.com/host-uk/core/pkg/io"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config implements the core.Config interface with layered resolution.
// It uses viper as the underlying configuration engine.
type Config struct {
	mu     sync.RWMutex
	v      *viper.Viper
	medium coreio.Medium
	path   string
}

// Option is a functional option for configuring a Config instance.
type Option func(*Config)

// WithMedium sets the storage medium for configuration file operations.
func WithMedium(m coreio.Medium) Option {
	return func(c *Config) {
		c.medium = m
	}
}

// WithPath sets the path to the configuration file.
func WithPath(path string) Option {
	return func(c *Config) {
		c.path = path
	}
}

// WithEnvPrefix sets the prefix for environment variables.
func WithEnvPrefix(prefix string) Option {
	return func(c *Config) {
		c.v.SetEnvPrefix(prefix)
	}
}

// New creates a new Config instance with the given options.
// If no medium is provided, it defaults to io.Local.
// If no path is provided, it defaults to ~/.core/config.yaml.
func New(opts ...Option) (*Config, error) {
	c := &Config{
		v: viper.New(),
	}

	// Configure viper defaults
	c.v.SetEnvPrefix("CORE_CONFIG")
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	for _, opt := range opts {
		opt(c)
	}

	if c.medium == nil {
		c.medium = coreio.Local
	}

	if c.path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, core.E("config.New", "failed to determine home directory", err)
		}
		c.path = filepath.Join(home, ".core", "config.yaml")
	}

	c.v.AutomaticEnv()

	// Load existing config file if it exists
	if c.medium.Exists(c.path) {
		if err := c.LoadFile(c.medium, c.path); err != nil {
			return nil, core.E("config.New", "failed to load config file", err)
		}
	}

	return c, nil
}

// LoadFile reads a configuration file from the given medium and path and merges it into the current config.
// It supports YAML and environment files (.env).
func (c *Config) LoadFile(m coreio.Medium, path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	content, err := m.Read(path)
	if err != nil {
		return core.E("config.LoadFile", "failed to read config file: "+path, err)
	}

	ext := filepath.Ext(path)
	if ext == "" && filepath.Base(path) == ".env" {
		c.v.SetConfigType("env")
	} else if ext != "" {
		c.v.SetConfigType(strings.TrimPrefix(ext, "."))
	} else {
		c.v.SetConfigType("yaml")
	}

	if err := c.v.MergeConfig(strings.NewReader(content)); err != nil {
		return core.E("config.LoadFile", "failed to parse config file: "+path, err)
	}

	return nil
}

// Get retrieves a configuration value by dot-notation key and stores it in out.
// If key is empty, it unmarshals the entire configuration into out.
// The out parameter must be a pointer to the target type.
func (c *Config) Get(key string, out any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if key == "" {
		return c.v.Unmarshal(out)
	}

	if !c.v.IsSet(key) {
		return core.E("config.Get", fmt.Sprintf("key not found: %s", key), nil)
	}

	return c.v.UnmarshalKey(key, out)
}

// Set stores a configuration value by dot-notation key and persists to disk.
func (c *Config) Set(key string, v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.v.Set(key, v)

	// Persist to disk
	if err := Save(c.medium, c.path, c.v.AllSettings()); err != nil {
		return core.E("config.Set", "failed to save config", err)
	}

	return nil
}

// All returns a deep copy of all configuration values.
func (c *Config) All() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.v.AllSettings()
}

// Path returns the path to the configuration file.
func (c *Config) Path() string {
	return c.path
}

// Load reads a YAML configuration file from the given medium and path.
// Returns the parsed data as a map, or an error if the file cannot be read or parsed.
// Deprecated: Use Config.LoadFile instead.
func Load(m coreio.Medium, path string) (map[string]any, error) {
	content, err := m.Read(path)
	if err != nil {
		return nil, core.E("config.Load", "failed to read config file: "+path, err)
	}

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(content)); err != nil {
		return nil, core.E("config.Load", "failed to parse config file: "+path, err)
	}

	return v.AllSettings(), nil
}

// Save writes configuration data to a YAML file at the given path.
// It ensures the parent directory exists before writing.
func Save(m coreio.Medium, path string, data map[string]any) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return core.E("config.Save", "failed to marshal config", err)
	}

	dir := filepath.Dir(path)
	if err := m.EnsureDir(dir); err != nil {
		return core.E("config.Save", "failed to create config directory: "+dir, err)
	}

	if err := m.Write(path, string(out)); err != nil {
		return core.E("config.Save", "failed to write config file: "+path, err)
	}

	return nil
}

// Ensure Config implements core.Config at compile time.
var _ core.Config = (*Config)(nil)
