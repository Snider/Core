package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/Snider/Core"
	"github.com/adrg/xdg"
)

const appName = "lethean"
const configFileName = "config.json"

// ErrSetupRequired is returned by ServiceStartup if config.json is missing.
var ErrSetupRequired = errors.New("setup required: config.json not found")

// Service provides access to the application's configuration.
var instance *API

// NewService creates and initializes a new configuration service.
// It loads an existing configuration or creates a default one if not found.
func Register(c *core.Core) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not resolve user home directory: %w", err)
	}
	userHomeDir := filepath.Join(homeDir, appName)
	configDir := filepath.Join(userHomeDir, "config")
	//configPath := filepath.Join(configDir, configFileName)

	instance = &API{
		core:          c,
		UserHomeDir:   userHomeDir,
		ConfigDir:     configDir,
		DataDir:       filepath.Join(userHomeDir, "data"),
		WorkspacesDir: filepath.Join(userHomeDir, "workspaces"),
		DefaultRoute:  "/",
		Features:      []string{},
		Language:      "en",
	}

	err = c.RegisterModule("config", instance)
	if err != nil {
		return err
	}
	c.RegisterAction(handleActionCall)
	return nil
}
func handleActionCall(c *core.Core, msg core.Message) error {
	switch m := msg.(type) {

	case core.ActionServiceStartup:
		//err := instance.ServiceStartup(context.Background(), application.ServiceOptions{})
		//if err != nil {
		//	return err
		//}
		c.App.Logger.Info("Config service started")
		return nil
	default:
		c.App.Logger.Error("Unknown message type", "type", fmt.Sprintf("%T", m))
		return nil
	}
}

// newDefaultConfig creates a default configuration with resolved paths and ensures directories exist.
func newDefaultConfig() (*API, error) {
	if strings.Contains(appName, "..") || strings.Contains(appName, string(filepath.Separator)) {
		return nil, fmt.Errorf("invalid app name '%s': contains path traversal characters", appName)
	}

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

	instance := &API{
		UserHomeDir:   userHomeDir,
		RootDir:       rootDir,
		CacheDir:      cacheDir,
		ConfigDir:     filepath.Join(userHomeDir, "config"),
		DataDir:       filepath.Join(userHomeDir, "data"),
		WorkspacesDir: filepath.Join(userHomeDir, "workspaces"),
		DefaultRoute:  "/",
		Features:      []string{},
		Language:      "en", // Hardcoded default, will be overridden if loaded or detected
	}

	dirs := []string{instance.RootDir, instance.ConfigDir, instance.DataDir, instance.CacheDir, instance.WorkspacesDir, instance.UserHomeDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("could not create directory %s: %w", dir, err)
		}
	}

	return instance, nil
}

// Save writes the current configuration to config.json.
func (c *API) Save() error {
	configPath := filepath.Join(c.ConfigDir, configFileName)

	data, err := json.MarshalIndent(*c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// IsFeatureEnabled checks if a given feature is enabled in the configuration.
func (c *API) IsFeatureEnabled(feature string) bool {
	for _, f := range c.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// EnableFeature adds a feature to the list of enabled features and saves the config.
func (c *API) EnableFeature(feature string) error {
	if c.IsFeatureEnabled(feature) {
		return nil
	}
	c.Features = append(c.Features, feature)
	if err := c.Save(); err != nil {
		return fmt.Errorf("failed to save config after enabling feature %s: %w", feature, err)
	}
	return nil
}
