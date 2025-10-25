package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Snider/Core"
	"github.com/adrg/xdg"
)

const appName = "lethean"
const configFileName = "config.json"

// ErrSetupRequired is returned if the config file is missing and cannot be created.
var ErrSetupRequired = errors.New("setup required: config.json not found")

// Options holds configuration for the config service.
type Options struct{}

// Service provides access to the application's configuration.
// It handles loading, saving, and providing access to configuration values.
type Service struct {
	*core.Runtime[Options] `json:"-"`

	// Non-persistent fields, derived at runtime.
	ConfigPath    string `json:"-"`
	UserHomeDir   string `json:"-"`
	RootDir       string `json:"-"`
	CacheDir      string `json:"-"`
	ConfigDir     string `json:"-"`
	DataDir       string `json:"-"`
	WorkspacesDir string `json:"-"`

	// Persistent fields, saved to config.json.
	DefaultRoute string   `json:"default_route"`
	Features     []string `json:"features"`
	Language     string   `json:"language"`
}

// New is a factory function that creates and initializes a new configuration service.
// It loads an existing configuration or creates a default one if not found.
func New(c *core.Core) (any, error) {
	// --- Path and Directory Setup ---
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not resolve user home directory: %w", err)
	}
	userHomeDir := filepath.Join(homeDir, appName)

	rootDir, err := xdg.DataFile(appName)
	if err != nil {
		return nil, fmt.Errorf("could not resolve data directory: %w", err)
	}

	cacheDir, err := xdg.CacheFile(appName)
	if err != nil {
		return nil, fmt.Errorf("could not resolve cache directory: %w", err)
	}

	s := &Service{
		Runtime:       core.NewRuntime(c, Options{}),
		UserHomeDir:   userHomeDir,
		RootDir:       rootDir,
		CacheDir:      cacheDir,
		ConfigDir:     filepath.Join(userHomeDir, "config"),
		DataDir:       filepath.Join(userHomeDir, "data"),
		WorkspacesDir: filepath.Join(userHomeDir, "workspaces"),
		DefaultRoute:  "/",
		Features:      []string{},
		Language:      "en",
	}
	s.ConfigPath = filepath.Join(s.ConfigDir, configFileName)

	dirs := []string{s.RootDir, s.ConfigDir, s.DataDir, s.CacheDir, s.WorkspacesDir, s.UserHomeDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("could not create directory %s: %w", dir, err)
		}
	}

	// --- Load or Create Configuration ---
	if data, err := os.ReadFile(s.ConfigPath); err == nil {
		// Config file exists, load it.
		if err := json.Unmarshal(data, s); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	} else if os.IsNotExist(err) {
		// Config file does not exist, create it with default values.
		if err := s.Save(); err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
	} else {
		// Another error occurred reading the file.
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	c.RegisterAction(s.handleIPCEvents)
	return s, nil
}

// handleIPCEvents is the central IPC handler for the config service.
func (s *Service) handleIPCEvents(c *core.Core, msg core.Message) error {
	switch msg.(type) {
	case core.ActionServiceStartup:
		c.App.Logger.Info("Config service started")
	default:
		// No other actions are handled by this service yet.
	}
	return nil
}

// Save writes the current configuration to config.json.
func (s *Service) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(s.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// IsFeatureEnabled checks if a given feature is enabled in the configuration.
func (s *Service) IsFeatureEnabled(feature string) bool {
	for _, f := range s.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// EnableFeature adds a feature to the list of enabled features and saves the config.
func (s *Service) EnableFeature(feature string) error {
	if s.IsFeatureEnabled(feature) {
		return nil
	}
	s.Features = append(s.Features, feature)
	if err := s.Save(); err != nil {
		return fmt.Errorf("failed to save config after enabling feature %s: %w", feature, err)
	}
	return nil
}

func (s *Service) Key(key string) (interface{}, error) {
	// Use reflection to inspect the struct fields.
	val := reflect.ValueOf(s).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name

		// Check the field name first.
		if strings.EqualFold(fieldName, key) {
			return val.Field(i).Interface(), nil
		}

		// Then check the `json` tag.
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			jsonName := strings.Split(jsonTag, ",")[0]
			if strings.EqualFold(jsonName, key) {
				return val.Field(i).Interface(), nil
			}
		}
	}

	return nil, fmt.Errorf("key '%s' not found in config", key)
}
