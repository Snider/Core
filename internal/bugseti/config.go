// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigService manages application configuration and persistence.
type ConfigService struct {
	config *Config
	path   string
	mu     sync.RWMutex
}

// Config holds all BugSETI configuration.
type Config struct {
	// Authentication
	GitHubToken string `json:"githubToken,omitempty"`

	// Repositories
	WatchedRepos []string `json:"watchedRepos"`
	Labels       []string `json:"labels"`

	// Scheduling
	WorkHours     *WorkHours `json:"workHours,omitempty"`
	FetchInterval int        `json:"fetchIntervalMinutes"`

	// Notifications
	NotificationsEnabled bool `json:"notificationsEnabled"`
	NotificationSound    bool `json:"notificationSound"`

	// Workspace
	WorkspaceDir string `json:"workspaceDir,omitempty"`
	DataDir      string `json:"dataDir,omitempty"`

	// Onboarding
	Onboarded   bool      `json:"onboarded"`
	OnboardedAt time.Time `json:"onboardedAt,omitempty"`

	// UI Preferences
	Theme         string `json:"theme"`
	ShowTrayPanel bool   `json:"showTrayPanel"`

	// Advanced
	MaxConcurrentIssues int  `json:"maxConcurrentIssues"`
	AutoSeedContext     bool `json:"autoSeedContext"`

	// Updates
	UpdateChannel       string    `json:"updateChannel"`       // stable, beta, nightly
	AutoUpdate          bool      `json:"autoUpdate"`          // Automatically install updates
	UpdateCheckInterval int       `json:"updateCheckInterval"` // Check interval in hours (0 = disabled)
	LastUpdateCheck     time.Time `json:"lastUpdateCheck,omitempty"`
}

// WorkHours defines when BugSETI should actively fetch issues.
type WorkHours struct {
	Enabled   bool   `json:"enabled"`
	StartHour int    `json:"startHour"` // 0-23
	EndHour   int    `json:"endHour"`   // 0-23
	Days      []int  `json:"days"`      // 0=Sunday, 6=Saturday
	Timezone  string `json:"timezone"`
}

// NewConfigService creates a new ConfigService with default values.
func NewConfigService() *ConfigService {
	// Determine config path
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}

	bugsetiDir := filepath.Join(configDir, "bugseti")
	if err := os.MkdirAll(bugsetiDir, 0755); err != nil {
		log.Printf("Warning: could not create config directory: %v", err)
	}

	return &ConfigService{
		path: filepath.Join(bugsetiDir, "config.json"),
		config: &Config{
			WatchedRepos: []string{},
			Labels: []string{
				"good first issue",
				"help wanted",
				"beginner-friendly",
			},
			FetchInterval:        15,
			NotificationsEnabled: true,
			NotificationSound:    true,
			Theme:                "dark",
			ShowTrayPanel:        true,
			MaxConcurrentIssues:  1,
			AutoSeedContext:      true,
			DataDir:              bugsetiDir,
			UpdateChannel:        "stable",
			AutoUpdate:           false,
			UpdateCheckInterval:  6, // Check every 6 hours
		},
	}
}

// ServiceName returns the service name for Wails.
func (c *ConfigService) ServiceName() string {
	return "ConfigService"
}

// Load reads the configuration from disk.
func (c *ConfigService) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file yet, use defaults
			return c.saveUnsafe()
		}
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	// Merge with defaults for any new fields
	c.mergeDefaults(&config)
	c.config = &config
	return nil
}

// Save persists the configuration to disk.
func (c *ConfigService) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.saveUnsafe()
}

// saveUnsafe writes config without acquiring lock.
func (c *ConfigService) saveUnsafe() error {
	data, err := json.MarshalIndent(c.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}

// mergeDefaults fills in default values for any unset fields.
func (c *ConfigService) mergeDefaults(config *Config) {
	if config.Labels == nil || len(config.Labels) == 0 {
		config.Labels = c.config.Labels
	}
	if config.FetchInterval == 0 {
		config.FetchInterval = 15
	}
	if config.Theme == "" {
		config.Theme = "dark"
	}
	if config.MaxConcurrentIssues == 0 {
		config.MaxConcurrentIssues = 1
	}
	if config.DataDir == "" {
		config.DataDir = c.config.DataDir
	}
	if config.UpdateChannel == "" {
		config.UpdateChannel = "stable"
	}
	if config.UpdateCheckInterval == 0 {
		config.UpdateCheckInterval = 6
	}
}

// GetConfig returns a copy of the current configuration.
func (c *ConfigService) GetConfig() Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return *c.config
}

// SetConfig updates the configuration and saves it.
func (c *ConfigService) SetConfig(config Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = &config
	return c.saveUnsafe()
}

// GetWatchedRepos returns the list of watched repositories.
func (c *ConfigService) GetWatchedRepos() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.WatchedRepos
}

// AddWatchedRepo adds a repository to the watch list.
func (c *ConfigService) AddWatchedRepo(repo string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, r := range c.config.WatchedRepos {
		if r == repo {
			return nil // Already watching
		}
	}

	c.config.WatchedRepos = append(c.config.WatchedRepos, repo)
	return c.saveUnsafe()
}

// RemoveWatchedRepo removes a repository from the watch list.
func (c *ConfigService) RemoveWatchedRepo(repo string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, r := range c.config.WatchedRepos {
		if r == repo {
			c.config.WatchedRepos = append(c.config.WatchedRepos[:i], c.config.WatchedRepos[i+1:]...)
			return c.saveUnsafe()
		}
	}

	return nil
}

// GetLabels returns the issue labels to filter by.
func (c *ConfigService) GetLabels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.Labels
}

// SetLabels updates the issue labels.
func (c *ConfigService) SetLabels(labels []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.Labels = labels
	return c.saveUnsafe()
}

// GetFetchInterval returns the fetch interval as a duration.
func (c *ConfigService) GetFetchInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Duration(c.config.FetchInterval) * time.Minute
}

// SetFetchInterval sets the fetch interval in minutes.
func (c *ConfigService) SetFetchInterval(minutes int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.FetchInterval = minutes
	return c.saveUnsafe()
}

// IsWithinWorkHours checks if the current time is within configured work hours.
func (c *ConfigService) IsWithinWorkHours() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.config.WorkHours == nil || !c.config.WorkHours.Enabled {
		return true // No work hours restriction
	}

	wh := c.config.WorkHours
	now := time.Now()

	// Check timezone
	if wh.Timezone != "" {
		loc, err := time.LoadLocation(wh.Timezone)
		if err == nil {
			now = now.In(loc)
		}
	}

	// Check day
	day := int(now.Weekday())
	dayAllowed := false
	for _, d := range wh.Days {
		if d == day {
			dayAllowed = true
			break
		}
	}
	if !dayAllowed {
		return false
	}

	// Check hour
	hour := now.Hour()
	if wh.StartHour <= wh.EndHour {
		return hour >= wh.StartHour && hour < wh.EndHour
	}
	// Handle overnight (e.g., 22:00 - 06:00)
	return hour >= wh.StartHour || hour < wh.EndHour
}

// GetWorkHours returns the work hours configuration.
func (c *ConfigService) GetWorkHours() *WorkHours {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.WorkHours
}

// SetWorkHours updates the work hours configuration.
func (c *ConfigService) SetWorkHours(wh *WorkHours) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.WorkHours = wh
	return c.saveUnsafe()
}

// IsNotificationsEnabled returns whether notifications are enabled.
func (c *ConfigService) IsNotificationsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.NotificationsEnabled
}

// SetNotificationsEnabled enables or disables notifications.
func (c *ConfigService) SetNotificationsEnabled(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.NotificationsEnabled = enabled
	return c.saveUnsafe()
}

// GetWorkspaceDir returns the workspace directory.
func (c *ConfigService) GetWorkspaceDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.WorkspaceDir
}

// SetWorkspaceDir sets the workspace directory.
func (c *ConfigService) SetWorkspaceDir(dir string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.WorkspaceDir = dir
	return c.saveUnsafe()
}

// GetDataDir returns the data directory.
func (c *ConfigService) GetDataDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.DataDir
}

// IsOnboarded returns whether the user has completed onboarding.
func (c *ConfigService) IsOnboarded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.Onboarded
}

// CompleteOnboarding marks onboarding as complete.
func (c *ConfigService) CompleteOnboarding() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.Onboarded = true
	c.config.OnboardedAt = time.Now()
	return c.saveUnsafe()
}

// GetTheme returns the current theme.
func (c *ConfigService) GetTheme() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.Theme
}

// SetTheme sets the theme.
func (c *ConfigService) SetTheme(theme string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.Theme = theme
	return c.saveUnsafe()
}

// IsAutoSeedEnabled returns whether automatic context seeding is enabled.
func (c *ConfigService) IsAutoSeedEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.AutoSeedContext
}

// SetAutoSeedEnabled enables or disables automatic context seeding.
func (c *ConfigService) SetAutoSeedEnabled(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.AutoSeedContext = enabled
	return c.saveUnsafe()
}

// UpdateSettings holds update-related configuration.
type UpdateSettings struct {
	Channel       string    `json:"channel"`
	AutoUpdate    bool      `json:"autoUpdate"`
	CheckInterval int       `json:"checkInterval"` // Hours
	LastCheck     time.Time `json:"lastCheck"`
}

// GetUpdateSettings returns the update settings.
func (c *ConfigService) GetUpdateSettings() UpdateSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return UpdateSettings{
		Channel:       c.config.UpdateChannel,
		AutoUpdate:    c.config.AutoUpdate,
		CheckInterval: c.config.UpdateCheckInterval,
		LastCheck:     c.config.LastUpdateCheck,
	}
}

// SetUpdateSettings updates the update settings.
func (c *ConfigService) SetUpdateSettings(settings UpdateSettings) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.UpdateChannel = settings.Channel
	c.config.AutoUpdate = settings.AutoUpdate
	c.config.UpdateCheckInterval = settings.CheckInterval
	return c.saveUnsafe()
}

// GetUpdateChannel returns the update channel.
func (c *ConfigService) GetUpdateChannel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.UpdateChannel
}

// SetUpdateChannel sets the update channel.
func (c *ConfigService) SetUpdateChannel(channel string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.UpdateChannel = channel
	return c.saveUnsafe()
}

// IsAutoUpdateEnabled returns whether automatic updates are enabled.
func (c *ConfigService) IsAutoUpdateEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.AutoUpdate
}

// SetAutoUpdateEnabled enables or disables automatic updates.
func (c *ConfigService) SetAutoUpdateEnabled(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.AutoUpdate = enabled
	return c.saveUnsafe()
}

// GetUpdateCheckInterval returns the update check interval in hours.
func (c *ConfigService) GetUpdateCheckInterval() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.UpdateCheckInterval
}

// SetUpdateCheckInterval sets the update check interval in hours.
func (c *ConfigService) SetUpdateCheckInterval(hours int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.UpdateCheckInterval = hours
	return c.saveUnsafe()
}

// GetLastUpdateCheck returns the last update check time.
func (c *ConfigService) GetLastUpdateCheck() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.LastUpdateCheck
}

// SetLastUpdateCheck sets the last update check time.
func (c *ConfigService) SetLastUpdateCheck(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.LastUpdateCheck = t
	return c.saveUnsafe()
}

// ShouldCheckForUpdates returns true if it's time to check for updates.
func (c *ConfigService) ShouldCheckForUpdates() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.config.UpdateCheckInterval <= 0 {
		return false // Updates disabled
	}

	if c.config.LastUpdateCheck.IsZero() {
		return true // Never checked
	}

	interval := time.Duration(c.config.UpdateCheckInterval) * time.Hour
	return time.Since(c.config.LastUpdateCheck) >= interval
}
