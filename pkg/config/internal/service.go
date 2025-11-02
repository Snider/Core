package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Snider/Core/pkg/core"
	"github.com/adrg/xdg"
)

const appName = "lethean"
const configFileName = "config.json"

// Options holds configuration for the config service.
type Options struct{}

// Service provides access to the application's configuration.
// It handles loading, saving, and providing access to configuration values.
type Service struct {
	*core.Runtime[Options] `json:"-"`

	// Persistent fields, saved to config.json.
	ConfigPath   string   `json:"configPath,omitempty"`
	UserHomeDir  string   `json:"userHomeDir,omitempty"`
	RootDir      string   `json:"rootDir,omitempty"`
	CacheDir     string   `json:"cacheDir,omitempty"`
	ConfigDir    string   `json:"configDir,omitempty"`
	DataDir      string   `json:"dataDir,omitempty"`
	WorkspaceDir string   `json:"workspaceDir,omitempty"`
	DefaultRoute string   `json:"default_route"`
	Features     []string `json:"features"`
	Language     string   `json:"language"`
}

// createServiceInstance contains the common logic for initializing a Service struct.
func createServiceInstance() (*Service, error) {
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
		UserHomeDir:  userHomeDir,
		RootDir:      rootDir,
		CacheDir:     cacheDir,
		ConfigDir:    filepath.Join(userHomeDir, "config"),
		DataDir:      filepath.Join(userHomeDir, "data"),
		WorkspaceDir: filepath.Join(userHomeDir, "workspace"),
		DefaultRoute: "/",
		Features:     []string{},
		Language:     "en",
	}
	s.ConfigPath = filepath.Join(s.ConfigDir, configFileName)

	dirs := []string{s.RootDir, s.ConfigDir, s.DataDir, s.CacheDir, s.WorkspaceDir, s.UserHomeDir}
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

	return s, nil
}

// New is the constructor for static dependency injection.
// It creates a Service instance without initializing the core.Runtime field.
func New() (*Service, error) {
	return createServiceInstance()
}

// Register is the constructor for dynamic dependency injection (used with core.WithService).
// It creates a Service instance and initializes its core.Runtime field.
func Register(c *core.Core) (any, error) {
	s, err := createServiceInstance()
	if err != nil {
		return nil, err
	}
	// Defensive check: createServiceInstance should not return nil service with nil error
	if s == nil {
		return nil, errors.New("config: createServiceInstance returned a nil service instance with no error")
	}
	s.Runtime = core.NewRuntime(c, Options{})
	return s, nil
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

// Get retrieves a configuration value by its key.
func (s *Service) Get(key string, out any) error {
	val := reflect.ValueOf(s).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			jsonName := strings.Split(jsonTag, ",")[0]
			if strings.EqualFold(jsonName, key) {
				outVal := reflect.ValueOf(out)
				if outVal.Kind() != reflect.Ptr || outVal.IsNil() {
					return errors.New("output argument must be a non-nil pointer")
				}
				targetVal := outVal.Elem()
				srcVal := val.Field(i)

				if !srcVal.Type().AssignableTo(targetVal.Type()) {
					return fmt.Errorf("cannot assign config value of type %s to output of type %s", srcVal.Type(), targetVal.Type())
				}
				targetVal.Set(srcVal)
				return nil
			}
		}
	}

	return fmt.Errorf("key '%s' not found in config", key)
}

// IsFeatureEnabled checks if a specific feature is enabled in the config.
func (s *Service) IsFeatureEnabled(feature string) bool {
	for _, f := range s.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// Set updates a configuration value and saves the config.
func (s *Service) Set(key string, v any) error {
	val := reflect.ValueOf(s).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			jsonName := strings.Split(jsonTag, ",")[0]
			if strings.EqualFold(jsonName, key) {
				fieldVal := val.Field(i)
				if !fieldVal.CanSet() {
					return fmt.Errorf("cannot set config field for key '%s'", key)
				}
				newVal := reflect.ValueOf(v)
				if !newVal.Type().AssignableTo(fieldVal.Type()) {
					return fmt.Errorf("type mismatch for key '%s': expected %s, got %s", key, fieldVal.Type(), newVal.Type())
				}
				fieldVal.Set(newVal)
				return s.Save()
			}
		}
	}

	return fmt.Errorf("key '%s' not found in config", key)
}
