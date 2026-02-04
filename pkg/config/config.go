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
	"strings"
	"sync"

	core "github.com/host-uk/core/pkg/framework/core"
	"github.com/host-uk/core/pkg/io"
)

// Config implements the core.Config interface with layered resolution.
// Values are resolved in order: defaults -> file -> env -> flags.
type Config struct {
	mu     sync.RWMutex
	medium io.Medium
	path   string
	data   map[string]any
}

// Option is a functional option for configuring a Config instance.
type Option func(*Config)

// WithMedium sets the storage medium for configuration file operations.
func WithMedium(m io.Medium) Option {
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

// New creates a new Config instance with the given options.
// If no medium is provided, it defaults to io.Local.
// If no path is provided, it defaults to ~/.core/config.yaml.
func New(opts ...Option) (*Config, error) {
	c := &Config{
		data: make(map[string]any),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.medium == nil {
		c.medium = io.Local
	}

	if c.path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, core.E("config.New", "failed to determine home directory", err)
		}
		c.path = home + "/.core/config.yaml"
	}

	// Load existing config file if it exists
	if c.medium.IsFile(c.path) {
		loaded, err := Load(c.medium, c.path)
		if err != nil {
			return nil, core.E("config.New", "failed to load config file", err)
		}
		c.data = loaded
	}

	// Overlay environment variables
	envData := LoadEnv("CORE_CONFIG_")
	for k, v := range envData {
		setNested(c.data, k, v)
	}

	return c, nil
}

// Get retrieves a configuration value by dot-notation key and stores it in out.
// The out parameter must be a pointer to the target type.
// Returns an error if the key is not found.
func (c *Config) Get(key string, out any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := getNested(c.data, key)
	if !ok {
		return core.E("config.Get", fmt.Sprintf("key not found: %s", key), nil)
	}

	return assign(val, out)
}

// Set stores a configuration value by dot-notation key and persists to disk.
func (c *Config) Set(key string, v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	setNested(c.data, key, v)

	if err := Save(c.medium, c.path, c.data); err != nil {
		return core.E("config.Set", "failed to save config", err)
	}

	return nil
}

// All returns a deep copy of all configuration values.
func (c *Config) All() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return deepCopyMap(c.data)
}

// deepCopyMap recursively copies a map[string]any.
func deepCopyMap(src map[string]any) map[string]any {
	result := make(map[string]any, len(src))
	for k, v := range src {
		switch val := v.(type) {
		case map[string]any:
			result[k] = deepCopyMap(val)
		case []any:
			cp := make([]any, len(val))
			copy(cp, val)
			result[k] = cp
		default:
			result[k] = v
		}
	}
	return result
}

// Path returns the path to the configuration file.
func (c *Config) Path() string {
	return c.path
}

// getNested retrieves a value from a nested map using dot-notation keys.
func getNested(data map[string]any, key string) (any, bool) {
	parts := strings.Split(key, ".")
	current := any(data)

	for i, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		val, exists := m[part]
		if !exists {
			return nil, false
		}
		if i == len(parts)-1 {
			return val, true
		}
		current = val
	}

	return nil, false
}

// setNested sets a value in a nested map using dot-notation keys,
// creating intermediate maps as needed.
func setNested(data map[string]any, key string, value any) {
	parts := strings.Split(key, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}
		next, ok := current[part]
		if !ok {
			next = make(map[string]any)
			current[part] = next
		}
		m, ok := next.(map[string]any)
		if !ok {
			m = make(map[string]any)
			current[part] = m
		}
		current = m
	}
}

// assign sets the value of out to val, handling type conversions.
func assign(val any, out any) error {
	switch ptr := out.(type) {
	case *string:
		switch v := val.(type) {
		case string:
			*ptr = v
		default:
			*ptr = fmt.Sprintf("%v", v)
		}
	case *int:
		switch v := val.(type) {
		case int:
			*ptr = v
		case float64:
			*ptr = int(v)
		case int64:
			*ptr = int(v)
		default:
			return core.E("config.assign", fmt.Sprintf("cannot assign %T to *int", val), nil)
		}
	case *bool:
		switch v := val.(type) {
		case bool:
			*ptr = v
		default:
			return core.E("config.assign", fmt.Sprintf("cannot assign %T to *bool", val), nil)
		}
	case *float64:
		switch v := val.(type) {
		case float64:
			*ptr = v
		case int:
			*ptr = float64(v)
		case int64:
			*ptr = float64(v)
		default:
			return core.E("config.assign", fmt.Sprintf("cannot assign %T to *float64", val), nil)
		}
	case *any:
		*ptr = val
	case *map[string]any:
		switch v := val.(type) {
		case map[string]any:
			*ptr = v
		default:
			return core.E("config.assign", fmt.Sprintf("cannot assign %T to *map[string]any", val), nil)
		}
	default:
		return core.E("config.assign", fmt.Sprintf("unsupported target type: %T", out), nil)
	}
	return nil
}

// Ensure Config implements core.Config at compile time.
var _ core.Config = (*Config)(nil)
