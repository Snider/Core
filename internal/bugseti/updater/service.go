// Package updater provides auto-update functionality for BugSETI.
package updater

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/host-uk/core/internal/bugseti"
)

// Service provides update functionality and Wails bindings.
type Service struct {
	config     *bugseti.ConfigService
	checker    *Checker
	downloader *Downloader
	installer  *Installer

	mu            sync.RWMutex
	lastResult    *UpdateCheckResult
	pendingUpdate *DownloadResult

	// Background check
	stopCh  chan struct{}
	running bool
}

// NewService creates a new update service.
func NewService(config *bugseti.ConfigService) (*Service, error) {
	downloader, err := NewDownloader()
	if err != nil {
		return nil, err
	}

	installer, err := NewInstaller()
	if err != nil {
		return nil, err
	}

	return &Service{
		config:     config,
		checker:    NewChecker(),
		downloader: downloader,
		installer:  installer,
	}, nil
}

// ServiceName returns the service name for Wails.
func (s *Service) ServiceName() string {
	return "UpdateService"
}

// Start begins the background update checker.
func (s *Service) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	go s.runBackgroundChecker()
}

// Stop stops the background update checker.
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopCh)
}

// runBackgroundChecker runs periodic update checks.
func (s *Service) runBackgroundChecker() {
	// Initial check after a short delay
	time.Sleep(30 * time.Second)

	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		if s.config.ShouldCheckForUpdates() {
			log.Println("Checking for updates...")
			_, err := s.CheckForUpdate()
			if err != nil {
				log.Printf("Update check failed: %v", err)
			}
		}

		// Check interval from config (minimum 1 hour)
		interval := time.Duration(s.config.GetUpdateCheckInterval()) * time.Hour
		if interval < time.Hour {
			interval = time.Hour
		}

		select {
		case <-s.stopCh:
			return
		case <-time.After(interval):
		}
	}
}

// GetSettings returns the update settings.
func (s *Service) GetSettings() bugseti.UpdateSettings {
	return s.config.GetUpdateSettings()
}

// SetSettings updates the update settings.
func (s *Service) SetSettings(settings bugseti.UpdateSettings) error {
	return s.config.SetUpdateSettings(settings)
}

// GetVersionInfo returns the current version information.
func (s *Service) GetVersionInfo() bugseti.VersionInfo {
	return bugseti.GetVersionInfo()
}

// GetChannels returns all available update channels.
func (s *Service) GetChannels() []ChannelInfo {
	return GetAllChannelInfo()
}

// CheckForUpdate checks if an update is available.
func (s *Service) CheckForUpdate() (*UpdateCheckResult, error) {
	currentVersion := bugseti.GetVersion()
	channelStr := s.config.GetUpdateChannel()

	channel, err := ParseChannel(channelStr)
	if err != nil {
		channel = ChannelStable
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.checker.CheckForUpdate(ctx, currentVersion, channel)
	if err != nil {
		return result, err
	}

	// Update last check time
	s.config.SetLastUpdateCheck(time.Now())

	// Store result
	s.mu.Lock()
	s.lastResult = result
	s.mu.Unlock()

	// If auto-update is enabled and an update is available, download it
	if result.Available && s.config.IsAutoUpdateEnabled() {
		go s.downloadUpdate(result.Release)
	}

	return result, nil
}

// GetLastCheckResult returns the last update check result.
func (s *Service) GetLastCheckResult() *UpdateCheckResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastResult
}

// downloadUpdate downloads an update in the background.
func (s *Service) downloadUpdate(release *ReleaseInfo) {
	if release == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	log.Printf("Downloading update %s...", release.Version)

	result, err := s.downloader.Download(ctx, release)
	if err != nil {
		log.Printf("Failed to download update: %v", err)
		return
	}

	log.Printf("Update %s downloaded and staged at %s", release.Version, result.BinaryPath)

	s.mu.Lock()
	s.pendingUpdate = result
	s.mu.Unlock()
}

// DownloadUpdate downloads the latest available update.
func (s *Service) DownloadUpdate() (*DownloadResult, error) {
	s.mu.RLock()
	lastResult := s.lastResult
	s.mu.RUnlock()

	if lastResult == nil || !lastResult.Available || lastResult.Release == nil {
		// Need to check first
		result, err := s.CheckForUpdate()
		if err != nil {
			return nil, err
		}
		if !result.Available {
			return nil, nil
		}
		lastResult = result
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	downloadResult, err := s.downloader.Download(ctx, lastResult.Release)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.pendingUpdate = downloadResult
	s.mu.Unlock()

	return downloadResult, nil
}

// InstallUpdate installs a previously downloaded update.
func (s *Service) InstallUpdate() (*InstallResult, error) {
	s.mu.RLock()
	pending := s.pendingUpdate
	s.mu.RUnlock()

	if pending == nil {
		// Try to download first
		downloadResult, err := s.DownloadUpdate()
		if err != nil {
			return nil, err
		}
		if downloadResult == nil {
			return &InstallResult{
				Success: false,
				Error:   "No update available",
			}, nil
		}
		pending = downloadResult
	}

	result, err := s.installer.Install(pending.BinaryPath)
	if err != nil {
		return result, err
	}

	// Clear pending update
	s.mu.Lock()
	s.pendingUpdate = nil
	s.mu.Unlock()

	return result, nil
}

// InstallAndRestart installs the update and restarts the application.
func (s *Service) InstallAndRestart() error {
	result, err := s.InstallUpdate()
	if err != nil {
		return err
	}

	if !result.Success {
		return nil
	}

	return s.installer.Restart()
}

// HasPendingUpdate returns true if there's a downloaded update ready to install.
func (s *Service) HasPendingUpdate() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingUpdate != nil
}

// GetPendingUpdate returns information about the pending update.
func (s *Service) GetPendingUpdate() *DownloadResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingUpdate
}

// CancelPendingUpdate cancels and removes the pending update.
func (s *Service) CancelPendingUpdate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pendingUpdate = nil
	return s.downloader.Cleanup()
}

// CanSelfUpdate returns true if the application can update itself.
func (s *Service) CanSelfUpdate() bool {
	return CanSelfUpdate()
}

// NeedsElevation returns true if the update requires elevated privileges.
func (s *Service) NeedsElevation() bool {
	return NeedsElevation()
}

// Rollback restores the previous version.
func (s *Service) Rollback() error {
	return s.installer.Rollback()
}

// CleanupAfterUpdate cleans up backup files after a successful update.
func (s *Service) CleanupAfterUpdate() error {
	return s.installer.CleanupBackup()
}
