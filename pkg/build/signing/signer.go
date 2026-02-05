// Package signing provides code signing for build artifacts.
package signing

import (
	"context"
	"os"
	"strings"

	"github.com/host-uk/core/pkg/io"
)

// Signer defines the interface for code signing implementations.
type Signer interface {
	// Name returns the signer's identifier.
	Name() string
	// Available checks if this signer can be used.
	Available() bool
	// Sign signs the artifact at the given path.
	Sign(ctx context.Context, fs io.Medium, path string) error
}

// SignConfig holds signing configuration from .core/build.yaml.
type SignConfig struct {
	Enabled bool          `yaml:"enabled"`
	GPG     GPGConfig     `yaml:"gpg,omitempty"`
	MacOS   MacOSConfig   `yaml:"macos,omitempty"`
	Windows WindowsConfig `yaml:"windows,omitempty"`
}

// GPGConfig holds GPG signing configuration.
type GPGConfig struct {
	Key string `yaml:"key"` // Key ID or fingerprint, supports $ENV
}

// MacOSConfig holds macOS codesign configuration.
type MacOSConfig struct {
	Identity    string `yaml:"identity"`     // Developer ID Application: ...
	Notarize    bool   `yaml:"notarize"`     // Submit to Apple for notarization
	AppleID     string `yaml:"apple_id"`     // Apple account email
	TeamID      string `yaml:"team_id"`      // Team ID
	AppPassword string `yaml:"app_password"` // App-specific password
}

// WindowsConfig holds Windows signtool configuration (placeholder).
type WindowsConfig struct {
	Certificate string `yaml:"certificate"` // Path to .pfx
	Password    string `yaml:"password"`    // Certificate password
}

// DefaultSignConfig returns sensible defaults.
func DefaultSignConfig() SignConfig {
	return SignConfig{
		Enabled: true,
		GPG: GPGConfig{
			Key: os.Getenv("GPG_KEY_ID"),
		},
		MacOS: MacOSConfig{
			Identity:    os.Getenv("CODESIGN_IDENTITY"),
			AppleID:     os.Getenv("APPLE_ID"),
			TeamID:      os.Getenv("APPLE_TEAM_ID"),
			AppPassword: os.Getenv("APPLE_APP_PASSWORD"),
		},
	}
}

// ExpandEnv expands environment variables in config values.
func (c *SignConfig) ExpandEnv() {
	c.GPG.Key = expandEnv(c.GPG.Key)
	c.MacOS.Identity = expandEnv(c.MacOS.Identity)
	c.MacOS.AppleID = expandEnv(c.MacOS.AppleID)
	c.MacOS.TeamID = expandEnv(c.MacOS.TeamID)
	c.MacOS.AppPassword = expandEnv(c.MacOS.AppPassword)
	c.Windows.Certificate = expandEnv(c.Windows.Certificate)
	c.Windows.Password = expandEnv(c.Windows.Password)
}

// expandEnv expands $VAR or ${VAR} in a string.
func expandEnv(s string) string {
	if strings.HasPrefix(s, "$") {
		return os.ExpandEnv(s)
	}
	return s
}
