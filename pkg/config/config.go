// Package config provides a configuration management service that handles
// loading, saving, and accessing application settings. It supports both a
// main JSON configuration file and auxiliary data stored in various formats
// like YAML, INI, and XML. The service is designed to be extensible and
// can be used with static or dynamic dependency injection.
//
// The Service struct is the core of the package, providing methods to
// interact with the configuration. It manages file paths, default values,
// and the serialization/deserialization of data.
//
// Basic usage involves creating a new Service instance and then using its
// methods to get, set, and manage configuration data.
//
// Example:
//
//	// Create a new config service.
//	cfg, err := config.New()
//	if err != nil {
//		log.Fatalf("failed to create config service: %v", err)
//	}
//
//	// Set a new value.
//	err = cfg.Set("language", "fr")
//	if err != nil {
//		log.Fatalf("failed to set config value: %v", err)
//	}
//
//	// Retrieve a value.
//	var lang string
//	err = cfg.Get("language", &lang)
//	if err != nil {
//		log.Fatalf("failed to get config value: %v", err)
//	}
//	fmt.Printf("Language: %s\n", lang)
package config

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

// HandleIPCEvents processes IPC messages for the config service.
func (s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error {
	switch msg.(type) {
	case core.ActionServiceStartup:
		// Config initializes during Register(), no additional startup needed.
		return nil
	default:
		if c.App != nil && c.App.Logger != nil {
			c.App.Logger.Debug("Config: Unhandled message type", "type", fmt.Sprintf("%T", msg))
		}
	}
	return nil
}

const appName = "lethean"
const configFileName = "config.json"

// Options holds configuration for the config service. This struct is provided
// for future extensibility and currently has no fields.
type Options struct{}

// Service provides access to the application's configuration.
// It handles loading, saving, and providing access to configuration values,
// abstracting away the details of file I/O and data serialization.
// The Service is designed to be a central point for all configuration-related
// operations within the application.
//
// The fields of the Service struct are automatically saved to and loaded from
// a JSON configuration file. The `json:"-"` tag on ServiceRuntime prevents
// it from being serialized.
type Service struct {
	*core.ServiceRuntime[Options] `json:"-"`

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

// createServiceInstance handles the setup of the configuration service. It
// resolves necessary paths, creates directories, and loads the configuration
// file if it exists. If the configuration file is not found, it creates a new
// one with default values. This function is not exported and is used internally
// by the New and Register constructors.
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

// New creates a new instance of the configuration service. This constructor is
// intended for static dependency injection, where the service is created and
// managed manually. It initializes the service with default paths and values,
// and loads any existing configuration from disk.
//
// Example:
//
//	cfg, err := config.New()
//	if err != nil {
//		log.Fatalf("Failed to initialize config: %v", err)
//	}
//	// Use cfg to access configuration settings.
func New() (*Service, error) {
	return createServiceInstance()
}

// Register creates a new instance of the configuration service and registers it
// with the application's core. This constructor is intended for dynamic
// dependency injection, where services are managed by a central core component.
// It performs the same initialization as New, but also integrates the service
// with the provided core instance.
func Register(c *core.Core) (any, error) {
	s, err := createServiceInstance()
	if err != nil {
		return nil, err
	}
	// Defensive check: createServiceInstance should not return nil service with nil error
	if s == nil {
		return nil, errors.New("config: createServiceInstance returned a nil service instance with no error")
	}
	s.ServiceRuntime = core.NewServiceRuntime(c, Options{})
	return s, nil
}

// Save writes the current configuration to a JSON file. The location of the file
// is determined by the ConfigPath field of the Service struct. This method is
// typically called automatically by Set, but can be used to explicitly save
// changes.
//
// Example:
//
//	err := cfg.Save()
//	if err != nil {
//		log.Printf("Error saving configuration: %v", err)
//	}
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

// Get retrieves a configuration value by its key. The key corresponds to the
// JSON tag of a field in the Service struct. The retrieved value is stored in
// the `out` parameter, which must be a non-nil pointer to a variable of the
// correct type.
//
// Example:
//
//	var currentLanguage string
//	err := cfg.Get("language", &currentLanguage)
//	if err != nil {
//		log.Printf("Could not retrieve language setting: %v", err)
//	}
//	fmt.Println("Current language is:", currentLanguage)
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

// SaveStruct saves an arbitrary struct to a JSON file in the config directory.
// This is useful for storing complex data that is not part of the main
// configuration. The `key` parameter is used as the filename (with a .json
// extension).
//
// Example:
//
//	type UserPreferences struct {
//		Theme string `json:"theme"`
//		Notifications bool `json:"notifications"`
//	}
//	prefs := UserPreferences{Theme: "dark", Notifications: true}
//	err := cfg.SaveStruct("user_prefs", prefs)
//	if err != nil {
//		log.Printf("Error saving user preferences: %v", err)
//	}
func (s *Service) SaveStruct(key string, data interface{}) error {
	filePath := filepath.Join(s.ConfigDir, key+".json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal struct for key '%s': %w", key, err)
	}
	return os.WriteFile(filePath, jsonData, 0644)
}

// LoadStruct loads an arbitrary struct from a JSON file in the config directory.
// The `key` parameter specifies the filename (without the .json extension). The
// loaded data is unmarshaled into the `data` parameter, which must be a
// non-nil pointer to a struct.
//
// Example:
//
//	var prefs UserPreferences
//	err := cfg.LoadStruct("user_prefs", &prefs)
//	if err != nil {
//		log.Printf("Error loading user preferences: %v", err)
//	}
//	fmt.Printf("User theme is: %s", prefs.Theme)
func (s *Service) LoadStruct(key string, data interface{}) error {
	filePath := filepath.Join(s.ConfigDir, key+".json")
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Return nil if the file doesn't exist
		}
		return fmt.Errorf("failed to read struct file for key '%s': %w", key, err)
	}
	return json.Unmarshal(jsonData, data)
}

// Set updates a configuration value and saves the change to the configuration
// file. The key corresponds to the JSON tag of a field in the Service struct.
// The provided value `v` must be of a type that is assignable to the field.
//
// Example:
//
//	err := cfg.Set("default_route", "/home")
//	if err != nil {
//		log.Printf("Failed to set default route: %v", err)
//	}
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
